package otf

import (
	"context"
	"errors"
)

// RunFactory is a factory for constructing Run objects.
type RunFactory struct {
	ConfigurationVersionService ConfigurationVersionService
	WorkspaceService            WorkspaceService
}

// New constructs a new run at the beginning of its lifecycle using the provided
// options.
func (f *RunFactory) New(opts RunCreateOptions) (*Run, error) {
	if opts.Workspace == nil {
		return nil, errors.New("workspace is required")
	}

	ws, err := f.WorkspaceService.Get(context.Background(), WorkspaceSpec{ID: &opts.Workspace.ID})
	if err != nil {
		return nil, err
	}

	cv, err := f.getConfigurationVersion(opts)
	if err != nil {
		return nil, err
	}

	run := NewRunFromDefaults(cv, ws)
	run.ReplaceAddrs = opts.ReplaceAddrs
	run.TargetAddrs = opts.TargetAddrs
	if opts.IsDestroy != nil {
		run.IsDestroy = *opts.IsDestroy
	}
	if opts.Message != nil {
		run.Message = *opts.Message
	}
	if opts.Refresh != nil {
		run.Refresh = *opts.Refresh
	}

	return run, nil
}

func (f *RunFactory) getConfigurationVersion(opts RunCreateOptions) (*ConfigurationVersion, error) {
	if opts.ConfigurationVersion == nil {
		// CV ID not provided, get workspace's latest CV
		return f.ConfigurationVersionService.GetLatest(opts.Workspace.ID)
	}
	return f.ConfigurationVersionService.Get(opts.ConfigurationVersion.ID)
}

// NewRunFromDefaults creates a new run with defaults.
func NewRunFromDefaults(cv *ConfigurationVersion, ws *Workspace) *Run {
	run := Run{
		ID:      NewID("run"),
		Refresh: DefaultRefresh,
		status:  RunPending,
	}
	run.ConfigurationVersion = &ConfigurationVersion{ID: cv.ID}
	run.Workspace = &Workspace{ID: ws.ID}
	run.Plan = newPlan(&run)
	run.Apply = newApply(&run)
	run.autoApply = ws.AutoApply
	run.speculative = cv.Speculative
	if run.IsSpeculative() {
		// immediately enqueue plans for speculative runs
		run.UpdateStatus(RunPlanQueued)
	}
	return &run
}

type TestRunOption func(*Run)

func TestRunStatus(status RunStatus) TestRunOption {
	return func(r *Run) {
		r.status = status
	}
}

func TestRunWorkspaceID(id string) TestRunOption {
	return func(r *Run) {
		r.Workspace = &Workspace{ID: id}
	}
}

func TestRunSpeculative() TestRunOption {
	return func(r *Run) {
		r.speculative = true
	}
}

// NewTestRun creates a new run expressly for testing purposes
func NewTestRun(id string, opts ...TestRunOption) *Run {
	run := Run{
		ID:                   id,
		Refresh:              DefaultRefresh,
		status:               RunPending,
		ConfigurationVersion: &ConfigurationVersion{},
	}
	for _, o := range opts {
		o(&run)
	}
	run.Plan = newPlan(&run)
	run.Apply = newApply(&run)
	run.setJob()
	return &run
}
