package otf

import (
	"context"
	"testing"
)

// RunFactory is a factory for constructing Run objects.
type RunFactory struct {
	ConfigurationVersionService ConfigurationVersionService
	WorkspaceService            WorkspaceService
}

// New constructs a new run at the beginning of its lifecycle using the provided
// options.
func (f *RunFactory) New(ctx context.Context, workspaceSpec WorkspaceSpec, opts RunCreateOptions) (*Run, error) {
	ws, err := f.WorkspaceService.Get(context.Background(), workspaceSpec)
	if err != nil {
		return nil, err
	}
	cv, err := f.getConfigurationVersion(ctx, ws.ID(), opts.ConfigurationVersionID)
	if err != nil {
		return nil, err
	}

	return NewRun(cv, ws, opts), nil
}

func (f *RunFactory) getConfigurationVersion(ctx context.Context, workspaceID string, cvID *string) (*ConfigurationVersion, error) {
	if cvID == nil {
		// CV ID not provided, get workspace's latest CV
		return f.ConfigurationVersionService.GetLatest(ctx, workspaceID)
	}
	return f.ConfigurationVersionService.Get(ctx, *cvID)
}

// NewRun creates a new run with defaults.
func NewRun(cv *ConfigurationVersion, ws *Workspace, opts RunCreateOptions) *Run {
	run := Run{
		id:               NewID("run"),
		createdAt:        CurrentTimestamp(),
		refresh:          DefaultRefresh,
		workspaceName:    ws.Name(),
		organizationName: ws.OrganizationName(),
	}
	run.ConfigurationVersion = &ConfigurationVersion{id: cv.ID()}
	run.Workspace = &Workspace{id: ws.ID()}
	run.Plan = newPlan(&run)
	run.Apply = newApply(&run)
	run.autoApply = ws.AutoApply()
	run.speculative = cv.Speculative()
	run.setJob()
	run.updateStatus(RunPending)
	if run.Speculative() {
		// immediately enqueue plans for speculative runs
		run.updateStatus(RunPlanQueued)
	}
	// apply options
	run.replaceAddrs = opts.ReplaceAddrs
	run.targetAddrs = opts.TargetAddrs
	if opts.IsDestroy != nil {
		run.isDestroy = *opts.IsDestroy
	}
	if opts.Message != nil {
		run.message = *opts.Message
	}
	if opts.Refresh != nil {
		run.refresh = *opts.Refresh
	}
	return &run
}

// NewTestRun creates a new run. Expressly for testing purposes
func NewTestRun(t *testing.T, id, workspaceID string, opts TestRunCreateOptions) *Run {
	ws := Workspace{id: workspaceID}
	cv := ConfigurationVersion{id: "cv-123", speculative: opts.Speculative}
	run := NewRun(&cv, &ws, RunCreateOptions{})
	if opts.Status != RunStatus("") {
		run.updateStatus(opts.Status)
	}
	return run
}
