package agent

import (
	"bytes"
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/leg100/otf"
	"github.com/leg100/otf/environment"
	"github.com/pkg/errors"
)

const (
	localStateFilename = "terraform.tfstate"
	planFilename       = "plan.out"
	jsonPlanFilename   = "plan.out.json"
	lockFilename       = ".terraform.lock.hcl"
)

type Job struct {
	otf.Run
	environment.Environment

	workspace otf.Workspace
}

func (r *Job) Do() error {
	if err := r.setupEnv(); err != nil {
		return err
	}
	switch r.Status {
	case otf.RunPlanning:
		return r.doPlan()
	case otf.RunApplying:
		return r.doApply()
	default:
		return fmt.Errorf("invalid status: %s", r.Status)
	}
}

// setupEnv invokes the necessary steps before a plan or apply can proceed.
func (r *Job) setupEnv() error {
	if err := r.RunFunc(r.downloadTerraform); err != nil {
		return err
	}
	if err := r.RunFunc(r.downloadConfig); err != nil {
		return err
	}
	if err := r.RunFunc(r.deleteBackendConfig); err != nil {
		return err
	}
	if err := r.RunFunc(r.downloadState); err != nil {
		return err
	}
	if r.Status == otf.RunApplying {
		// Download lock file from plan phase for the apply phase, to ensure
		// same providers are used in both phases.
		if err := r.RunFunc(r.downloadLockFile); err != nil {
			return err
		}
	}
	if err := r.RunTerraform("init"); err != nil {
		return fmt.Errorf("running terraform init: %w", err)
	}
	return nil
}

func (r *Job) doPlan() error {
	if err := r.doTerraformPlan(); err != nil {
		return err
	}
	if err := r.RunCLI("sh", "-c", fmt.Sprintf("%s show -json %s > %s", r.TerraformPath(), planFilename, jsonPlanFilename)); err != nil {
		return err
	}
	if err := r.RunFunc(r.uploadPlan); err != nil {
		return err
	}
	if err := r.RunFunc(r.uploadJSONPlan); err != nil {
		return err
	}
	// upload lock file for use in the apply phase - see note in setupEnv.
	if err := r.RunFunc(r.uploadLockFile); err != nil {
		return err
	}
	return nil
}

func (r *Job) doApply() error {
	if err := r.RunFunc(r.downloadPlanFile); err != nil {
		return err
	}
	if err := r.doTerraformApply(); err != nil {
		return err
	}
	if err := r.RunFunc(r.uploadState); err != nil {
		return err
	}
	return nil
}

func (r *Job) deleteBackendConfig(ctx context.Context) error {
	if err := otf.RewriteHCL(r.WorkingDir(), otf.RemoveBackendBlock); err != nil {
		return fmt.Errorf("removing backend config: %w", err)
	}
	return nil
}

func (r *Job) downloadTerraform(ctx context.Context) error {
	_, err := r.Download(ctx, r.workspace.TerraformVersion, r)
	if err != nil {
		return err
	}
	return nil
}

func (r *Job) downloadConfig(ctx context.Context) error {
	// Download config
	cv, err := r.DownloadConfig(ctx, r.ConfigurationVersionID)
	if err != nil {
		return fmt.Errorf("unable to download config: %w", err)
	}
	// Decompress and untar config into environment root
	if err := otf.Unpack(bytes.NewBuffer(cv), r.Path()); err != nil {
		return fmt.Errorf("unable to unpack config: %w", err)
	}
	return nil
}

// downloadState downloads current state to disk. If there is no state yet
// nothing will be downloaded and no error will be reported.
func (r *Job) downloadState(ctx context.Context) error {
	statefile, err := r.DownloadCurrentState(ctx, r.WorkspaceID)
	if errors.Is(err, otf.ErrResourceNotFound) {
		return nil
	} else if err != nil {
		return fmt.Errorf("downloading state version: %w", err)
	}
	if err := os.WriteFile(filepath.Join(r.WorkingDir(), localStateFilename), statefile, 0o644); err != nil {
		return fmt.Errorf("saving state to local disk: %w", err)
	}
	return nil
}

func (r *Job) uploadPlan(ctx context.Context) error {
	file, err := os.ReadFile(filepath.Join(r.WorkingDir(), planFilename))
	if err != nil {
		return err
	}

	if err := r.UploadPlanFile(ctx, r.ID, file, otf.PlanFormatBinary); err != nil {
		return fmt.Errorf("unable to upload plan: %w", err)
	}

	return nil
}

func (r *Job) uploadJSONPlan(ctx context.Context) error {
	jsonFile, err := os.ReadFile(filepath.Join(r.WorkingDir(), jsonPlanFilename))
	if err != nil {
		return err
	}
	if err := r.UploadPlanFile(ctx, r.ID, jsonFile, otf.PlanFormatJSON); err != nil {
		return fmt.Errorf("unable to upload JSON plan: %w", err)
	}
	return nil
}

// downloadLockFile downloads the .terraform.lock.hcl file into the working
// directory. If one has not been uploaded then this will simply write an empty
// file, which is harmless.
func (r *Job) downloadLockFile(ctx context.Context) error {
	lockFile, err := r.GetLockFile(ctx, r.ID)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(r.WorkingDir(), lockFilename), lockFile, 0o644)
}

func (r *Job) uploadLockFile(ctx context.Context) error {
	lockFile, err := os.ReadFile(filepath.Join(r.WorkingDir(), lockFilename))
	if errors.Is(err, fs.ErrNotExist) {
		// there is no lock file to upload, which is ok
		return nil
	} else if err != nil {
		return errors.Wrap(err, "reading lock file")
	}
	if err := r.UploadLockFile(ctx, r.ID, lockFile); err != nil {
		return fmt.Errorf("unable to upload lock file: %w", err)
	}
	return nil
}

func (r *Job) downloadPlanFile(ctx context.Context) error {
	plan, err := r.GetPlanFile(ctx, r.ID, otf.PlanFormatBinary)
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(r.WorkingDir(), planFilename), plan, 0o644)
}

// uploadState reads, parses, and uploads terraform state
func (r *Job) uploadState(ctx context.Context) error {
	state, err := os.ReadFile(filepath.Join(r.WorkingDir(), localStateFilename))
	if err != nil {
		return err
	}
	err = r.CreateStateVersion(ctx, otf.CreateStateVersionOptions{
		WorkspaceID: &r.WorkspaceID,
		State:       state,
	})
	return err
}

// doTerraformPlan invokes terraform plan
func (r *Job) doTerraformPlan() error {
	var args []string
	if r.IsDestroy {
		args = append(args, "-destroy")
	}
	args = append(args, "-out="+planFilename)
	return r.RunTerraform("plan", args...)
}

// doTerraformApply invokes terraform apply
func (r *Job) doTerraformApply() error {
	var args []string
	if r.IsDestroy {
		args = append(args, "-destroy")
	}
	args = append(args, planFilename)
	return r.RunTerraform("apply", args...)
}
