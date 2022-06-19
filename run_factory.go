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
	run.configurationVersionID = cv.ID()
	run.workspaceID = ws.ID()
	run.Plan = newPlan(&run)
	run.Apply = newApply(&run)
	run.autoApply = ws.AutoApply()
	run.speculative = cv.Speculative()

	setupRunStates(&run)
	run.setState(run.pendingState)

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

func setupRunStates(run *Run) {
	run.pendingState = newPendingState(run)
	run.planQueuedState = newPlanQueuedState(run)
	run.planningState = newPlanningState(run)
	run.plannedState = newPlannedState(run)
	run.plannedAndFinishedState = newPlannedAndFinishedState(run)
	run.applyQueuedState = newApplyQueuedState(run)
	run.applyingState = newApplyingState(run)
	run.appliedState = newAppliedState(run)
	run.discardedState = newDiscardedState(run)
	run.erroredState = newErroredState(run)
	run.canceledState = newCanceledState(run)
}

// NewTestRun creates a new run. Expressly for testing purposes
func NewTestRun(t *testing.T, id, workspaceID string, opts TestRunCreateOptions) *Run {
	ws := Workspace{id: workspaceID}
	cv := ConfigurationVersion{id: "cv-123", speculative: opts.Speculative}
	run := NewRun(&cv, &ws, RunCreateOptions{})
	return run
}
