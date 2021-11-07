package app

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/leg100/otf"
)

var _ otf.PlanService = (*PlanService)(nil)

type PlanService struct {
	db otf.RunStore

	logs otf.ChunkStore

	otf.EventService

	cache otf.Cache

	logr.Logger
}

func NewPlanService(db otf.RunStore, logs otf.ChunkStore, logger logr.Logger, es otf.EventService, cache otf.Cache) *PlanService {
	return &PlanService{
		db:           db,
		EventService: es,
		logs:         logs,
		cache:        cache,
		Logger:       logger,
	}
}

func (s PlanService) Get(id string) (*otf.Plan, error) {
	run, err := s.db.Get(otf.RunGetOptions{PlanID: &id})
	if err != nil {
		return nil, err
	}
	return run.Plan, nil
}

// GetPlanJSON returns the JSON formatted plan file for the plan.
func (s PlanService) GetPlanJSON(id string) ([]byte, error) {
	if plan, err := s.cache.Get(otf.JSONPlanCacheKey(id)); err == nil {
		return plan, nil
	}

	run, err := s.db.Get(otf.RunGetOptions{PlanID: &id})
	if err != nil {
		return nil, err
	}

	if err := s.cache.Set(otf.JSONPlanCacheKey(id), run.Plan.PlanJSON); err != nil {
		return nil, fmt.Errorf("caching plan: %w", err)
	}

	return run.Plan.PlanJSON, nil
}

// GetChunk reads a chunk of logs for a terraform plan.
func (s PlanService) GetChunk(ctx context.Context, id string, opts otf.GetChunkOptions) ([]byte, error) {
	logs, err := s.logs.GetChunk(ctx, id, opts)
	if err != nil {
		s.Error(err, "reading plan logs", "id", id, "offset", opts.Offset, "limit", opts.Limit)
		return nil, err
	}

	return logs, nil
}

// PutChunk writes a chunk of logs for a terraform plan.
func (s PlanService) PutChunk(ctx context.Context, id string, chunk []byte, opts otf.PutChunkOptions) error {
	err := s.logs.PutChunk(ctx, id, chunk, opts)
	if err != nil {
		s.Error(err, "writing plan logs", "id", id, "start", opts.Start, "end", opts.End)
		return err
	}

	return nil
}

// Start marks a plan as having started
func (s PlanService) Start(ctx context.Context, id string, opts otf.JobStartOptions) (*otf.Run, error) {
	run, err := s.db.Update(otf.RunGetOptions{PlanID: otf.String(id)}, func(run *otf.Run) error {
		return run.Plan.Start(run)
	})
	if err != nil {
		s.Error(err, "starting plan")
		return nil, err
	}

	s.V(0).Info("started plan", "id", run.ID)

	return run, nil
}

// Finish marks a plan as having finished.  An event is emitted to notify any
// subscribers of the new state.
func (s PlanService) Finish(ctx context.Context, id string, opts otf.JobFinishOptions) (*otf.Run, error) {
	var event *otf.Event

	run, err := s.db.Update(otf.RunGetOptions{PlanID: otf.String(id)}, func(run *otf.Run) (err error) {
		event, err = run.Plan.Finish(run)
		if err != nil {
			return err
		}
		return err
	})
	if err != nil {
		s.Error(err, "finishing plan", "id", id)
		return nil, err
	}

	s.V(0).Info("finished plan", "id", id)

	s.Publish(*event)

	return run, nil
}
