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
	db   *sql.DB
	logs otf.ChunkStore
	otf.EventService
	cache otf.Cache
	logr.Logger
}

func NewApplyService(db *sql.DB, logs otf.ChunkStore, logger logr.Logger, es otf.EventService, cache otf.Cache) (*ApplyService, error) {
	logs, err := inmem.NewChunkProxy(cache, logs)
	if err != nil {
		return nil, fmt.Errorf("constructing chunk proxy: %w", err)
	}
	return &ApplyService{
		db:           db,
		EventService: es,
		logs:         logs,
		cache:        cache,
		Logger:       logger,
	}, nil
}

func (s ApplyService) Get(ctx context.Context, id string) (*otf.Apply, error) {
	run, err := s.db.GetRun(ctx, otf.RunGetOptions{ApplyID: &id})
	if err != nil {
		return nil, err
	}
	return run.Apply, nil
}

// GetChunk reads a chunk of logs for a terraform apply.
func (s ApplyService) GetChunk(ctx context.Context, id string, opts otf.GetChunkOptions) (otf.Chunk, error) {
	logs, err := s.logs.GetChunk(ctx, id, opts)
	if err != nil {
		s.Error(err, "reading apply logs", "id", id, "offset", opts.Offset, "limit", opts.Limit)
		return otf.Chunk{}, err
	}
	s.V(2).Info("read apply logs", "id", id, "offset", opts.Offset, "limit", opts.Limit)
	return logs, nil
}

// PutChunk writes a chunk of logs for a terraform apply.
func (s ApplyService) PutChunk(ctx context.Context, applyID string, chunk otf.Chunk) error {
	err := s.logs.PutChunk(ctx, applyID, chunk)
	if err != nil {
		s.Error(err, "writing apply logs", "id", applyID, "start", chunk.Start, "end", chunk.End)
		return err
	}
	s.V(2).Info("written apply logs", "id", applyID, "start", chunk.Start, "end", chunk.End)
	if !chunk.End {
		return nil
	}
	// Last chunk uploaded. A summary of applied changes can now be parsed from
	// the full logs and set on the apply obj.
	chunk, err = s.logs.GetChunk(ctx, applyID, otf.GetChunkOptions{})
	if err != nil {
		s.Error(err, "reading apply logs", "id", applyID)
		return err
	}
	report, err := otf.ParseApplyOutput(string(chunk.Data))
	if err != nil {
		s.Error(err, "compiling report of applied changes", "id", applyID)
		return err
	}
	if err := s.db.CreateApplyReport(ctx, applyID, report); err != nil {
		s.Error(err, "saving applied changes report", "id", applyID)
		return err
	}
	s.V(1).Info("created applied changes report", "id", applyID,
		"adds", report.Additions,
		"changes", report.Changes,
		"destructions", report.Destructions)
	return nil
}

// Claim claims an apply job on behalf of an agent.
func (s ApplyService) Claim(ctx context.Context, applyID string, opts otf.JobClaimOptions) (otf.Job, error) {
	run, err := s.db.UpdateStatus(ctx, otf.RunGetOptions{ApplyID: &applyID}, func(run *otf.Run) error {
		return run.Apply.Start()
	})
	if err != nil {
		s.Error(err, "starting apply", "id", applyID)
		return nil, err
	}
	s.V(0).Info("started apply", "run_id", run.ID(), "id", applyID)
	return run, nil
}

// Finish marks a apply as having finished.  An event is emitted to notify any
// subscribers of the new state.
func (s ApplyService) Finish(ctx context.Context, applyID string, opts otf.JobFinishOptions) (otf.Job, error) {
	run, err := s.db.UpdateStatus(ctx, otf.RunGetOptions{ApplyID: &applyID}, func(run *otf.Run) (err error) {
		return run.Apply.Finish()
	})
	if err != nil {
		s.Error(err, "finishing apply", "id", applyID)
		return nil, err
	}
	s.V(0).Info("finished apply", "id", applyID)
	s.Publish(otf.Event{Payload: run, Type: otf.EventRunApplied})
	return run, nil
}
