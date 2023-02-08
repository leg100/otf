package run

import (
	"context"

	"github.com/leg100/otf"
)

// factory constructs runs
type factory struct {
	otf.ConfigurationVersionService
	otf.WorkspaceService
}

// NewRun constructs a new run at the beginning of its lifecycle using the
// provided options.
func (f *factory) NewRun(ctx context.Context, workspaceID string, opts RunCreateOptions) (*Run, error) {
	ws, err := f.WorkspaceService.GetWorkspace(ctx, workspaceID)
	if err != nil {
		return nil, err
	}

	var cv otf.ConfigurationVersion
	if opts.ConfigurationVersionID != nil {
		cv, err = f.GetConfigurationVersion(ctx, *opts.ConfigurationVersionID)
	} else {
		cv, err = f.GetLatestConfigurationVersion(ctx, workspaceID)
	}
	if err != nil {
		return nil, err
	}

	return NewRun(cv, ws, opts), nil
}

// NewRun creates a new run with defaults.
func NewRun(cv otf.ConfigurationVersion, ws otf.Workspace, opts RunCreateOptions) *Run {
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
