package otf

import (
	"context"
)

// RunFactory is a factory for constructing Run objects.
type RunFactory struct {
	ConfigurationVersionService ConfigurationVersionService
	WorkspaceService            WorkspaceService
}

// New constructs a new run at the beginning of its lifecycle using the provided
// options.
func (f *RunFactory) New(opts RunCreateOptions) (*Run, error) {
	ws, err := f.WorkspaceService.Get(context.Background(), WorkspaceSpec{ID: String(opts.WorkspaceID)})
	if err != nil {
		return nil, err
	}

	cv, err := f.getConfigurationVersion(opts)
	if err != nil {
		return nil, err
	}

	run := NewRunFromDefaults(cv, ws)
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

	return run, nil
}

func (f *RunFactory) getConfigurationVersion(opts RunCreateOptions) (*ConfigurationVersion, error) {
	if opts.ConfigurationVersionID == nil {
		// CV ID not provided, get workspace's latest CV
		return f.ConfigurationVersionService.GetLatest(opts.WorkspaceID)
	}
	return f.ConfigurationVersionService.Get(*opts.ConfigurationVersionID)
}

// NewRunFromDefaults creates a new run with defaults.
func NewRunFromDefaults(cv *ConfigurationVersion, ws *Workspace) *Run {
	run := Run{
		id:        NewID("run"),
		createdAt: CurrentTimestamp(),
		refresh:   DefaultRefresh,
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
		r.Workspace = &Workspace{id: id}
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
		id:                   id,
		refresh:              DefaultRefresh,
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
