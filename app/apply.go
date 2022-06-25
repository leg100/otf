package app

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/leg100/otf"
	"github.com/leg100/otf/inmem"
	"github.com/leg100/otf/sql"
)

var _ otf.ApplyService = (*ApplyService)(nil)

type ApplyService struct {
	proxy otf.ChunkStore
	db    *sql.DB
	otf.EventService
	logr.Logger
}

func NewApplyService(db *sql.DB, logger logr.Logger, es otf.EventService, cache otf.Cache) (*ApplyService, error) {
	proxy, err := inmem.NewChunkProxy(cache, db.ApplyLogStore())
	if err != nil {
		return nil, fmt.Errorf("constructing chunk proxy: %w", err)
	}
	return &ApplyService{
		proxy:        proxy,
		db:           db,
		EventService: es,
		Logger:       logger,
	}, nil
}

func (s ApplyService) Get(ctx context.Context, id string) (*otf.Apply, error) {
	run, err := s.db.GetRun(ctx, otf.RunGetOptions{ApplyID: &id})
	if err != nil {
		return nil, err
	}
	return run.Apply(), nil
}

// GetChunk reads a chunk of logs for an apply.
func (s ApplyService) GetChunk(ctx context.Context, applyID string, opts otf.GetChunkOptions) (otf.Chunk, error) {
	logs, err := s.proxy.GetChunk(ctx, applyID, opts)
	if err != nil {
		s.Error(err, "reading logs", "id", applyID, "offset", opts.Offset, "limit", opts.Limit)
		return otf.Chunk{}, err
	}
	s.V(2).Info("read logs", "id", applyID, "offset", opts.Offset, "limit", opts.Limit)
	return logs, nil
}

// PutChunk writes a chunk of logs for an apply.
func (s ApplyService) PutChunk(ctx context.Context, applyID string, chunk otf.Chunk) error {
	err := s.proxy.PutChunk(ctx, applyID, chunk)
	if err != nil {
		s.Error(err, "writing logs", "id", applyID, "start", chunk.Start, "end", chunk.End)
		return err
	}
	s.V(2).Info("written logs", "id", applyID, "start", chunk.Start, "end", chunk.End)
	return nil
}

// Start apply phase.
func (s ApplyService) Start(ctx context.Context, applyID string, opts otf.PhaseStartOptions) (*otf.Run, error) {
	run, err := s.db.UpdateStatus(ctx, otf.RunGetOptions{ApplyID: &applyID}, func(run *otf.Run) error {
		return run.Start()
	})
	if err != nil {
		s.Error(err, "starting apply phase", "id", applyID)
		return nil, err
	}
	s.V(0).Info("started apply phase", "id", applyID)
	return run, nil
}

// Finish apply phase.
func (s ApplyService) Finish(ctx context.Context, applyID string, opts otf.PhaseFinishOptions) (*otf.Run, error) {
	run, err := s.db.UpdateStatus(ctx, otf.RunGetOptions{ApplyID: &applyID}, func(run *otf.Run) (err error) {
		return run.Finish(opts)
	})
	if err != nil {
		s.Error(err, "finishing apply phase", "id", applyID)
		return nil, err
	}
	s.V(0).Info("finished apply phase", "id", applyID)

	if err := s.CreateReport(ctx, applyID); err != nil {
		return nil, err
	}
	s.Publish(otf.Event{Type: otf.EventRunStatusUpdate, Payload: run})
	return run, nil
}

// CreateReport creates a report of applied changes.
func (s ApplyService) CreateReport(ctx context.Context, applyID string) error {
	chunk, err := s.GetChunk(ctx, applyID, otf.GetChunkOptions{})
	if err != nil {
		return err
	}
	report, err := otf.ParseApplyOutput(string(chunk.Data))
	if err != nil {
		return fmt.Errorf("compiling report of applied changes: %w", err)
	}
	if err := s.db.CreateApplyReport(ctx, applyID, report); err != nil {
		return fmt.Errorf("saving applied changes report: %w", err)
	}
	s.V(0).Info("compiled apply report",
		"id", applyID,
		"adds", report.Additions,
		"changes", report.Changes,
		"destructions", report.Destructions)

	return nil
}
