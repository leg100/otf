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
	proxy otf.ChunkStore
	db    *sql.DB
	otf.EventService
	logr.Logger
}

func NewPlanService(db *sql.DB, logger logr.Logger, es otf.EventService, cache otf.Cache) (*PlanService, error) {
	proxy, err := inmem.NewChunkProxy(cache, db.PlanLogStore())
	if err != nil {
		return nil, fmt.Errorf("constructing chunk proxy: %w", err)
	}
	return &PlanService{
		proxy:        proxy,
		db:           db,
		EventService: es,
		Logger:       logger,
	}, nil
}

func (s PlanService) Get(ctx context.Context, planID string) (*otf.Plan, error) {
	run, err := s.db.GetRun(ctx, otf.RunGetOptions{PlanID: &planID})
	if err != nil {
		s.Error(err, "retrieving plan", "id", planID)
		return nil, err
	}
	s.V(2).Info("retrieved plan", "id", planID)
	return run.Plan(), nil
}

// GetChunk reads a chunk of logs for a plan.
func (s PlanService) GetChunk(ctx context.Context, planID string, opts otf.GetChunkOptions) (otf.Chunk, error) {
	logs, err := s.proxy.GetChunk(ctx, planID, opts)
	if err != nil {
		s.Error(err, "reading logs", "id", planID, "offset", opts.Offset, "limit", opts.Limit)
		return otf.Chunk{}, err
	}
	s.V(2).Info("read logs", "id", planID, "offset", opts.Offset, "limit", opts.Limit)
	return logs, nil
}

// PutChunk writes a chunk of logs for a plan.
func (s PlanService) PutChunk(ctx context.Context, planID string, chunk otf.Chunk) error {
	err := s.proxy.PutChunk(ctx, planID, chunk)
	if err != nil {
		s.Error(err, "writing logs", "id", planID, "start", chunk.Start, "end", chunk.End)
		return err
	}
	s.V(2).Info("written logs", "id", planID, "start", chunk.Start, "end", chunk.End)
	return nil
}

// Start plan phase.
func (s PlanService) Start(ctx context.Context, planID string, opts otf.PhaseStartOptions) (*otf.Run, error) {
	run, err := s.db.UpdateStatus(ctx, otf.RunGetOptions{PlanID: &planID}, func(run *otf.Run) error {
		return run.Start()
	})
	if err != nil {
		s.Error(err, "starting plan phase", "id", planID)
		return nil, err
	}
	s.V(0).Info("started plan phase", "id", run.ID())
	return run, nil
}

// Finish plan phase.
func (s PlanService) Finish(ctx context.Context, planID string, opts otf.PhaseFinishOptions) (*otf.Run, error) {
	var event *otf.Event
	run, err := s.db.UpdateStatus(ctx, otf.RunGetOptions{PlanID: &planID}, func(run *otf.Run) (err error) {
		event, err = run.Finish(opts)
		return err
	})
	if err != nil {
		s.Error(err, "finishing plan phase", "id", planID)
		return nil, err
	}
	s.V(0).Info("finished plan phase", "id", planID)
	s.Publish(*event)
	return run, nil
}
