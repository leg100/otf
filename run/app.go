package run

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/leg100/otf"
	"github.com/leg100/otf/logs"
	"github.com/leg100/otf/rbac"
)

type app interface {
	create(ctx context.Context, workspaceID string, opts RunCreateOptions) (*Run, error)
	get(ctx context.Context, runID string) (*Run, error)
	list(ctx context.Context, opts otf.RunListOptions) (*RunList, error)
	// apply enqueues an apply for the run.
	apply(ctx context.Context, runID string) error
	discard(ctx context.Context, runID string) error
	// cancel a run. If a run is in progress then a cancelation signal will be
	// sent out.
	cancel(ctx context.Context, runID string) error
	// forceCancel forcefully cancels a run.
	forceCancel(ctx context.Context, runID string) error
	// enqueuePlan enqueues a plan for the run.
	//
	// NOTE: this is an internal action, invoked by the scheduler only.
	enqueuePlan(ctx context.Context, runID string) (*Run, error)
	// getPlanFile returns the plan file for the run.
	getPlanFile(ctx context.Context, runID string, format otf.PlanFormat) ([]byte, error)
	// uploadPlanFile persists a run's plan file. The plan format should be either
	// be binary or json.
	uploadPlanFile(ctx context.Context, runID string, plan []byte, format otf.PlanFormat) error
	// delete deletes a run.
	delete(ctx context.Context, runID string) error
	// startPhase starts a run phase.
	startPhase(ctx context.Context, runID string, phase otf.PhaseType, _ otf.PhaseStartOptions) (*Run, error)
	// finishPhase finishes a phase. Creates a report of changes before updating the status of
	// the run.
	finishPhase(ctx context.Context, runID string, phase otf.PhaseType, opts otf.PhaseFinishOptions) (*Run, error)
	// createReport creates a report of changes for the phase.
	createReport(ctx context.Context, runID string, phase otf.PhaseType) (ResourceReport, error)
	createPlanReport(ctx context.Context, runID string) (ResourceReport, error)
	createApplyReport(ctx context.Context, runID string) (ResourceReport, error)

	lockFileApp
}

type Application struct {
	otf.Authorizer
	logr.Logger
	logs.Service
	otf.PubSubService
	otf.WorkspaceService

	cache otf.Cache
	db    *pgdb
	*factory
}

func (a *Application) create(ctx context.Context, workspaceID string, opts RunCreateOptions) (*Run, error) {
	subject, err := a.CanAccessWorkspaceByID(ctx, rbac.CreateRunAction, workspaceID)
	if err != nil {
		return nil, err
	}

	run, err := a.NewRun(ctx, workspaceID, opts)
	if err != nil {
		a.Error(err, "constructing new run", "subject", subject)
		return nil, err
	}

	if err = a.db.CreateRun(ctx, run); err != nil {
		a.Error(err, "creating run", "id", run.ID(), "workspace_id", run.WorkspaceID(), "subject", subject)
		return nil, err
	}
	a.V(1).Info("created run", "id", run.ID(), "workspace_id", run.WorkspaceID(), "subject", subject)

	a.Publish(otf.Event{Type: otf.EventRunCreated, Payload: run})

	return run, nil
}

// GetRun retrieves a run from the db.
func (a *Application) get(ctx context.Context, runID string) (*Run, error) {
	subject, err := a.CanAccessRun(ctx, rbac.GetRunAction, runID)
	if err != nil {
		return nil, err
	}

	run, err := a.db.GetRun(ctx, runID)
	if err != nil {
		a.Error(err, "retrieving run", "id", runID, "subject", subject)
		return nil, err
	}
	a.V(2).Info("retrieved run", "id", runID, "subject", subject)

	return run, nil
}

// ListRuns retrieves multiple run objs. Use opts to filter and paginate the
// list.
func (a *Application) list(ctx context.Context, opts otf.RunListOptions) (*RunList, error) {
	var subject otf.Subject
	var err error
	if opts.Organization != nil && opts.WorkspaceName != nil {
		// subject needs perms on workspace to list runs in workspace
		subject, err = a.CanAccessWorkspaceByName(ctx, rbac.GetWorkspaceAction,
			*opts.WorkspaceName,
			*opts.Organization,
		)
	} else if opts.WorkspaceID != nil {
		// subject needs perms on workspace to list runs in workspace
		subject, err = a.CanAccessWorkspaceByID(ctx, rbac.GetWorkspaceAction, *opts.WorkspaceID)
	} else if opts.Organization != nil {
		// subject needs perms on org to list runs in org
		subject, err = a.CanAccessOrganization(ctx, rbac.ListRunsAction, *opts.Organization)
	} else {
		// subject needs to be site admin to list runs across site
		subject, err = a.CanAccessSite(ctx, rbac.ListRunsAction)
	}
	if err != nil {
		return nil, err
	}

	rl, err := a.db.ListRuns(ctx, opts)
	if err != nil {
		a.Error(err, "listing runs", "subject", subject)
		return nil, err
	}

	a.V(2).Info("listed runs", "count", len(rl.Items), "subject", subject)

	return rl, nil
}

// apply enqueues an apply for the run.
func (a *Application) apply(ctx context.Context, runID string) error {
	subject, err := a.CanAccessRun(ctx, rbac.ApplyRunAction, runID)
	if err != nil {
		return err
	}
	run, err := a.db.UpdateStatus(ctx, runID, func(run *Run) error {
		return run.EnqueueApply()
	})
	if err != nil {
		a.Error(err, "enqueuing apply", "id", runID, "subject", subject)
		return err
	}

	a.V(0).Info("enqueued apply", "id", runID, "subject", subject)

	a.Publish(otf.Event{Type: otf.EventRunStatusUpdate, Payload: run})

	return err
}

// discard discards the run.
func (a *Application) discard(ctx context.Context, runID string) error {
	subject, err := a.CanAccessRun(ctx, rbac.DiscardRunAction, runID)
	if err != nil {
		return err
	}

	run, err := a.db.UpdateStatus(ctx, runID, func(run *Run) error {
		return run.Discard()
	})
	if err != nil {
		a.Error(err, "discarding run", "id", runID, "subject", subject)
		return err
	}

	a.V(0).Info("discarded run", "id", runID, "subject", subject)

	a.Publish(otf.Event{Type: otf.EventRunStatusUpdate, Payload: run})

	return err
}

// cancel a run. If a run is in progress then a cancelation signal will be
// sent out.
func (a *Application) cancel(ctx context.Context, runID string) error {
	subject, err := a.CanAccessRun(ctx, rbac.CancelRunAction, runID)
	if err != nil {
		return err
	}

	var enqueue bool
	run, err := a.db.UpdateStatus(ctx, runID, func(run *Run) (err error) {
		enqueue, err = run.Cancel()
		return err
	})
	if err != nil {
		a.Error(err, "canceling run", "id", runID, "subject", subject)
		return err
	}
	a.V(0).Info("canceled run", "id", runID, "subject", subject)
	if enqueue {
		// notify agent which'll send a SIGINT to terraform
		a.Publish(otf.Event{Type: otf.EventRunCancel, Payload: run})
	}
	a.Publish(otf.Event{Type: otf.EventRunStatusUpdate, Payload: run})
	return nil
}

// ForceCancelRun forcefully cancels a run.
func (a *Application) forceCancel(ctx context.Context, runID string) error {
	subject, err := a.CanAccessRun(ctx, rbac.CancelRunAction, runID)
	if err != nil {
		return err
	}
	run, err := a.db.UpdateStatus(ctx, runID, func(run *Run) error {
		return run.ForceCancel()
	})
	if err != nil {
		a.Error(err, "force canceling run", "id", runID, "subject", subject)
		return err
	}
	a.V(0).Info("force canceled run", "id", runID, "subject", subject)

	// notify agent which'll send a SIGKILL to terraform
	a.Publish(otf.Event{Type: otf.EventRunForceCancel, Payload: run})

	return err
}

// enqueuePlan enqueues a plan for the run.
//
// NOTE: this is an internal action, invoked by the scheduler only.
func (a *Application) enqueuePlan(ctx context.Context, runID string) (*Run, error) {
	subject, err := a.CanAccessRun(ctx, rbac.EnqueuePlanAction, runID)
	if err != nil {
		return nil, err
	}

	run, err := a.db.UpdateStatus(ctx, runID, func(run *Run) error {
		return run.EnqueuePlan()
	})
	if err != nil {
		a.Error(err, "enqueuing plan", "id", runID, "subject", subject)
		return nil, err
	}
	a.V(0).Info("enqueued plan", "id", runID, "subject", subject)

	a.Publish(otf.Event{Type: otf.EventRunStatusUpdate, Payload: run})

	return run, nil
}

// getPlanFile returns the plan file for the run.
func (a *Application) getPlanFile(ctx context.Context, runID string, format otf.PlanFormat) ([]byte, error) {
	subject, err := a.CanAccessRun(ctx, rbac.GetPlanFileAction, runID)
	if err != nil {
		return nil, err
	}

	if plan, err := a.cache.Get(format.CacheKey(runID)); err == nil {
		return plan, nil
	}
	// Cache is empty; retrieve from DB
	file, err := a.db.GetPlanFile(ctx, runID, format)
	if err != nil {
		a.Error(err, "retrieving plan file", "id", runID, "format", format, "subject", subject)
		return nil, err
	}
	// Cache plan before returning
	if err := a.cache.Set(format.CacheKey(runID), file); err != nil {
		return nil, fmt.Errorf("caching plan: %w", err)
	}
	return file, nil
}

// uploadPlanFile persists a run's plan file. The plan format should be either
// be binary or json.
func (a *Application) uploadPlanFile(ctx context.Context, runID string, plan []byte, format otf.PlanFormat) error {
	subject, err := a.CanAccessRun(ctx, rbac.UploadPlanFileAction, runID)
	if err != nil {
		return err
	}

	if err := a.db.SetPlanFile(ctx, runID, plan, format); err != nil {
		a.Error(err, "uploading plan file", "id", runID, "format", format, "subject", subject)
		return err
	}

	a.V(1).Info("uploaded plan file", "id", runID, "format", format, "subject", subject)

	if err := a.cache.Set(format.CacheKey(runID), plan); err != nil {
		return fmt.Errorf("caching plan: %w", err)
	}

	return nil
}

// delete a run.
func (a *Application) delete(ctx context.Context, runID string) error {
	subject, err := a.CanAccessRun(ctx, rbac.DeleteRunAction, runID)
	if err != nil {
		return err
	}

	// get run first so that we can include it in an event below
	run, err := a.db.GetRun(ctx, runID)
	if err != nil {
		return err
	}
	if err := a.db.DeleteRun(ctx, runID); err != nil {
		a.Error(err, "deleting run", "id", runID, "subject", subject)
		return err
	}
	a.V(0).Info("deleted run", "id", runID, "subject", subject)
	a.Publish(otf.Event{Type: otf.EventRunDeleted, Payload: run})
	return nil
}

// startPhase starts a run phase.
func (a *Application) startPhase(ctx context.Context, runID string, phase otf.PhaseType, _ otf.PhaseStartOptions) (*Run, error) {
	subject, err := a.CanAccessRun(ctx, rbac.StartPhaseAction, runID)
	if err != nil {
		return nil, err
	}

	run, err := a.db.UpdateStatus(ctx, runID, func(run *Run) error {
		return run.Start(phase)
	})
	if err != nil {
		a.Error(err, "starting "+string(phase), "id", runID, "subject", subject)
		return nil, err
	}
	a.V(0).Info("started "+string(phase), "id", runID, "subject", subject)
	a.Publish(otf.Event{Type: otf.EventRunStatusUpdate, Payload: run})
	return run, nil
}

// finishPhase finishes a phase. Creates a report of changes before updating the status of
// the run.
func (a *Application) finishPhase(ctx context.Context, runID string, phase otf.PhaseType, opts otf.PhaseFinishOptions) (*Run, error) {
	subject, err := a.CanAccessRun(ctx, rbac.FinishPhaseAction, runID)
	if err != nil {
		return nil, err
	}

	var report ResourceReport
	if !opts.Errored {
		var err error
		report, err = a.createReport(ctx, runID, phase)
		if err != nil {
			a.Error(err, "creating report", "id", runID, "phase", phase, "subject", subject)
			opts.Errored = true
		}
	}
	run, err := a.db.UpdateStatus(ctx, runID, func(run *Run) error {
		return run.Finish(phase, opts)
	})
	if err != nil {
		a.Error(err, "finishing "+string(phase), "id", runID, "subject", subject)
		return nil, err
	}
	a.V(0).Info("finished "+string(phase), "id", runID,
		"additions", report.Additions,
		"changes", report.Changes,
		"destructions", report.Destructions,
		"subject", subject,
	)
	a.Publish(otf.Event{Type: otf.EventRunStatusUpdate, Payload: run})
	return run, nil
}

// createReport creates a report of changes for the phase.
func (a *Application) createReport(ctx context.Context, runID string, phase otf.PhaseType) (ResourceReport, error) {
	switch phase {
	case otf.PlanPhase:
		return a.createPlanReport(ctx, runID)
	case otf.ApplyPhase:
		return a.createApplyReport(ctx, runID)
	default:
		return ResourceReport{}, fmt.Errorf("unknown supported phase for creating report: %s", phase)
	}
}

func (a *Application) createPlanReport(ctx context.Context, runID string) (ResourceReport, error) {
	plan, err := a.getPlanFile(ctx, runID, otf.PlanFormatJSON)
	if err != nil {
		return ResourceReport{}, err
	}
	report, err := CompilePlanReport(plan)
	if err != nil {
		return ResourceReport{}, err
	}
	if err := a.db.CreatePlanReport(ctx, runID, report); err != nil {
		return ResourceReport{}, err
	}
	return report, nil
}

func (a *Application) createApplyReport(ctx context.Context, runID string) (ResourceReport, error) {
	logs, err := a.GetChunk(ctx, logs.GetChunkOptions{
		RunID: runID,
		Phase: otf.ApplyPhase,
	})
	if err != nil {
		return ResourceReport{}, err
	}
	report, err := ParseApplyOutput(string(logs.Data))
	if err != nil {
		return ResourceReport{}, err
	}
	if err := a.db.CreateApplyReport(ctx, runID, report); err != nil {
		return ResourceReport{}, err
	}
	return report, nil
}
