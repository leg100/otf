package app

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/leg100/otf"
)

var _ otf.JobService = (*JobService)(nil)

type JobService struct {
	db otf.RunStore

	logs otf.ChunkStore

	otf.EventService

	cache otf.Cache

	logr.Logger
}

func NewJobService(db otf.RunStore, logs otf.ChunkStore, logger logr.Logger, es otf.EventService, cache otf.Cache) *JobService {
	return &JobService{
		db:           db,
		EventService: es,
		logs:         logs,
		cache:        cache,
		Logger:       logger,
	}
}

func (s JobService) Queued(ctx context.Context, id string) ([]*otf.Job, error) {
	run, err := s.db.Get(ctx, otf.RunGetOptions{JobID: &id})
	if err != nil {
		return nil, err
	}
	return run.Job, nil
}

// GetChunk reads a chunk of logs for a terraform plan.
func (s JobService) GetChunk(ctx context.Context, id string, opts otf.GetChunkOptions) (otf.Chunk, error) {
	logs, err := s.logs.GetChunk(ctx, id, opts)
	if err != nil {
		s.Error(err, "reading plan logs", "id", id, "offset", opts.Offset, "limit", opts.Limit)
		return otf.Chunk{}, err
	}

	return logs, nil
}

// PutChunk writes a chunk of logs for a terraform plan.
func (s JobService) PutChunk(ctx context.Context, id string, chunk otf.Chunk) error {
	err := s.logs.PutChunk(ctx, id, chunk)
	if err != nil {
		s.Error(err, "writing plan logs", "id", id, "start", chunk.Start, "end", chunk.End)
		return err
	}

	s.V(2).Info("written plan logs", "id", id, "start", chunk.Start, "end", chunk.End)

	return nil
}

// Claim implements Job
func (s JobService) Claim(ctx context.Context, planID string, opts otf.JobClaimOptions) (otf.Job, error) {
	run, err := s.db.UpdateStatus(ctx, otf.RunGetOptions{JobID: &planID}, func(run *otf.Run) error {
		return run.Job.Start(run)
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
func (s JobService) Finish(ctx context.Context, planID string, opts otf.JobFinishOptions) (otf.Job, error) {
	var event *otf.Event

	run, err := s.db.UpdateStatus(ctx, otf.RunGetOptions{JobID: &planID}, func(run *otf.Run) (err error) {
		event, err = run.Job.Finish(opts)
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
