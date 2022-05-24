package app

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/leg100/otf"
)

var _ otf.RunService = (*RunService)(nil)

type RunService struct {
	db otf.RunStore
	es otf.EventService

	planLogs  otf.ChunkStore
	applyLogs otf.ChunkStore

	cache otf.Cache

	logr.Logger

	*otf.RunFactory
}

func NewRunService(db otf.RunStore, logger logr.Logger, wss otf.WorkspaceService, cvs otf.ConfigurationVersionService, es otf.EventService, planLogs, applyLogs otf.ChunkStore, cache otf.Cache) *RunService {
	return &RunService{
		db:        db,
		es:        es,
		planLogs:  planLogs,
		applyLogs: applyLogs,
		cache:     cache,
		Logger:    logger,
		RunFactory: &otf.RunFactory{
			WorkspaceService:            wss,
			ConfigurationVersionService: cvs,
		},
	}
}

// Create constructs and persists a new run object to the db, before scheduling
// the run.
func (s RunService) Create(ctx context.Context, opts otf.RunCreateOptions) (*otf.Run, error) {
	run, err := s.New(opts)
	if err != nil {
		s.Error(err, "constructing new run")
		return nil, err
	}

	if err = s.db.Create(run); err != nil {
		s.Error(err, "creating run", "id", run.ID())
		return nil, err
	}

	s.V(1).Info("created run", "id", run.ID())

	s.es.Publish(otf.Event{Type: otf.EventRunCreated, Payload: run})
	if run.Speculative() {
		s.es.Publish(otf.Event{Type: otf.EventPlanQueued, Payload: run})
	}

	return run, nil
}

// Get retrieves a run obj with the given ID from the db.
func (s RunService) Get(ctx context.Context, id string) (*otf.Run, error) {
	run, err := s.db.Get(otf.RunGetOptions{ID: &id})
	if err != nil {
		s.Error(err, "retrieving run", "id", id)
		return nil, err
	}

	s.V(2).Info("retrieved run", "id", run.ID())

	return run, nil
}

// List retrieves multiple run objs. Use opts to filter and paginate the list.
func (s RunService) List(ctx context.Context, opts otf.RunListOptions) (*otf.RunList, error) {
	rl, err := s.db.List(opts)
	if err != nil {
		s.Error(err, "listing runs")
		return nil, err
	}

	s.V(2).Info("listed runs", "count", len(rl.Items))

	return rl, nil
}

func (s RunService) Apply(ctx context.Context, id string, opts otf.RunApplyOptions) error {
	run, err := s.db.UpdateStatus(otf.RunGetOptions{ID: &id}, func(run *otf.Run) error {
		return run.ApplyRun()
	})
	if err != nil {
		s.Error(err, "applying run", "id", id)
		return err
	}

	s.V(0).Info("applied run", "id", id)

	s.es.Publish(otf.Event{Type: otf.EventApplyQueued, Payload: run})

	return err
}

func (s RunService) Discard(ctx context.Context, id string, opts otf.RunDiscardOptions) error {
	run, err := s.db.UpdateStatus(otf.RunGetOptions{ID: &id}, func(run *otf.Run) error {
		return run.Discard()
	})
	if err != nil {
		s.Error(err, "discarding run", "id", id)
		return err
	}

	s.V(0).Info("discarded run", "id", id)

	s.es.Publish(otf.Event{Type: otf.EventRunCompleted, Payload: run})

	return err
}

// Cancel enqueues a cancel request to cancel a currently queued or active plan
// or apply.
func (s RunService) Cancel(ctx context.Context, id string, opts otf.RunCancelOptions) error {
	_, err := s.db.UpdateStatus(otf.RunGetOptions{ID: &id}, func(run *otf.Run) error {
		return run.Cancel()
	})
	return err
}

func (s RunService) ForceCancel(ctx context.Context, id string, opts otf.RunForceCancelOptions) error {
	_, err := s.db.UpdateStatus(otf.RunGetOptions{ID: &id}, func(run *otf.Run) error {
		return run.ForceCancel()

		// TODO: send KILL signal to running terraform process

		// TODO: unlock workspace - could do this by publishing event and having
		// workspace scheduler subscribe to event
	})

	return err
}

func (s RunService) Start(ctx context.Context, id string) (*otf.Run, error) {
	run, err := s.db.UpdateStatus(otf.RunGetOptions{ID: &id}, func(run *otf.Run) error {
		return run.EnqueuePlan()
	})
	if err != nil {
		s.Error(err, "started run", "id", id)
		return nil, err
	}

	s.V(0).Info("started run", "id", id)

	s.es.Publish(otf.Event{Type: otf.EventPlanQueued, Payload: run})

	return run, err
}

// GetPlanFile returns the plan file for the run.
func (s RunService) GetPlanFile(ctx context.Context, spec otf.RunGetOptions, format otf.PlanFormat) ([]byte, error) {
	var id string

	// We need the run ID so if caller has specified plan or apply ID instead
	// then we need to get run ID first
	if spec.PlanID != nil || spec.ApplyID != nil {
		run, err := s.db.Get(spec)
		if err != nil {
			s.Error(err, "retrieving run for plan file", "id", spec.String())
			return nil, err
		}
		id = run.ID()
	} else {
		id = *spec.ID
	}

	// Now use run ID to look up cache
	if plan, err := s.cache.Get(format.CacheKey(id)); err == nil {
		return plan, nil
	}

	file, err := s.db.GetPlanFile(id, format)
	if err != nil {
		s.Error(err, "retrieving plan file", "id", id, "format", format)
		return nil, err
	}

	// Cache plan before returning
	if err := s.cache.Set(format.CacheKey(id), file); err != nil {
		return nil, fmt.Errorf("caching plan: %w", err)
	}

	return file, nil
}

// UploadPlanFile persists a run's plan file. The plan file is expected to have
// been produced using `terraform plan`. If the plan file is JSON serialized
// then its parsed for a summary of planned changes and the Plan object is
// updated accordingly.
func (s RunService) UploadPlanFile(ctx context.Context, runID string, plan []byte, format otf.PlanFormat) error {
	if err := s.db.SetPlanFile(runID, plan, format); err != nil {
		s.Error(err, "uploading plan file", "id", runID, "format", format)
		return err
	}

	s.V(0).Info("uploaded plan file", "id", runID, "format", format)

	if format == otf.PlanFormatJSON {
		report, err := otf.CompilePlanReport(plan)
		if err != nil {
			s.Error(err, "compiling planned changes report", "id", runID)
			return err
		}
		if err := s.db.CreatePlanReport(runID, report); err != nil {
			s.Error(err, "saving planned changes report", "id", runID)
			return err
		}
		s.V(1).Info("created planned changes report", "id", runID,
			"adds", report.Additions,
			"changes", report.Changes,
			"destructions", report.Destructions)
	}

	if err := s.cache.Set(format.CacheKey(runID), plan); err != nil {
		return fmt.Errorf("caching plan: %w", err)
	}

	return nil
}

// Delete deletes a terraform run.
func (s RunService) Delete(ctx context.Context, id string) error {
	if err := s.db.Delete(id); err != nil {
		s.Error(err, "deleting run", "id", id)
		return err
	}

	s.V(0).Info("deleted run", "id", id)

	return nil
}
