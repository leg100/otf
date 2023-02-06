package run

import (
	"context"
	"fmt"

	"github.com/leg100/otf"
	"github.com/leg100/otf/rbac"
)

// CreateRun creates a run. Caller needs to have created a config version first.
func (a *Application) CreateRun(ctx context.Context, workspaceID string, opts otf.RunCreateOptions) (*otf.Run, error) {
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
func (a *Application) GetRun(ctx context.Context, runID string) (*otf.Run, error) {
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
func (a *Application) ListRuns(ctx context.Context, opts otf.RunListOptions) (*otf.RunList, error) {
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

	a.V(2).Info("listed runs", append(opts.LogFields(), "count", len(rl.Items), "subject", subject)...)

	return rl, nil
}

// ApplyRun enqueues an apply for the run.
func (a *Application) ApplyRun(ctx context.Context, runID string, opts otf.RunApplyOptions) error {
	subject, err := a.CanAccessRun(ctx, rbac.ApplyRunAction, runID)
	if err != nil {
		return err
	}
	run, err := a.db.UpdateStatus(ctx, runID, func(run *otf.Run) error {
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

// DiscardRun the run.
func (a *Application) DiscardRun(ctx context.Context, runID string, opts otf.RunDiscardOptions) error {
	subject, err := a.CanAccessRun(ctx, rbac.DiscardRunAction, runID)
	if err != nil {
		return err
	}

	run, err := a.db.UpdateStatus(ctx, runID, func(run *otf.Run) error {
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

// CancelRun a run. If a run is in progress then a cancelation signal will be
// sent out.
func (a *Application) CancelRun(ctx context.Context, runID string, opts otf.RunCancelOptions) error {
	subject, err := a.CanAccessRun(ctx, rbac.CancelRunAction, runID)
	if err != nil {
		return err
	}

	var enqueue bool
	run, err := a.db.UpdateStatus(ctx, runID, func(run *otf.Run) (err error) {
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
func (a *Application) ForceCancelRun(ctx context.Context, runID string, opts otf.RunForceCancelOptions) error {
	subject, err := a.CanAccessRun(ctx, rbac.CancelRunAction, runID)
	if err != nil {
		return err
	}
	run, err := a.db.UpdateStatus(ctx, runID, func(run *otf.Run) error {
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

// EnqueuePlan enqueues a plan for the run.
//
// NOTE: this is an internal action, invoked by the scheduler only.
func (a *Application) EnqueuePlan(ctx context.Context, runID string) (*otf.Run, error) {
	subject, err := a.CanAccessRun(ctx, rbac.EnqueuePlanAction, runID)
	if err != nil {
		return nil, err
	}

	run, err := a.DB().UpdateStatus(ctx, runID, func(run *otf.Run) error {
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

// GetPlanFile returns the plan file for the run.
func (a *Application) GetPlanFile(ctx context.Context, runID string, format otf.PlanFormat) ([]byte, error) {
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

// UploadPlanFile persists a run's plan file. The plan format should be either
// be binary or json.
func (a *Application) UploadPlanFile(ctx context.Context, runID string, plan []byte, format otf.PlanFormat) error {
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

// GetLockFile returns the lock file for the run.
func (a *Application) GetLockFile(ctx context.Context, runID string) ([]byte, error) {
	subject, err := a.CanAccessRun(ctx, rbac.GetLockFileAction, runID)
	if err != nil {
		return nil, err
	}

	if plan, err := a.cache.Get(otf.LockFileCacheKey(runID)); err == nil {
		return plan, nil
	}
	// Cache is empty; retrieve from DB
	file, err := a.db.GetLockFile(ctx, runID)
	if err != nil {
		a.Error(err, "retrieving lock file", "id", runID, "subject", subject)
		return nil, err
	}
	// Cache plan before returning
	if err := a.cache.Set(otf.LockFileCacheKey(runID), file); err != nil {
		return nil, fmt.Errorf("caching lock file: %w", err)
	}
	return file, nil
}

// UploadLockFile persists the lock file for a run.
func (a *Application) UploadLockFile(ctx context.Context, runID string, plan []byte) error {
	subject, err := a.CanAccessRun(ctx, rbac.UploadLockFileAction, runID)
	if err != nil {
		return err
	}

	if err := a.db.SetLockFile(ctx, runID, plan); err != nil {
		a.Error(err, "uploading lock file", "id", runID, "subject", subject)
		return err
	}
	a.V(2).Info("uploaded lock file", "id", runID)
	if err := a.cache.Set(otf.LockFileCacheKey(runID), plan); err != nil {
		return fmt.Errorf("caching plan: %w", err)
	}
	return nil
}

// DeleteRun deletes a run.
func (a *Application) DeleteRun(ctx context.Context, runID string) error {
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

// StartPhase starts a run phase.
func (a *Application) StartPhase(ctx context.Context, runID string, phase otf.PhaseType, _ otf.PhaseStartOptions) (*otf.Run, error) {
	subject, err := a.CanAccessRun(ctx, rbac.StartPhaseAction, runID)
	if err != nil {
		return nil, err
	}

	run, err := a.db.UpdateStatus(ctx, runID, func(run *otf.Run) error {
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

// FinishPhase finishes a phase. Creates a report of changes before updating the status of
// the run.
func (a *Application) FinishPhase(ctx context.Context, runID string, phase otf.PhaseType, opts otf.PhaseFinishOptions) (*otf.Run, error) {
	subject, err := a.CanAccessRun(ctx, rbac.FinishPhaseAction, runID)
	if err != nil {
		return nil, err
	}

	var report otf.ResourceReport
	if !opts.Errored {
		var err error
		report, err = a.createReport(ctx, runID, phase)
		if err != nil {
			a.Error(err, "creating report", "id", runID, "phase", phase, "subject", subject)
			opts.Errored = true
		}
	}
	run, err := a.db.UpdateStatus(ctx, runID, func(run *otf.Run) error {
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

// GetChunk reads a chunk of logs for a phase.
//
// NOTE: unauthenticated - access granted only via signed URL
func (a *Application) GetChunk(ctx context.Context, opts otf.GetChunkOptions) (otf.Chunk, error) {
	logs, err := a.proxy.GetChunk(ctx, opts)
	if err == otf.ErrResourceNotFound {
		// ignore resource not found because no log chunks may not have been
		// written yet
		return otf.Chunk{}, nil
	} else if err != nil {
		a.Error(err, "reading logs", "id", opts.RunID, "offset", opts.Offset, "limit", opts.Limit)
		return otf.Chunk{}, err
	}
	a.V(2).Info("read logs", "id", opts.RunID, "offset", opts.Offset, "limit", opts.Limit)
	return logs, nil
}

// PutChunk writes a chunk of logs for a phase.
func (a *Application) PutChunk(ctx context.Context, chunk otf.Chunk) error {
	_, err := a.CanAccessRun(ctx, rbac.PutChunkAction, chunk.RunID)
	if err != nil {
		return err
	}

	persisted, err := a.db.PutChunk(ctx, chunk)
	if err != nil {
		a.Error(err, "writing logs", "id", chunk.RunID, "phase", chunk.Phase, "offset", chunk.Offset)
		return err
	}
	a.V(2).Info("written logs", "id", chunk.RunID, "phase", chunk.Phase, "offset", chunk.Offset)

	a.Publish(otf.Event{
		Type:    otf.EventLogChunk,
		Payload: persisted,
	})

	return nil
}

// Tail logs for a phase. Offset specifies the number of bytes into the logs
// from which to start tailing.
func (a *Application) Tail(ctx context.Context, opts otf.GetChunkOptions) (<-chan otf.Chunk, error) {
	subject, err := a.CanAccessRun(ctx, rbac.TailLogsAction, opts.RunID)
	if err != nil {
		return nil, err
	}

	// Subscribe first and only then retrieve from DB, guaranteeing that we
	// won't miss any updates
	sub, err := a.Subscribe(ctx, "tail-"+otf.GenerateRandomString(6))
	if err != nil {
		return nil, err
	}

	chunk, err := a.proxy.GetChunk(ctx, opts)
	if err == otf.ErrResourceNotFound {
		// ignore resource not found because no log chunks may not have been
		// written yet
	} else if err != nil {
		a.Error(err, "tailing logs", "id", opts.RunID, "offset", opts.Offset, "subject", subject)
		return nil, err
	}
	opts.Offset += len(chunk.Data)

	ch := make(chan otf.Chunk)
	go func() {
		// send existing chunk
		if len(chunk.Data) > 0 {
			ch <- chunk
		}

		for {
			select {
			case ev, ok := <-sub:
				if !ok {
					close(ch)
					return
				}
				chunk, ok := ev.Payload.(otf.PersistedChunk)
				if !ok {
					// skip non-log events
					continue
				}
				if opts.RunID != chunk.RunID || opts.Phase != chunk.Phase {
					// skip logs for different run/phase
					continue
				}
				if chunk.Offset < opts.Offset {
					// chunk has overlapping offset

					if chunk.Offset+len(chunk.Data) <= opts.Offset {
						// skip entirely overlapping chunk
						continue
					}
					// remove overlapping portion of chunk
					chunk.Chunk = chunk.Cut(otf.GetChunkOptions{Offset: opts.Offset})
				}
				if len(chunk.Data) == 0 {
					// don't send empty chunks
					continue
				}
				ch <- chunk.Chunk
				if chunk.IsEnd() {
					close(ch)
					return
				}
			case <-ctx.Done():
				close(ch)
				return
			}
		}
	}()
	a.V(2).Info("tailing logs", "id", opts.RunID, "phase", opts.Phase, "subject", subject)
	return ch, nil
}

// createReport creates a report of changes for the phase.
func (a *Application) createReport(ctx context.Context, runID string, phase otf.PhaseType) (otf.ResourceReport, error) {
	switch phase {
	case otf.PlanPhase:
		return a.createPlanReport(ctx, runID)
	case otf.ApplyPhase:
		return a.createApplyReport(ctx, runID)
	default:
		return otf.ResourceReport{}, fmt.Errorf("unknown supported phase for creating report: %s", phase)
	}
}

func (a *Application) createPlanReport(ctx context.Context, runID string) (otf.ResourceReport, error) {
	plan, err := a.GetPlanFile(ctx, runID, otf.PlanFormatJSON)
	if err != nil {
		return otf.ResourceReport{}, err
	}
	report, err := otf.CompilePlanReport(plan)
	if err != nil {
		return otf.ResourceReport{}, err
	}
	if err := a.db.CreatePlanReport(ctx, runID, report); err != nil {
		return otf.ResourceReport{}, err
	}
	return report, nil
}

func (a *Application) createApplyReport(ctx context.Context, runID string) (otf.ResourceReport, error) {
	logs, err := a.GetChunk(ctx, otf.GetChunkOptions{
		RunID: runID,
		Phase: otf.ApplyPhase,
	})
	if err != nil {
		return otf.ResourceReport{}, err
	}
	report, err := otf.ParseApplyOutput(string(logs.Data))
	if err != nil {
		return otf.ResourceReport{}, err
	}
	if err := a.db.CreateApplyReport(ctx, runID, report); err != nil {
		return otf.ResourceReport{}, err
	}
	return report, nil
}
