package agent

import (
	"bytes"
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

	RunService                  ots.RunService
	ConfigurationVersionService ots.ConfigurationVersionService
	StateVersionService         ots.StateVersionService

	Spooler

	// Overridable plan runner constructor for testing purposes
	planRunnerFn NewPlanRunnerFn

	// Overridable apply runner constructor for testing purposes
	applyRunnerFn NewApplyRunnerFn
}

func NewSupervisor(spooler Spooler, cvs ots.ConfigurationVersionService, svs ots.StateVersionService, rs ots.RunService, logger logr.Logger, concurrency int) *Supervisor {
	return &Supervisor{
		Spooler:                     spooler,
		RunService:                  rs,
		StateVersionService:         svs,
		ConfigurationVersionService: cvs,
		Logger:                      logger,
		concurrency:                 concurrency,
		planRunnerFn:                NewPlanRunner,
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

	// For logs
	out := new(bytes.Buffer)

	switch run.Status {
	case tfe.RunPlanQueued:
		runner := s.planRunnerFn(run, s.ConfigurationVersionService,
			s.StateVersionService, s.RunService, s.Logger)

		runErr := runner.Run(ctx, path, out)

		// Upload logs regardless of runner error
		if err := s.RunService.UploadPlanLogs(run.ID, out.Bytes()); err != nil {
			s.Error(err, "unable to upload plan logs", "run", run.ID)
		}

		if runErr != nil {
			_, err := s.RunService.UpdatePlanStatus(run.ID, tfe.PlanErrored)
			if err != nil {
				s.Error(err, "unable to update plan status", "run", run.ID)
			}
		}
	case tfe.RunApplyQueued:
		runner := s.applyRunnerFn(run, s.ConfigurationVersionService,
			s.StateVersionService, s.RunService, s.Logger)

		runErr := runner.Run(ctx, path, out)

		// Upload logs regardless of runner error
		if err := s.RunService.UploadApplyLogs(run.ID, out.Bytes()); err != nil {
			s.Error(err, "unable to upload apply logs", "run", run.ID)
		}

		if runErr != nil {
			_, err := s.RunService.UpdateApplyStatus(run.ID, tfe.ApplyErrored)
			if err != nil {
				s.Error(err, "unable to update apply status", "run", run.ID)
			}
		}
	default:
		s.Error(nil, "unexpected run status", "status", run.Status)
	}
}
