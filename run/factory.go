package otf

import (
	"context"
)

// RunFactory is a factory for constructing Run objects.
type RunFactory struct {
	ConfigurationVersionService ConfigurationVersionService
	WorkspaceService            WorkspaceService
}

// NewRun constructs a new run at the beginning of its lifecycle using the
// provided options.
func (f *RunFactory) NewRun(ctx context.Context, workspaceID string, opts RunCreateOptions) (*Run, error) {
	ws, err := f.WorkspaceService.GetWorkspace(ctx, workspaceID)
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
		return f.ConfigurationVersionService.GetLatestConfigurationVersion(ctx, workspaceID)
	}
	return f.ConfigurationVersionService.GetConfigurationVersion(ctx, *cvID)
}

// NewRun creates a new run with defaults.
func NewRun(cv *ConfigurationVersion, ws *Workspace, opts RunCreateOptions) *Run {
	run := Run{
		id:                     NewID("run"),
		createdAt:              CurrentTimestamp(),
		refresh:                DefaultRefresh,
		organization:           ws.Organization(),
		configurationVersionID: cv.ID(),
		workspaceID:            ws.ID(),
		speculative:            cv.Speculative(),
		replaceAddrs:           opts.ReplaceAddrs,
		targetAddrs:            opts.TargetAddrs,
		executionMode:          ws.ExecutionMode(),
		autoApply:              ws.AutoApply(),
	}
	run.plan = newPlan(&run)
	run.apply = newApply(&run)
	run.updateStatus(RunPending)

	if opts.IsDestroy != nil {
		run.isDestroy = *opts.IsDestroy
	}
	if opts.Message != nil {
		run.message = *opts.Message
	}
	if opts.Refresh != nil {
		run.refresh = *opts.Refresh
	}
	if opts.AutoApply != nil {
		run.autoApply = *opts.AutoApply
	}
	if cv.ingressAttributes != nil {
		run.commit = &cv.ingressAttributes.CommitSHA
	}
	return &run
}
