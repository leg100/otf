package agent

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/leg100/ots"
)

const (
	DefaultConcurrency = 5
)

// Supervisor supervises concurrently running workers.
type Supervisor struct {
	// concurrency is the max number of concurrent workers
	concurrency int

	logr.Logger

	RunService                  ots.RunService
	ConfigurationVersionService ots.ConfigurationVersionService
	StateVersionService         ots.StateVersionService

	Spooler

	*Terminator
}

// NewSupervisor is the constructor for Supervisor
func NewSupervisor(spooler Spooler, cvs ots.ConfigurationVersionService, svs ots.StateVersionService, rs ots.RunService, logger logr.Logger, concurrency int) *Supervisor {
	return &Supervisor{
		Spooler:                     spooler,
		RunService:                  rs,
		StateVersionService:         svs,
		ConfigurationVersionService: cvs,
		Logger:                      logger,
		concurrency:                 concurrency,
		Terminator:                  NewTerminator(),
	}
}

// Start starts the supervisor's workers.
func (s *Supervisor) Start(ctx context.Context) {
	for i := 0; i < s.concurrency; i++ {
		w := &Worker{Supervisor: s}
		w.Start(ctx)
	}

	for {
		select {
		case job := <-s.GetCancelation():
			// TODO: support force cancelations too.
			s.Cancel(job.GetID(), false)
		case <-ctx.Done():
			return
		}
	}
}
