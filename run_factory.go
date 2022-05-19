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
		Status:  RunPending,
	}
	run.ConfigurationVersion = cv
	run.Workspace = ws
	run.Plan = newPlan(&run)
	run.Apply = newApply(&run)
	run.setJob()
	return &run
}
