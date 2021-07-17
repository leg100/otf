package agent

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"

	"github.com/leg100/go-tfe"
	"github.com/leg100/ots"
	"github.com/rs/zerolog"
)

const (
	LocalStateFilename = "terraform.tfstate"
)

type Processor interface {
	Process(ctx context.Context, run *ots.Run, path string) error
}

type processor struct {
	Run *ots.Run

	zerolog.Logger

	TerraformRunner TerraformRunner

	Path string

	ConfigurationVersionService ots.ConfigurationVersionService
	StateVersionService         ots.StateVersionService
	PlanService                 ots.PlanService
	RunService                  ots.RunService
}

// Process processes a run
func (p *processor) Process(ctx context.Context, run *ots.Run, path string) error {
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

	// Update status
	_, err := p.RunService.UpdatePlanStatus(run.ExternalID, tfe.PlanRunning)
	if err != nil {
		return fmt.Errorf("unable to update plan status: %w", err)
	}

	// Run terraform init then plan
	out, err := p.TerraformRunner.Plan(ctx, path)
	if err != nil {
		return err
	}

	// Upload logs
	if err := p.RunService.UploadPlanLogs(run.ExternalID, out); err != nil {
		return fmt.Errorf("unable to upload plan logs: %w", err)
	}

	// Upload state
	//if err := p.uploadState(ctx, run, path); err != nil {
	//	return err
	//}

	// Parse plan output
	plan, err := parsePlanOutput(string(out))
	if err != nil {
		return fmt.Errorf("unable to parse plan output: %w", err)
	}

	// Update status
	_, err = p.RunService.FinishPlan(run.ExternalID, ots.PlanFinishOptions{
		ResourceAdditions:    plan.adds,
		ResourceChanges:      plan.changes,
		ResourceDestructions: plan.deletions,
	})
	if err != nil {
		return fmt.Errorf("unable to finish plan: %w", err)
	}

	p.Info().
		Str("run", run.ExternalID).
		Int("additions", plan.adds).
		Int("changes", plan.changes).
		Int("deletions", plan.deletions).
		Msg("job completed")

	return nil
}

func (p *processor) downloadConfig(ctx context.Context, run *ots.Run, path string) error {
	// Download config
	cv, err := p.ConfigurationVersionService.Download(run.ConfigurationVersion.ExternalID)
	if err != nil {
		return err
	}

	// Decompress and untar config
	if err := Unpack(bytes.NewBuffer(cv), path); err != nil {
		return err
	}

	return nil
}

// Download current state to disk. If there is no state yet nothing will be
// downloaded and no error will be reported.
func (p *processor) downloadState(ctx context.Context, run *ots.Run, path string) error {
	state, err := p.StateVersionService.Current(run.Workspace.ExternalID)
	if ots.IsNotFound(err) {
		return nil
	} else if err != nil {
		return err
	}

	statefile, err := p.StateVersionService.Download(state.ExternalID)
	if err != nil {
		return err
	}

	if err := os.WriteFile(filepath.Join(path, LocalStateFilename), statefile, 0644); err != nil {
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

	_, err = p.StateVersionService.Create(run.Workspace.ExternalID, tfe.StateVersionCreateOptions{
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
