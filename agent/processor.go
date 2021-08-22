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
	LocalStateFilename = "terraform.tfstate"
	PlanFilename       = "plan.out"
	JSONPlanFilename   = "plan.out.json"
)

type Processor interface {
	Plan(ctx context.Context, run *ots.Run, path string) error
	Apply(ctx context.Context, run *ots.Run, path string) error
}

type processor struct {
	Run *ots.Run

	logr.Logger

	TerraformRunner TerraformRunner

	ConfigurationVersionService ots.ConfigurationVersionService
	StateVersionService         ots.StateVersionService
	PlanService                 ots.PlanService
	RunService                  ots.RunService
}

// Plan processes a run plan
func (p *processor) Plan(ctx context.Context, run *ots.Run, path string) error {
	var steps []ots.Step

	steps = append(steps, ots.NewFuncStep(func(ctx context.Context, path string) error {
		return p.downloadConfig(ctx, run, path)
	}))

	steps = append(steps, ots.NewFuncStep(func(ctx context.Context, path string) error {
		return deleteBackendConfigFromDirectory(ctx, path)
	}))

	steps = append(steps, ots.NewCommandStep("terraform", "init", "-no-color"))
	steps = append(steps, ots.NewCommandStep("terraform", "plan", "-no-color", fmt.Sprintf("-out=%s", PlanFilename)))
	steps = append(steps, ots.NewCommandStep("sh", "-c",
		fmt.Sprintf("terraform show -json %s > %s", PlanFilename, JSONPlanFilename)))

	steps = append(steps, ots.NewFuncStep(func(ctx context.Context, path string) error {
		file, err := os.ReadFile(PlanFilename)
		if err != nil {
			return err
		}
		jsonFile, err := os.ReadFile(JSONPlanFilename)
		if err != nil {
			return err
		}

		// Parse plan output
		info, err := parsePlanOutput(string(out))
		if err != nil {
			return fmt.Errorf("unable to parse plan output: %w", err)
		}

		// Update status
		_, err = p.RunService.FinishPlan(run.ID, ots.PlanFinishOptions{
			ResourceAdditions:    info.adds,
			ResourceChanges:      info.changes,
			ResourceDestructions: info.deletions,
			Plan:                 file,
			PlanJSON:             jsonFile,
		})
		if err != nil {
			return fmt.Errorf("unable to finish plan: %w", err)
		}
	}))

	// Update status
	_, err := p.RunService.UpdatePlanStatus(run.ID, tfe.PlanRunning)
	if err != nil {
		return fmt.Errorf("unable to update plan status: %w", err)
	}

	// Upload logs regardless of whether plan errored
	if err := p.RunService.UploadPlanLogs(run.ID, out); err != nil {
		return fmt.Errorf("unable to upload plan logs: %w", err)
	}

	// Go no further if plan errored
	if planErr != nil {
		return planErr
	}

	planFile := ots.PlanFile{}
	if err := json.Unmarshal(planJSON, &planFile); err != nil {
		return err
	}

	// Parse plan output
	adds, updates, deletes := planFile.Changes()

	// Update status
	_, err = p.RunService.FinishPlan(run.ID, ots.PlanFinishOptions{
		ResourceAdditions:    adds,
		ResourceChanges:      updates,
		ResourceDestructions: deletes,
		Plan:                 plan,
		PlanJSON:             planJSON,
	})
	if err != nil {
		return fmt.Errorf("unable to finish plan: %w", err)
	}

	p.Info("job completed", "run", run.ID,
		"additions", adds,
		"changes", updates,
		"deletions", deletes,
	)

	return nil
}

// Apply processes a run apply
func (p *processor) Apply(ctx context.Context, run *ots.Run, path string) error {
	// Download config
	if err := p.downloadConfig(ctx, run, path); err != nil {
		return fmt.Errorf("unable to download config: %w", err)
	}

	// Delete backend config
	if err := deleteBackendConfigFromDirectory(ctx, path); err != nil {
		return fmt.Errorf("unable to delete backend config: %w", err)
	}

	// Download state
	if err := p.downloadState(ctx, run, path); err != nil {
		return fmt.Errorf("unable to download state: %w", err)
	}

	// Download plan file
	if err := p.downloadPlanFile(ctx, run, path); err != nil {
		return fmt.Errorf("unable to download plan file: %w", err)
	}

	// Update status
	_, err := p.RunService.UpdateApplyStatus(run.ID, tfe.ApplyRunning)
	if err != nil {
		return fmt.Errorf("unable to update apply status: %w", err)
	}

	// Run terraform init then apply
	out, applyErr := p.TerraformRunner.Apply(ctx, path)

	// Upload logs regardless of whether apply failed
	if err := p.RunService.UploadApplyLogs(run.ID, out); err != nil {
		return fmt.Errorf("unable to upload apply logs: %w", err)
	}

	// Go no further if apply errored
	if applyErr != nil {
		return applyErr
	}

	// Parse apply output
	info, err := parseApplyOutput(string(out))
	if err != nil {
		return fmt.Errorf("unable to parse apply output: %w", err)
	}

	// Upload state if there are changes
	if err := p.uploadState(ctx, run, path); err != nil {
		return err
	}

	// Update status
	_, err = p.RunService.FinishApply(run.ID, ots.ApplyFinishOptions{
		ResourceAdditions:    info.adds,
		ResourceChanges:      info.changes,
		ResourceDestructions: info.deletions,
	})
	if err != nil {
		return fmt.Errorf("unable to finish apply: %w", err)
	}

	p.Info("job completed", "run", run.ID,
		"additions", info.adds,
		"changes", info.changes,
		"deletions", info.deletions)

	return nil
}

func (p *processor) downloadConfig(ctx context.Context, run *ots.Run, path string) error {
	// Download config
	cv, err := p.ConfigurationVersionService.Download(run.ConfigurationVersion.ID)
	if err != nil {
		return fmt.Errorf("unable to download config: %w", err)
	}

	// Decompress and untar config
	if err := Unpack(bytes.NewBuffer(cv), path); err != nil {
		return fmt.Errorf("unable to unpack config: %w", err)
	}

	return nil
}

// Download current state to disk. If there is no state yet nothing will be
// downloaded and no error will be reported.
func (p *processor) downloadState(ctx context.Context, run *ots.Run, path string) error {
	state, err := p.StateVersionService.Current(run.Workspace.ID)
	if ots.IsNotFound(err) {
		return nil
	} else if err != nil {
		return err
	}

	statefile, err := p.StateVersionService.Download(state.ID)
	if err != nil {
		return err
	}

	if err := os.WriteFile(filepath.Join(path, LocalStateFilename), statefile, 0644); err != nil {
		return err
	}

	return nil
}

// downloadPlanFile downloads a plan file, for use with terraform apply.
func (p *processor) downloadPlanFile(ctx context.Context, run *ots.Run, path string) error {
	plan, err := p.RunService.GetPlanFile(run.ID)
	if err != nil {
		return err
	}

	if err := os.WriteFile(filepath.Join(path, PlanFilename), plan, 0644); err != nil {
		return err
	}

	return nil
}

// Read, parse, and upload state
func (p *processor) uploadState(ctx context.Context, run *ots.Run, path string) error {
	stateFile, err := os.ReadFile(filepath.Join(path, LocalStateFilename))
	if err != nil {
		return err
	}

	state, err := ots.Parse(stateFile)
	if err != nil {
		return err
	}

	_, err = p.StateVersionService.Create(run.Workspace.ID, tfe.StateVersionCreateOptions{
		State:   ots.String(base64.StdEncoding.EncodeToString(stateFile)),
		MD5:     ots.String(fmt.Sprintf("%x", md5.Sum(stateFile))),
		Lineage: &state.Lineage,
		Serial:  ots.Int64(state.Serial),
	})
	if err != nil {
		return err
	}

	return nil
}
