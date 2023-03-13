package agent

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"strings"

	"github.com/fatih/color"
	"github.com/leg100/otf"
	"github.com/leg100/otf/run"
	"github.com/leg100/otf/variable"
)

const (
	localStateFilename = "terraform.tfstate"
	planFilename       = "plan.out"
	jsonPlanFilename   = "plan.out.json"
	lockFilename       = ".terraform.lock.hcl"
)

type (
	step func(context.Context) error

	stepsBuilder struct {
		*run.Run
		*Environment
	}

	runner struct {
		cancelFunc context.CancelFunc // Cancel context func for currently running func
		canceled   bool               // Whether cancelation has been requested
		out        io.WriteCloser     // for writing out error message to user
	}
)

func buildSteps(env *Environment, run *run.Run) (steps []step) {
	bldr := &stepsBuilder{Environment: env, Run: run}

	// default setup steps
	steps = append(steps, bldr.downloadTerraform)
	steps = append(steps, bldr.downloadConfig)
	steps = append(steps, bldr.writeTerraformVars)
	steps = append(steps, bldr.deleteBackendConfig)
	steps = append(steps, bldr.downloadState)

	switch run.Phase() {
	case otf.PlanPhase:
		steps = append(steps, bldr.terraformInit)
		steps = append(steps, bldr.terraformPlan)
		steps = append(steps, bldr.convertPlanToJSON)
		steps = append(steps, bldr.uploadPlan)
		steps = append(steps, bldr.uploadJSONPlan)
		steps = append(steps, bldr.uploadLockFile)
	case otf.ApplyPhase:
		// Download lock file from plan phase for the apply phase, to ensure
		// same providers are used in both phases.
		steps = append(steps, bldr.downloadLockFile)
		steps = append(steps, bldr.downloadPlanFile)
		steps = append(steps, bldr.terraformInit)
		steps = append(steps, bldr.terraformApply)
		steps = append(steps, bldr.uploadState)
	}

	return
}

func (r *runner) processSteps(ctx context.Context, steps []step) error {
	ctx, cancel := context.WithCancel(ctx)
	r.cancelFunc = cancel

	for _, s := range steps {
		if r.canceled {
			return fmt.Errorf("execution canceled")
		}
		if err := s(ctx); err != nil {
			// write error message to output
			errbuilder := strings.Builder{}
			errbuilder.WriteRune('\n')

			red := color.New(color.FgHiRed)
			red.EnableColor() // force color on non-tty output
			red.Fprint(&errbuilder, "Error: ")

			errbuilder.WriteString(err.Error())
			errbuilder.WriteRune('\n')
			fmt.Fprint(r.out, errbuilder.String())
			return err
		}
	}
	return nil
}

func (r *runner) cancel(force bool) {
	r.canceled = true

	// cancel func only if forced and there is a context to cancel
	if force && r.cancelFunc != nil {
		r.cancelFunc()
	}
}

func (b *stepsBuilder) downloadTerraform(ctx context.Context) error {
	_, err := b.Download(ctx, b.version, b.out)
	return err
}

func (b *stepsBuilder) downloadConfig(ctx context.Context) error {
	cv, err := b.DownloadConfig(ctx, b.ConfigurationVersionID)
	if err != nil {
		return fmt.Errorf("unable to download config: %w", err)
	}
	// Decompress and untar config into root dir
	if err := otf.Unpack(bytes.NewBuffer(cv), b.workdir.root); err != nil {
		return fmt.Errorf("unable to unpack config: %w", err)
	}
	return nil
}

func (b *stepsBuilder) deleteBackendConfig(ctx context.Context) error {
	if err := otf.RewriteHCL(b.workdir.String(), otf.RemoveBackendBlock); err != nil {
		return fmt.Errorf("removing backend config: %w", err)
	}
	return nil
}

// downloadState downloads current state to disk. If there is no state yet
// nothing will be downloaded and no error will be reported.
func (b *stepsBuilder) downloadState(ctx context.Context) error {
	statefile, err := b.DownloadCurrentState(ctx, b.WorkspaceID)
	if errors.Is(err, otf.ErrResourceNotFound) {
		return nil
	} else if err != nil {
		return fmt.Errorf("downloading state version: %w", err)
	}
	if err := b.WriteFile(localStateFilename, statefile); err != nil {
		return fmt.Errorf("saving state to local disk: %w", err)
	}
	return nil
}

// downloadLockFile downloads the .terraform.lock.hcl file into the working
// directory. If one has not been uploaded then this will simply write an empty
// file, which is harmless.
func (b *stepsBuilder) downloadLockFile(ctx context.Context) error {
	lockFile, err := b.GetLockFile(ctx, b.ID)
	if err != nil {
		return err
	}
	return b.WriteFile(lockFilename, lockFile)
}

func (b *stepsBuilder) writeTerraformVars(ctx context.Context) error {
	if err := variable.WriteTerraformVars(b.workdir.String(), b.variables); err != nil {
		return fmt.Errorf("writing terraform.fvars: %w", err)
	}
	return nil
}

func (b *stepsBuilder) terraformInit(ctx context.Context) error {
	return b.executeTerraform([]string{"init"})
}

func (b *stepsBuilder) terraformPlan(ctx context.Context) error {
	args := []string{"plan"}
	if b.IsDestroy {
		args = append(args, "-destroy")
	}
	args = append(args, "-out="+planFilename)
	return b.executeTerraform(args)
}

func (b *stepsBuilder) terraformApply(ctx context.Context) error {
	args := []string{"apply"}
	if b.IsDestroy {
		args = append(args, "-destroy")
	}
	args = append(args, planFilename)
	return b.executeTerraform(args, sandboxIfEnabled())
}

func (b *stepsBuilder) convertPlanToJSON(ctx context.Context) error {
	args := []string{"show", "-json", planFilename}
	return b.executeTerraform(args, redirectStdout(jsonPlanFilename))
}

func (b *stepsBuilder) uploadPlan(ctx context.Context) error {
	file, err := b.ReadFile(planFilename)
	if err != nil {
		return err
	}

	if err := b.UploadPlanFile(ctx, b.ID, file, run.PlanFormatBinary); err != nil {
		return fmt.Errorf("unable to upload plan: %w", err)
	}

	return nil
}

func (b *stepsBuilder) uploadJSONPlan(ctx context.Context) error {
	jsonFile, err := b.ReadFile(jsonPlanFilename)
	if err != nil {
		return err
	}
	if err := b.UploadPlanFile(ctx, b.ID, jsonFile, run.PlanFormatJSON); err != nil {
		return fmt.Errorf("unable to upload JSON plan: %w", err)
	}
	return nil
}

func (b *stepsBuilder) uploadLockFile(ctx context.Context) error {
	lockFile, err := b.ReadFile(lockFilename)
	if errors.Is(err, fs.ErrNotExist) {
		// there is no lock file to upload, which is ok
		return nil
	} else if err != nil {
		return fmt.Errorf("reading lock file: %w", err)
	}
	if err := b.UploadLockFile(ctx, b.ID, lockFile); err != nil {
		return fmt.Errorf("unable to upload lock file: %w", err)
	}
	return nil
}

func (b *stepsBuilder) downloadPlanFile(ctx context.Context) error {
	plan, err := b.GetPlanFile(ctx, b.ID, run.PlanFormatBinary)
	if err != nil {
		return err
	}

	return b.WriteFile(planFilename, plan)
}

// uploadState reads, parses, and uploads terraform state
func (b *stepsBuilder) uploadState(ctx context.Context) error {
	state, err := b.ReadFile(localStateFilename)
	if err != nil {
		return err
	}
	return b.CreateStateVersion(ctx, otf.CreateStateVersionOptions{
		WorkspaceID: &b.WorkspaceID,
		State:       state,
	})
}
