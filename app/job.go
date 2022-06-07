package app

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/leg100/otf"
)

var _ otf.JobService = (*JobService)(nil)

type JobService struct {
	db    otf.DB
	cache otf.Cache
	otf.EventService
	logr.Logger
}

func NewJobService(db otf.DB, logger logr.Logger, es otf.EventService, cache otf.Cache) *JobService {
	return &JobService{
		db:           db,
		EventService: es,
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

// GetChunk reads a chunk of logs for a job.
func (s JobService) GetChunk(ctx context.Context, jobID string, opts otf.GetChunkOptions) (otf.Chunk, error) {
	chunk, err := s.db.GetChunk(ctx, jobID, opts)
	if err != nil {
		s.Error(err, "reading logs", "id", jobID, "offset", opts.Offset, "limit", opts.Limit)
		return otf.Chunk{}, err
	}
	s.V(2).Info("read logs", "id", jobID, "start", chunk.Start, "end", chunk.End)

	return chunk, nil
}

// PutChunk writes a chunk of logs for a job.
func (s JobService) PutChunk(ctx context.Context, jobID string, chunk otf.Chunk) error {
	err := s.db.PutChunk(ctx, jobID, chunk)
	if err != nil {
		s.Error(err, "writing logs", "id", jobID, "start", chunk.Start, "end", chunk.End)
		return err
	}
	s.V(2).Info("written logs", "id", jobID, "start", chunk.Start, "end", chunk.End)

	return nil
}

// Claim implements Job
func (s JobService) Claim(ctx context.Context, jobID string, opts otf.JobClaimOptions) (*otf.Job, error) {
	job, err := s.db.UpdateJobStatus(ctx, jobID, otf.JobClaimed)
	if err != nil {
		s.Error(err, "claiming job", "id", jobID)
		return nil, err
	}
	s.V(0).Info("claimed job", "id", jobID)
	return job, nil
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
