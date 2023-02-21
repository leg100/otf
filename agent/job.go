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
	"github.com/leg100/otf/run"
	"github.com/pkg/errors"
)

type Job struct {
	otf.Run
	environment.Environment
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
	if err := r.RunCLI("sh", "-c", fmt.Sprintf("%s show -json %s > %s", r.TerraformPath(), PlanFilename, JSONPlanFilename)); err != nil {
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
	ws, err := r.GetWorkspace(ctx, r.WorkspaceID)
	if err != nil {
		return err
	}
	_, err = r.Download(ctx, ws.TerraformVersion(), r)
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
	if err := os.WriteFile(filepath.Join(r.WorkingDir(), run.LocalStateFilename), statefile, 0o644); err != nil {
		return fmt.Errorf("saving state to local disk: %w", err)
	}
	return nil
}

func (r *Job) uploadPlan(ctx context.Context) error {
	file, err := os.ReadFile(filepath.Join(r.WorkingDir(), run.PlanFilename))
	if err != nil {
		return err
	}

	if err := r.UploadPlanFile(ctx, r.ID, file, otf.PlanFormatBinary); err != nil {
		return fmt.Errorf("unable to upload plan: %w", err)
	}

	return nil
}

func (r *Job) uploadJSONPlan(ctx context.Context, env environment.Environment) error {
	jsonFile, err := os.ReadFile(filepath.Join(env.WorkingDir(), run.JSONPlanFilename))
	if err != nil {
		return err
	}
	if err := env.UploadPlanFile(ctx, r.ID, jsonFile, otf.PlanFormatJSON); err != nil {
		return fmt.Errorf("unable to upload JSON plan: %w", err)
	}
	return nil
}

// downloadLockFile downloads the .terraform.lock.hcl file into the working
// directory. If one has not been uploaded then this will simply write an empty
// file, which is harmless.
func (r *Job) downloadLockFile(ctx context.Context, env environment.Environment) error {
	lockFile, err := env.GetLockFile(ctx, r.ID)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(env.WorkingDir(), run.LockFilename), lockFile, 0o644)
}

func (r *Job) uploadLockFile(ctx context.Context, env environment.Environment) error {
	lockFile, err := os.ReadFile(filepath.Join(env.WorkingDir(), run.LockFilename))
	if errors.Is(err, fs.ErrNotExist) {
		// there is no lock file to upload, which is ok
		return nil
	} else if err != nil {
		return errors.Wrap(err, "reading lock file")
	}
	if err := env.UploadLockFile(ctx, r.ID, lockFile); err != nil {
		return fmt.Errorf("unable to upload lock file: %w", err)
	}
	return nil
}

func (r *Job) downloadPlanFile(ctx context.Context, env environment.Environment) error {
	plan, err := env.GetPlanFile(ctx, r.ID, otf.PlanFormatBinary)
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(env.WorkingDir(), run.PlanFilename), plan, 0o644)
}

// uploadState reads, parses, and uploads terraform state
func (r *Job) uploadState(ctx context.Context, env environment.Environment) error {
	state, err := os.ReadFile(filepath.Join(env.WorkingDir(), run.LocalStateFilename))
	if err != nil {
		return err
	}
	err = env.CreateStateVersion(ctx, otf.CreateStateVersionOptions{
		WorkspaceID: &r.WorkspaceID,
		State:       state,
	})
	return err
}

// doTerraformPlan invokes terraform plan
func (r *Job) doTerraformPlan(env environment.Environment) error {
	var args []string
	if r.IsDestroy {
		args = append(args, "-destroy")
	}
	args = append(args, "-out="+run.PlanFilename)
	return env.RunTerraform("plan", args...)
}

// doTerraformApply invokes terraform apply
func (r *Job) doTerraformApply() error {
	var args []string
	if r.IsDestroy {
		args = append(args, "-destroy")
	}
	args = append(args, run.PlanFilename)
	return r.RunTerraform("apply", args...)
}
