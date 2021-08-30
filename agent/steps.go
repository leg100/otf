package agent

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-logr/logr"
	"github.com/leg100/go-tfe"
	"github.com/leg100/ots"
)

const (
	LocalStateFilename  = "terraform.tfstate"
	PlanFilename        = "plan.out"
	JSONPlanFilename    = "plan.out.json"
	ApplyOutputFilename = "apply.out"
)

var (
	DeleteBackendStep = ots.NewFuncStep(deleteBackendConfigFromDirectory)
	InitStep          = ots.NewCommandStep("terraform", "init", "-no-color")
	PlanStep          = ots.NewCommandStep("terraform", "plan", "-no-color", fmt.Sprintf("-out=%s", PlanFilename))
	JSONPlanStep      = ots.NewCommandStep("sh", "-c", fmt.Sprintf("terraform show -json %s > %s", PlanFilename, JSONPlanFilename))
	ApplyStep         = ots.NewCommandStep("sh", "-c", fmt.Sprintf("terraform apply -no-color %s | tee %s", PlanFilename, ApplyOutputFilename))
)

func DownloadConfigStep(run *ots.Run, cvs ots.ConfigurationVersionService) *ots.FuncStep {
	return ots.NewFuncStep(func(ctx context.Context, path string) error {
		// Download config
		cv, err := cvs.Download(run.ConfigurationVersion.ID)
		if err != nil {
			return fmt.Errorf("unable to download config: %w", err)
		}

		// Decompress and untar config
		if err := Unpack(bytes.NewBuffer(cv), path); err != nil {
			return fmt.Errorf("unable to unpack config: %w", err)
		}

		return nil
	})
}

func UpdatePlanStatusStep(run *ots.Run, rs ots.RunService) *ots.FuncStep {
	return ots.NewFuncStep(func(ctx context.Context, path string) error {
		_, err := rs.UpdateStatus(run.ID, tfe.RunPlanning)
		return err
	})
}

func UpdateApplyStatusStep(run *ots.Run, rs ots.RunService) *ots.FuncStep {
	return ots.NewFuncStep(func(ctx context.Context, path string) error {
		_, err := rs.UpdateStatus(run.ID, tfe.RunApplying)
		return err
	})
}

func UploadPlanStep(run *ots.Run, rs ots.RunService) *ots.FuncStep {
	return ots.NewFuncStep(func(ctx context.Context, path string) error {
		file, err := os.ReadFile(filepath.Join(path, PlanFilename))
		if err != nil {
			return err
		}

		// Upload plan
		if err := rs.UploadPlan(run.ID, file, false); err != nil {
			return fmt.Errorf("unable to upload plan: %w", err)
		}

		return nil
	})
}

func UploadJSONPlanStep(run *ots.Run, rs ots.RunService) *ots.FuncStep {
	return ots.NewFuncStep(func(ctx context.Context, path string) error {
		jsonFile, err := os.ReadFile(filepath.Join(path, JSONPlanFilename))
		if err != nil {
			return err
		}

		// Upload plan
		if err := rs.UploadPlan(run.ID, jsonFile, true); err != nil {
			return fmt.Errorf("unable to upload JSON plan: %w", err)
		}

		return nil
	})
}

func FinishPlanStep(run *ots.Run, rs ots.RunService, logger logr.Logger) *ots.FuncStep {
	return ots.NewFuncStep(func(ctx context.Context, path string) error {
		jsonFile, err := os.ReadFile(filepath.Join(path, JSONPlanFilename))
		if err != nil {
			return err
		}

		planFile := ots.PlanFile{}
		if err := json.Unmarshal(jsonFile, &planFile); err != nil {
			return err
		}

		// Parse plan output
		adds, updates, deletes := planFile.Changes()

		// Update status
		_, err = rs.FinishPlan(run.ID, ots.PlanFinishOptions{
			ResourceAdditions:    adds,
			ResourceChanges:      updates,
			ResourceDestructions: deletes,
		})
		if err != nil {
			return fmt.Errorf("unable to finish plan: %w", err)
		}

		logger.Info("job completed", "run", run.ID,
			"additions", adds,
			"changes", updates,
			"deletions", deletes,
		)

		return nil
	})
}

func FinishApplyStep(run *ots.Run, rs ots.RunService, logger logr.Logger) *ots.FuncStep {
	return ots.NewFuncStep(func(ctx context.Context, path string) error {
		out, err := os.ReadFile(filepath.Join(path, ApplyOutputFilename))
		if err != nil {
			return err
		}

		// Parse apply output
		info, err := parseApplyOutput(string(out))
		if err != nil {
			return fmt.Errorf("unable to parse apply output: %w", err)
		}

		// Update status
		_, err = rs.FinishApply(run.ID, ots.ApplyFinishOptions{
			ResourceAdditions:    info.adds,
			ResourceChanges:      info.changes,
			ResourceDestructions: info.deletions,
		})
		if err != nil {
			return fmt.Errorf("unable to finish apply: %w", err)
		}

		logger.Info("job completed", "run", run.ID,
			"additions", info.adds,
			"changes", info.changes,
			"deletions", info.deletions)

		return nil
	})
}

// DownloadStateStep downloads current state to disk. If there is no state yet
// nothing will be downloaded and no error will be reported.
func DownloadStateStep(run *ots.Run, svs ots.StateVersionService, logger logr.Logger) *ots.FuncStep {
	return ots.NewFuncStep(func(ctx context.Context, path string) error {
		state, err := svs.Current(run.Workspace.ID)
		if ots.IsNotFound(err) {
			return nil
		} else if err != nil {
			return err
		}

		statefile, err := svs.Download(state.ID)
		if err != nil {
			return err
		}

		if err := os.WriteFile(filepath.Join(path, LocalStateFilename), statefile, 0644); err != nil {
			return err
		}

		return nil
	})
}

func DownloadPlanFileStep(run *ots.Run, rs ots.RunService) *ots.FuncStep {
	return ots.NewFuncStep(func(ctx context.Context, path string) error {
		plan, err := rs.GetPlanFile(run.ID)
		if err != nil {
			return err
		}

		return os.WriteFile(filepath.Join(path, PlanFilename), plan, 0644)
	})
}

// UploadStateStep reads, parses, and uploads state
func UploadStateStep(run *ots.Run, svs ots.StateVersionService) *ots.FuncStep {
	return ots.NewFuncStep(func(ctx context.Context, path string) error {
		stateFile, err := os.ReadFile(filepath.Join(path, LocalStateFilename))
		if err != nil {
			return err
		}

		state, err := ots.Parse(stateFile)
		if err != nil {
			return err
		}

		_, err = svs.Create(run.Workspace.ID, tfe.StateVersionCreateOptions{
			State:   ots.String(base64.StdEncoding.EncodeToString(stateFile)),
			MD5:     ots.String(fmt.Sprintf("%x", md5.Sum(stateFile))),
			Lineage: &state.Lineage,
			Serial:  ots.Int64(state.Serial),
		})
		if err != nil {
			return err
		}

		return nil
	})
}
