package agent

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/leg100/otf"
)

const (
	DefaultConcurrency = 5
)

// Supervisor supervises concurrently running workers.
type Supervisor struct {
	// concurrency is the max number of concurrent workers
	concurrency int

	logr.Logger

	otf.Executor

	Spooler

	*Terminator
}

// NewSupervisor is the constructor for Supervisor
func NewSupervisor(spooler Spooler, cvs otf.ConfigurationVersionService, svs otf.StateVersionService, rs otf.RunService, logger logr.Logger, concurrency int) *Supervisor {
	return &Supervisor{
		Spooler: spooler,
		Executor: otf.Executor{
			RunService:                  rs,
			StateVersionService:         svs,
			ConfigurationVersionService: cvs,
			Logger:                      logger,
			AgentID:                     DefaultID,
		},
		concurrency: concurrency,
		Terminator:  NewTerminator(),
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
		case run := <-s.GetCancelation():
			// TODO: support force cancelations too.
			s.Cancel(run.GetID(), false)
		case <-ctx.Done():
			return
		}
	}
}
