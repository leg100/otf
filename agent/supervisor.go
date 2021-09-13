package agent

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/leg100/go-tfe"
	"github.com/leg100/ots"
)

const (
	DefaultConcurrency = 5
)

type newRunnerFn func(
	*ots.Run,
	ots.ConfigurationVersionService,
	ots.StateVersionService,
	ots.RunService,
	ots.RunLogger,
	logr.Logger) *ots.Runner

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
	planRunnerFn newRunnerFn

	// Overridable apply runner constructor for testing purposes
	applyRunnerFn newRunnerFn
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
		applyRunnerFn:               NewApplyRunner,
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
	s.Info("processing job", "run", run.ID, "status", run.Status)

	var runner *ots.Runner
	switch run.Status {
	case tfe.RunPlanQueued:
		runner = s.planRunnerFn(run, s.ConfigurationVersionService,
			s.StateVersionService, s.RunService, s.RunService.UploadPlanLogs,
			s.Logger)
	case tfe.RunApplyQueued:
		runner = s.applyRunnerFn(run, s.ConfigurationVersionService,
			s.StateVersionService, s.RunService, s.RunService.UploadApplyLogs,
			s.Logger)
	default:
		s.Error(nil, "unexpected run status", "status", run.Status)
	}

	if err := runner.Run(ctx); err != nil {
		s.Error(err, "unable to process run", "run", run.ID, "status", run.Status)

		_, err := s.RunService.UpdateStatus(run.ID, tfe.RunErrored)
		if err != nil {
			s.Error(err, "unable to update status", "run", run.ID)
		}
	}
}
