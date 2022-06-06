package app

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/leg100/otf"
	"github.com/leg100/otf/inmem"
	"github.com/leg100/otf/sql"
)

var _ otf.PlanService = (*PlanService)(nil)

type PlanService struct {
	db *sql.DB

	logs otf.ChunkStore

	otf.EventService

	cache otf.Cache

	logr.Logger
}

func NewPlanService(db *sql.DB, logs otf.ChunkStore, logger logr.Logger, es otf.EventService, cache otf.Cache) (*PlanService, error) {
	logs, err := inmem.NewChunkProxy(cache, logs)
	if err != nil {
		return nil, fmt.Errorf("constructing chunk proxy: %w", err)
	}
	return &PlanService{
		db:           db,
		EventService: es,
		logs:         logs,
		cache:        cache,
		Logger:       logger,
	}, nil
}

func (s PlanService) Get(ctx context.Context, id string) (*otf.Plan, error) {
	run, err := s.db.GetRun(ctx, otf.RunGetOptions{PlanID: &id})
	if err != nil {
		return nil, err
	}
	return run.Plan, nil
}

// GetChunk reads a chunk of logs for a terraform plan.
func (s PlanService) GetChunk(ctx context.Context, id string, opts otf.GetChunkOptions) (otf.Chunk, error) {
	logs, err := s.logs.GetChunk(ctx, id, opts)
	if err != nil {
		s.Error(err, "reading plan logs", "id", id, "offset", opts.Offset, "limit", opts.Limit)
		return otf.Chunk{}, err
	}

	return logs, nil
}

// PutChunk writes a chunk of logs for a terraform plan.
func (s PlanService) PutChunk(ctx context.Context, id string, chunk otf.Chunk) error {
	err := s.logs.PutChunk(ctx, id, chunk)
	if err != nil {
		s.Error(err, "writing plan logs", "id", id, "start", chunk.Start, "end", chunk.End)
		return err
	}

	s.V(2).Info("written plan logs", "id", id, "start", chunk.Start, "end", chunk.End)

	return nil
}

// Claim implements Job
func (s PlanService) Claim(ctx context.Context, planID string, opts otf.JobClaimOptions) (otf.Job, error) {
	run, err := s.db.UpdateStatus(ctx, otf.RunGetOptions{PlanID: &planID}, func(run *otf.Run) error {
		return run.Plan.Start(run)
	})
	if err != nil {
		s.Error(err, "starting plan", "plan_id", planID)
		return nil, err
	}

	s.V(0).Info("started plan", "id", run.ID())

	return run, nil
}

// Finish marks a plan as having finished.  An event is emitted to notify any
// subscribers of the new state.
func (s PlanService) Finish(ctx context.Context, planID string, opts otf.JobFinishOptions) (otf.Job, error) {
	var event *otf.Event

	run, err := s.db.UpdateStatus(ctx, otf.RunGetOptions{PlanID: &planID}, func(run *otf.Run) (err error) {
		event, err = run.Plan.Finish(opts)
		return err
	})
	if err != nil {
		s.Error(err, "finishing plan", "id", planID)
		return nil, err
	}

	s.V(0).Info("finished plan", "id", planID)

	s.Publish(*event)

	return run, nil
}
