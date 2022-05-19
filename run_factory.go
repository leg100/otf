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

// NewRun constructs a run object.
func (f *RunFactory) NewRun(opts RunCreateOptions) (*Run, error) {
	if opts.Workspace == nil {
		return nil, errors.New("workspace is required")
	}

	id := NewID("run")
	run := Run{
		ID:           id,
		Refresh:      DefaultRefresh,
		ReplaceAddrs: opts.ReplaceAddrs,
		TargetAddrs:  opts.TargetAddrs,
		// run always starts in pending state
		Status: RunPending,
	}
	run.Plan = newPlan(&run)
	run.Apply = newApply(&run)

	ws, err := f.WorkspaceService.Get(context.Background(), WorkspaceSpec{ID: &opts.Workspace.ID})
	if err != nil {
		return nil, err
	}
	run.Workspace = ws

	cv, err := f.getConfigurationVersion(opts)
	if err != nil {
		return nil, err
	}
	run.ConfigurationVersion = cv

	if opts.IsDestroy != nil {
		run.IsDestroy = *opts.IsDestroy
	}

	if opts.Message != nil {
		run.Message = *opts.Message
	}

	if opts.Refresh != nil {
		run.Refresh = *opts.Refresh
	}

	return &run, nil
}

func (f *RunFactory) getConfigurationVersion(opts RunCreateOptions) (*ConfigurationVersion, error) {
	// Unless CV ID provided, get workspace's latest CV
	if opts.ConfigurationVersion != nil {
		return f.ConfigurationVersionService.Get(opts.ConfigurationVersion.ID)
	}
	return f.ConfigurationVersionService.GetLatest(opts.Workspace.ID)
}
