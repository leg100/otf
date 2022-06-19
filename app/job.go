package app

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/leg100/otf"
)

type jobService struct {
	db otf.DB
	rs otf.RunService

	otf.EventService
	otf.ChunkService
	logr.Logger
}

func newJobService(db otf.DB, logger logr.Logger, es otf.EventService, cs otf.ChunkService, rs otf.RunService) *jobService {
	return &jobService{
		db:           db,
		ChunkService: cs,
		rs:           rs,
		EventService: es,
		Logger:       logger,
	}
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
	run, err := s.db.UpdateStatus(ctx, otf.RunGetOptions{JobID: &jobID}, func(run *otf.Run) error {
		return run.Finish(s.rs, opts)
	})
	if err != nil {
		s.Error(err, "finishing job", "id", jobID)
		return nil, err
	}
	s.V(0).Info("finished job", "id", jobID)
	s.Publish(*event)
	return run, nil
}
