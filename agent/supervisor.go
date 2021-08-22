package agent

import (
	"context"
	"os"

	"github.com/go-logr/logr"
	"github.com/leg100/go-tfe"
	"github.com/leg100/ots"
)

const (
	DefaultConcurrency = 5
)

// Supervisor supervises jobs
type Supervisor struct {
	// concurrency is the max number of concurrent jobs
	concurrency int

	logr.Logger

	RunService ots.RunService

	Processor

	Spooler
}

func NewSupervisor(spooler Spooler, processor Processor, rs ots.RunService, logger logr.Logger, concurrency int) *Supervisor {
	return &Supervisor{
		Spooler:     spooler,
		Processor:   processor,
		RunService:  rs,
		Logger:      logger,
		concurrency: concurrency,
	}
}

// Start starts the supervisor daemon and workers
func (s *Supervisor) Start(ctx context.Context) {
	s.startWorkers(ctx)

	<-ctx.Done()
}

func (s *Supervisor) startWorkers(ctx context.Context) {
	for i := 0; i < s.concurrency; i++ {
		go func() {
			for run := range s.GetJob() {
				s.handleJob(ctx, run)
			}
		}()
	}
}

func (s *Supervisor) handleJob(ctx context.Context, run *ots.Run) {
	path, err := os.MkdirTemp("", "ots-plan")
	if err != nil {
		// TODO: update run status with error
		s.Error(err, "unable to create temp path")
		return
	}

	s.Info("processing job", "run", run.ID, "status", run.Status, "dir", path)

	switch run.Status {
	case tfe.RunPlanQueued:
		if err := s.Plan(ctx, run, path); err != nil {
			s.Error(err, "unable to process run", "run", run.ID)

			_, err := s.RunService.UpdatePlanStatus(run.ID, tfe.PlanErrored)
			if err != nil {
				s.Error(err, "unable to update plan status", "run", run.ID)
			}
		}
	case tfe.RunApplyQueued:
		if err := s.Apply(ctx, run, path); err != nil {
			s.Error(err, "unable to process run", "run", run.ID)

			_, err := s.RunService.UpdateApplyStatus(run.ID, tfe.ApplyErrored)
			if err != nil {
				s.Error(err, "unable to update apply status", "run", run.ID)
			}
		}
	default:
		s.Error(nil, "unexpected run status", "status", run.Status)
	}
}
