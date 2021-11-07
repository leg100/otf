package app

import (
	"context"
	"fmt"
	"io"
	"time"

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
	if err := opts.Valid(); err != nil {
		s.Error(err, "creating invalid run")
		return nil, err
	}

	run, err := s.NewRun(opts)
	if err != nil {
		s.Error(err, "constructing new run")
		return nil, err
	}

	_, err = s.db.Create(run)
	if err != nil {
		s.Error(err, "creating run", "id", run.ID)
		return nil, err
	}

	s.V(1).Info("created run", "id", run.ID)

	s.es.Publish(otf.Event{Type: otf.EventRunCreated, Payload: run})

	return run, nil
}

// Get retrieves a run obj with the given ID from the db.
func (s RunService) Get(id string) (*otf.Run, error) {
	run, err := s.db.Get(otf.RunGetOptions{ID: &id})
	if err != nil {
		s.Error(err, "retrieving run", "id", id)
		return nil, err
	}

	s.V(2).Info("retrieved run", "id", run.ID)

	return run, nil
}

// List retrieves multiple run objs. Use opts to filter and paginate the list.
func (s RunService) List(opts otf.RunListOptions) (*otf.RunList, error) {
	rl, err := s.db.List(opts)
	if err != nil {
		s.Error(err, "listing runs")
		return nil, err
	}

	s.V(2).Info("listed runs", "count", len(rl.Items))

	return rl, nil
}

func (s RunService) Apply(id string, opts otf.RunApplyOptions) error {
	run, err := s.db.Update(otf.RunGetOptions{ID: otf.String(id)}, func(run *otf.Run) error {
		run.UpdateStatus(otf.RunApplyQueued)

		return nil
	})
	if err != nil {
		s.Error(err, "applying run", "id", id)
		return err
	}

	s.V(0).Info("applied run", "id", id)

	s.es.Publish(otf.Event{Type: otf.EventApplyQueued, Payload: run})

	return err
}

func (s RunService) Discard(id string, opts otf.RunDiscardOptions) error {
	run, err := s.db.Update(otf.RunGetOptions{ID: otf.String(id)}, func(run *otf.Run) error {
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
func (s RunService) Cancel(id string, opts otf.RunCancelOptions) error {
	_, err := s.db.Update(otf.RunGetOptions{ID: otf.String(id)}, func(run *otf.Run) error {
		return run.Cancel()
	})
	return err
}

func (s RunService) ForceCancel(id string, opts otf.RunForceCancelOptions) error {
	_, err := s.db.Update(otf.RunGetOptions{ID: otf.String(id)}, func(run *otf.Run) error {
		if err := run.ForceCancel(); err != nil {
			return err
		}

		// TODO: send KILL signal to running terraform process

		// TODO: unlock workspace

		return nil
	})

	return err
}

func (s RunService) EnqueuePlan(id string) error {
	run, err := s.db.Update(otf.RunGetOptions{ID: otf.String(id)}, func(run *otf.Run) error {
		run.UpdateStatus(otf.RunPlanQueued)
		return nil
	})
	if err != nil {
		s.Error(err, "enqueuing plan", "id", id)
		return err
	}

	s.V(0).Info("enqueued plan", "id", id)

	s.es.Publish(otf.Event{Type: otf.EventPlanQueued, Payload: run})

	return err
}

// GetPlanFile returns the plan file for the run.
func (s RunService) GetPlanFile(ctx context.Context, runID string, opts otf.PlanFileOptions) ([]byte, error) {
	switch opts.Format {
	case otf.PlanJSONFormat:
		return s.getJSONPlanFile(ctx, runID)
	case otf.PlanBinaryFormat:
		return s.getBinaryPlanFile(ctx, runID)
	default:
		return nil, fmt.Errorf("unknown plan file format specified: %s", opts.Format)
	}
}

// UploadPlanFile persists a run's plan file. The plan file is expected to have
// been produced using `terraform plan`. If the plan file is JSON serialized
// then its parsed for a summary of planned changes and the Plan object is
// updated accordingly.
func (s RunService) UploadPlanFile(ctx context.Context, id string, plan []byte, opts otf.PlanFileOptions) error {
	switch opts.Format {
	case otf.PlanJSONFormat:
		return s.putJSONPlanFile(ctx, id, plan)
	case otf.PlanBinaryFormat:
		return s.putBinaryPlanFile(ctx, id, plan)
	default:
		return fmt.Errorf("unknown plan file format specified: %s", opts.Format)
	}
}

// GetLogs gets the logs for a run, combining the logs of both its plan and
// apply.
func (s RunService) GetLogs(ctx context.Context, id string) (io.Reader, error) {
	run, err := s.Get(id)
	if err != nil {
		s.Error(err, "getting run for reading logs", "id", id)
		return nil, err
	}

	streamer := otf.NewRunStreamer(run, s.planLogs, s.applyLogs, time.Millisecond*500)
	go streamer.Stream(ctx)

	return streamer, nil
}

// GetPlanFile returns the plan file in json format for the run.
func (s RunService) getJSONPlanFile(ctx context.Context, runID string) ([]byte, error) {
	run, err := s.db.Get(otf.RunGetOptions{ID: otf.String(runID), IncludePlanJSON: true})
	if err != nil {
		s.Error(err, "retrieving json plan file", "id", runID)
		return nil, err
	}

	return run.Plan.PlanJSON, nil
}

// GetPlanFile returns the plan file in json format for the run.
func (s RunService) getBinaryPlanFile(ctx context.Context, runID string) ([]byte, error) {
	run, err := s.db.Get(otf.RunGetOptions{ID: otf.String(runID), IncludePlanFile: true})
	if err != nil {
		s.Error(err, "retrieving binary plan file", "id", runID)
		return nil, err
	}

	return run.Plan.PlanFile, nil
}

func (s RunService) putBinaryPlanFile(ctx context.Context, id string, plan []byte) error {
	_, err := s.db.Update(otf.RunGetOptions{ID: otf.String(id)}, func(run *otf.Run) error {
		run.Plan.PlanFile = plan

		return nil
	})
	if err != nil {
		s.Error(err, "uploading binary plan file", "id", id)
		return err
	}

	if err := s.cache.Set(otf.BinaryPlanCacheKey(id), plan); err != nil {
		return fmt.Errorf("caching plan: %w", err)
	}

	s.V(0).Info("uploaded binary plan file", "id", id)

	return nil
}

func (s RunService) putJSONPlanFile(ctx context.Context, id string, plan []byte) error {
	_, err := s.db.Update(otf.RunGetOptions{ID: otf.String(id)}, func(run *otf.Run) error {
		run.Plan.PlanJSON = plan

		return run.Plan.CalculateTotals()
	})
	if err != nil {
		s.Error(err, "uploading json plan file", "id", id)
		return err
	}

	if err := s.cache.Set(otf.JSONPlanCacheKey(id), plan); err != nil {
		return fmt.Errorf("caching plan: %w", err)
	}

	s.V(0).Info("uploaded json plan file", "id", id)

	return nil
}
