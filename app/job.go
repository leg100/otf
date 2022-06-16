package app

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/leg100/otf"
	"github.com/leg100/otf/inmem"
)

var _ otf.ChunkService = (*jobService)(nil)

type jobService struct {
	proxy otf.ChunkStore
	db    otf.DB
	otf.EventService
	logr.Logger
}

func newJobService(db otf.DB, logger logr.Logger, es otf.EventService, cache otf.Cache) (*jobService, error) {
	proxy, err := inmem.NewChunkProxy(cache, db)
	if err != nil {
		return nil, fmt.Errorf("constructing chunk proxy: %w", err)
	}
	return &jobService{
		db:           db,
		proxy:        proxy,
		EventService: es,
		Logger:       logger,
	}, nil
}

// GetChunk reads a chunk of logs for a job.
func (s jobService) GetChunk(ctx context.Context, jobID string, opts otf.GetChunkOptions) (otf.Chunk, error) {
	logs, err := s.proxy.GetChunk(ctx, jobID, opts)
	if err != nil {
		s.Error(err, "reading logs", "id", jobID, "offset", opts.Offset, "limit", opts.Limit)
		return otf.Chunk{}, err
	}
	return logs, nil
}

// PutChunk writes a chunk of logs for a job.
func (s jobService) PutChunk(ctx context.Context, jobID string, chunk otf.Chunk) error {
	err := s.proxy.PutChunk(ctx, jobID, chunk)
	if err != nil {
		s.Error(err, "writing logs", "id", jobID, "start", chunk.Start, "end", chunk.End)
		return err
	}
	s.V(2).Info("written logs", "id", jobID, "start", chunk.Start, "end", chunk.End, "content", string(chunk.Data))
	return nil
}

// Claim a job.
func (s jobService) Claim(ctx context.Context, jobID string, opts otf.JobClaimOptions) (otf.Job, error) {
	run, err := s.db.UpdateStatus(ctx, otf.RunGetOptions{JobID: &jobID}, func(run *otf.Run) error {
		return run.Start()
	})
	if err != nil {
		s.Error(err, "starting job", "id", jobID)
		return nil, err
	}
	s.V(0).Info("started job", "id", run.ID())
	return run, nil
}

// Finish a job.
func (s jobService) Finish(ctx context.Context, jobID string, opts otf.JobFinishOptions) (otf.Job, error) {
	var event *otf.Event
	run, err := s.db.UpdateStatus(ctx, otf.RunGetOptions{JobID: &jobID}, func(run *otf.Run) (err error) {
		event, err = run.Finish(opts)
		return err
	})
	if err != nil {
		s.Error(err, "finishing job", "id", jobID)
		return nil, err
	}
	s.V(0).Info("finished job", "id", jobID)
	s.Publish(*event)
	return run, nil
}
