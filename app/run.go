package app

import (
	"context"
	"fmt"

	"github.com/leg100/otf"
)

// CreateRun constructs and persists a new run object to the db, before
// scheduling the run.
func (a *Application) CreateRun(ctx context.Context, spec otf.WorkspaceSpec, opts otf.RunCreateOptions) (*otf.Run, error) {
	if !a.CanAccessWorkspace(ctx, spec) {
		return nil, otf.ErrAccessNotPermitted
	}

	run, err := a.NewRun(ctx, spec, opts)
	if err != nil {
		a.Error(err, "constructing new run")
		return nil, err
	}

	if err = a.db.CreateRun(ctx, run); err != nil {
		a.Error(err, "creating run", "id", run.ID())
		return nil, err
	}
	a.V(1).Info("created run", "id", run.ID())

	a.Mapper.AddRun(run)

	a.Publish(otf.Event{Type: otf.EventRunCreated, Payload: run})

	return run, nil
}

// GetRun retrieves a run from the db.
func (a *Application) GetRun(ctx context.Context, runID string) (*otf.Run, error) {
	if !a.CanAccessRun(ctx, runID) {
		return nil, otf.ErrAccessNotPermitted
	}

	run, err := a.db.GetRun(ctx, runID)
	if err != nil {
		a.Error(err, "retrieving run", "id", runID)
		return nil, err
	}
	a.V(2).Info("retrieved run", "id", runID)

	return run, nil
}

// ListRuns retrieves multiple run objs. Use opts to filter and paginate the
// list.
func (a *Application) ListRuns(ctx context.Context, opts otf.RunListOptions) (*otf.RunList, error) {
	if !otf.CanAccess(ctx, opts.OrganizationName) {
		return nil, otf.ErrAccessNotPermitted
	}

	rl, err := a.db.ListRuns(ctx, opts)
	if err != nil {
		a.Error(err, "listing runs")
		return nil, err
	}

	a.V(2).Info("listed runs", append(opts.LogFields(), "count", len(rl.Items))...)

	return rl, nil
}

// ListWatchRun lists runs and then watches for changes to runs. The run listing
// is ordered by creation date, oldest first. Note: The options filter the list
// but not the watch.
func (a *Application) ListWatchRun(ctx context.Context, opts otf.RunListOptions) (<-chan *otf.Run, error) {
	if !otf.CanAccess(ctx, opts.OrganizationName) {
		return nil, otf.ErrAccessNotPermitted
	}
	// retrieve existing runs, page by page
	var existing []*otf.Run
	for {
		page, err := a.db.ListRuns(ctx, opts)
		if err != nil {
			return nil, err
		}
		existing = append(existing, page.Items...)
		if page.NextPage() == nil {
			break
		}
		opts.PageNumber = *page.NextPage()
	}
	// db returns runs ordered by creation date, newest first, but we want
	// oldest first, so we reverse the order
	var oldest []*otf.Run
	for _, r := range existing {
		oldest = append([]*otf.Run{r}, oldest...)
	}
	// send the retrieved runs down the channel to be returned to caller
	spool := make(chan *otf.Run, len(oldest))
	for _, r := range oldest {
		spool <- r
	}
	// the same channel receives run events from here-in on
	sub, err := a.Subscribe("run-listwatch")
	if err != nil {
		return nil, err
	}
	go func() {
		for {
			select {
			case <-ctx.Done():
				// context cancelled; shutdown spooler
				close(spool)
				return
			case event, ok := <-sub.C():
				if !ok {
					// sender closed channel; shutdown spooler
					close(spool)
					return
				}
				run, ok := event.Payload.(*otf.Run)
				if !ok {
					// skip non-run events
					continue
				}
				spool <- run
			}
		}
	}()
	return spool, nil
}

// ApplyRun enqueues an apply for the run.
func (a *Application) ApplyRun(ctx context.Context, runID string, opts otf.RunApplyOptions) error {
	if !a.CanAccessRun(ctx, runID) {
		return otf.ErrAccessNotPermitted
	}
	run, err := a.db.UpdateStatus(ctx, runID, func(run *otf.Run) error {
		return run.EnqueueApply()
	})
	if err != nil {
		a.Error(err, "enqueuing apply", "id", runID)
		return err
	}

	a.V(0).Info("enqueued apply", "id", runID)

	a.Publish(otf.Event{Type: otf.EventRunStatusUpdate, Payload: run})

	return err
}

// DiscardRun the run.
func (a *Application) DiscardRun(ctx context.Context, runID string, opts otf.RunDiscardOptions) error {
	if !a.CanAccessRun(ctx, runID) {
		return otf.ErrAccessNotPermitted
	}
	run, err := a.db.UpdateStatus(ctx, runID, func(run *otf.Run) error {
		return run.Discard()
	})
	if err != nil {
		a.Error(err, "discarding run", "id", runID)
		return err
	}

	a.V(0).Info("discarded run", "id", runID)

	a.Publish(otf.Event{Type: otf.EventRunStatusUpdate, Payload: run})

	return err
}

// CancelRun a run. If a run is in progress then a cancelation signal will be
// sent out.
func (a *Application) CancelRun(ctx context.Context, runID string, opts otf.RunCancelOptions) error {
	if !a.CanAccessRun(ctx, runID) {
		return otf.ErrAccessNotPermitted
	}
	var enqueue bool
	run, err := a.db.UpdateStatus(ctx, runID, func(run *otf.Run) (err error) {
		enqueue, err = run.Cancel()
		return err
	})
	if err != nil {
		a.Error(err, "canceling run", "id", runID)
		return err
	}
	a.V(0).Info("canceled run", "id", runID)
	if enqueue {
		// notify agent which'll send a SIGINT to terraform
		a.Publish(otf.Event{Type: otf.EventRunCancel, Payload: run})
	}
	a.Publish(otf.Event{Type: otf.EventRunStatusUpdate, Payload: run})
	return nil
}

// ForceCancelRun forcefully cancels a run.
func (a *Application) ForceCancelRun(ctx context.Context, runID string, opts otf.RunForceCancelOptions) error {
	if !a.CanAccessRun(ctx, runID) {
		return otf.ErrAccessNotPermitted
	}
	run, err := a.db.UpdateStatus(ctx, runID, func(run *otf.Run) error {
		return run.ForceCancel()
	})
	if err != nil {
		a.Error(err, "force canceling run", "id", runID)
		return err
	}
	a.V(0).Info("force canceled run", "id", runID)

	// notify agent which'll send a SIGKILL to terraform
	a.Publish(otf.Event{Type: otf.EventRunForceCancel, Payload: run})

	return err
}

// EnqueuePlan enqueues a plan for the run.
func (a *Application) EnqueuePlan(ctx context.Context, runID string) (*otf.Run, error) {
	if !a.CanAccessRun(ctx, runID) {
		return nil, otf.ErrAccessNotPermitted
	}

	var run *otf.Run
	err := a.Tx(ctx, func(tx *Application) (err error) {
		run, err = tx.db.UpdateStatus(ctx, runID, func(run *otf.Run) error {
			return run.EnqueuePlan(ctx, tx)
		})
		return err
	})
	if err != nil {
		a.Error(err, "enqueuing plan", "id", runID)
		return nil, err
	}
	a.V(0).Info("enqueued plan", "id", runID)

	a.Publish(otf.Event{Type: otf.EventRunStatusUpdate, Payload: run})

	return run, nil
}

// WatchLatest watches for updates to the latest run for the specified
// workspace.
func (a *Application) WatchLatest(ctx context.Context, spec otf.WorkspaceSpec) (<-chan *otf.Run, error) {
	if !a.CanAccessWorkspace(ctx, spec) {
		return nil, otf.ErrAccessNotPermitted
	}
	return a.latest.Watch(ctx, a.LookupWorkspaceID(spec))
}

// Watch watches for updates to the specified run
func (a *Application) Watch(ctx context.Context, runID string) (<-chan *otf.Run, error) {
	if !a.CanAccessRun(ctx, runID) {
		return nil, otf.ErrAccessNotPermitted
	}
	sub, err := a.EventService.Subscribe("watch-run-" + otf.GenerateRandomString(6))
	if err != nil {
		return nil, err
	}
	c := make(chan *otf.Run)
	go func() {
		for {
			select {
			case <-ctx.Done():
				sub.Close()
				return
			case event, ok := <-sub.C():
				if !ok {
					// sender closed channel
					return
				}
				run, ok := event.Payload.(*otf.Run)
				if !ok {
					// skip non-run events
					continue
				}
				if run.ID() == runID {
					c <- run
				}
			}
		}
	}()

	return c, nil
}

// Watch watches for updates to runs belonging to the specified workspace.
func (a *Application) WatchWorkspaceRuns(ctx context.Context, spec otf.WorkspaceSpec) (<-chan *otf.Event, error) {
	if !a.CanAccessWorkspace(ctx, spec) {
		return nil, otf.ErrAccessNotPermitted
	}
	sub, err := a.EventService.Subscribe("watch-ws-runs-" + otf.GenerateRandomString(6))
	if err != nil {
		return nil, err
	}
	c := make(chan *otf.Event)
	go func() {
		for {
			select {
			case <-ctx.Done():
				sub.Close()
				return
			case event, ok := <-sub.C():
				if !ok {
					// sender closed channel
					return
				}
				run, ok := event.Payload.(*otf.Run)
				if !ok {
					// skip non-run events
					continue
				}
				if run.WorkspaceID() == a.LookupWorkspaceID(spec) {
					c <- &event
				}
			}
		}
	}()

	return c, nil
}

// GetPlanFile returns the plan file for the run.
func (a *Application) GetPlanFile(ctx context.Context, runID string, format otf.PlanFormat) ([]byte, error) {
	if !a.CanAccessRun(ctx, runID) {
		return nil, otf.ErrAccessNotPermitted
	}

	if plan, err := a.cache.Get(format.CacheKey(runID)); err == nil {
		return plan, nil
	}
	// Cache is empty; retrieve from DB
	file, err := a.db.GetPlanFile(ctx, runID, format)
	if err != nil {
		a.Error(err, "retrieving plan file", "id", runID, "format", format)
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
	if !a.CanAccessRun(ctx, runID) {
		return otf.ErrAccessNotPermitted
	}

	if err := a.db.SetPlanFile(ctx, runID, plan, format); err != nil {
		a.Error(err, "uploading plan file", "id", runID, "format", format)
		return err
	}

	a.V(1).Info("uploaded plan file", "id", runID, "format", format)

	if err := a.cache.Set(format.CacheKey(runID), plan); err != nil {
		return fmt.Errorf("caching plan: %w", err)
	}

	return nil
}

// GetLockFile returns the lock file for the run.
func (a *Application) GetLockFile(ctx context.Context, runID string) ([]byte, error) {
	if !a.CanAccessRun(ctx, runID) {
		return nil, otf.ErrAccessNotPermitted
	}

	if plan, err := a.cache.Get(otf.LockFileCacheKey(runID)); err == nil {
		return plan, nil
	}
	// Cache is empty; retrieve from DB
	file, err := a.db.GetLockFile(ctx, runID)
	if err != nil {
		a.Error(err, "retrieving lock file", "id", runID)
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
	if !a.CanAccessRun(ctx, runID) {
		return otf.ErrAccessNotPermitted
	}

	if err := a.db.SetLockFile(ctx, runID, plan); err != nil {
		a.Error(err, "uploading lock file", "id", runID)
		return err
	}
	a.V(2).Info("uploaded lock file", "id", runID)
	if err := a.cache.Set(otf.LockFileCacheKey(runID), plan); err != nil {
		return fmt.Errorf("caching plan: %w", err)
	}
	return nil
}

// DeleteRun deletes a terraform run.
func (a *Application) DeleteRun(ctx context.Context, runID string) error {
	if !a.CanAccessRun(ctx, runID) {
		return otf.ErrAccessNotPermitted
	}

	// get run first so that we can include it in an event below
	run, err := a.db.GetRun(ctx, runID)
	if err != nil {
		return err
	}
	if err := a.db.DeleteRun(ctx, runID); err != nil {
		a.Error(err, "deleting run", "id", runID)
		return err
	}
	a.V(0).Info("deleted run", "id", runID)
	a.Mapper.RemoveRun(run)
	a.Publish(otf.Event{Type: otf.EventRunDeleted, Payload: run})
	return nil
}

// StartPhase phase.
func (a *Application) StartPhase(ctx context.Context, runID string, phase otf.PhaseType, opts otf.PhaseStartOptions) (*otf.Run, error) {
	if !a.CanAccessRun(ctx, runID) {
		return nil, otf.ErrAccessNotPermitted
	}

	run, err := a.db.UpdateStatus(ctx, runID, func(run *otf.Run) error {
		return run.Start(phase)
	})
	if err != nil {
		a.Error(err, "starting "+string(phase), "id", runID)
		return nil, err
	}
	a.V(0).Info("started "+string(phase), "id", runID)
	a.Publish(otf.Event{Type: otf.EventRunStatusUpdate, Payload: run})
	return run, nil
}

// FinishPhase phase. Creates a report of changes before updating the status of
// the run.
func (a *Application) FinishPhase(ctx context.Context, runID string, phase otf.PhaseType, opts otf.PhaseFinishOptions) (*otf.Run, error) {
	if !a.CanAccessRun(ctx, runID) {
		return nil, otf.ErrAccessNotPermitted
	}

	var report otf.ResourceReport
	if !opts.Errored {
		var err error
		report, err = a.CreateReport(ctx, runID, phase)
		if err != nil {
			a.Error(err, "creating report", "id", runID, "phase", phase)
			opts.Errored = true
		}
	}
	run, err := a.db.UpdateStatus(ctx, runID, func(run *otf.Run) error {
		return run.Finish(phase, opts)
	})
	if err != nil {
		a.Error(err, "finishing "+string(phase), "id", runID)
		return nil, err
	}
	a.V(0).Info("finished "+string(phase), "id", runID,
		"additions", report.Additions,
		"changes", report.Changes,
		"destructions", report.Destructions,
	)
	a.Publish(otf.Event{Type: otf.EventRunStatusUpdate, Payload: run})
	return run, nil
}

// CreateReport creates a report of changes for the phase.
func (a *Application) CreateReport(ctx context.Context, runID string, phase otf.PhaseType) (otf.ResourceReport, error) {
	switch phase {
	case otf.PlanPhase:
		return a.createPlanReport(ctx, runID)
	case otf.ApplyPhase:
		return a.createApplyReport(ctx, runID)
	default:
		return otf.ResourceReport{}, fmt.Errorf("unknown supported phase for creating report: %s", phase)
	}
}

// GetChunk reads a chunk of logs for a phase.
func (a *Application) GetChunk(ctx context.Context, runID string, phase otf.PhaseType, opts otf.GetChunkOptions) (otf.Chunk, error) {
	logs, err := a.proxy.GetChunk(ctx, runID, phase, opts)
	if err != nil {
		a.Error(err, "reading logs", "id", runID, "offset", opts.Offset, "limit", opts.Limit)
		return otf.Chunk{}, err
	}
	a.V(2).Info("read logs", "id", runID, "offset", opts.Offset, "limit", opts.Limit)
	return logs, nil
}

// PutChunk writes a chunk of logs for a phase.
func (a *Application) PutChunk(ctx context.Context, runID string, phase otf.PhaseType, chunk otf.Chunk) error {
	// pass chunk onto tail server for relaying onto clients
	err := a.tailServer.PutChunk(ctx, otf.PhaseSpec{RunID: runID, Phase: phase}, chunk)
	if err != nil {
		a.Error(err, "writing logs", "id", runID, "start", chunk.Start, "end", chunk.End, "data", string(chunk.Data))
		return err
	}
	a.V(2).Info("written logs", "id", runID, "start", chunk.Start, "end", chunk.End, "data", string(chunk.Data))

	return nil
}

// Tail logs for a phase. Offset specifies the number of bytes into the logs
// from which to start tailing.
func (a *Application) Tail(ctx context.Context, runID string, phase otf.PhaseType, offset int) (otf.TailClient, error) {
	// register hook
	client, err := a.tailServer.Tail(ctx, otf.PhaseSpec{RunID: runID, Phase: phase}, offset)
	if err != nil {
		a.Error(err, "tailing logs", "id", runID, "phase", phase)
		return nil, err
	}
	a.V(2).Info("tailing logs", "id", runID, "phase", phase)

	return client, nil
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
	logs, err := a.GetChunk(ctx, runID, otf.ApplyPhase, otf.GetChunkOptions{})
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
