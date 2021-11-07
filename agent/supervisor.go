package agent

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-logr/logr"
	"github.com/leg100/otf"
)

const (
	DefaultConcurrency = 5
)

var (
	PluginCacheDir = filepath.Join(os.TempDir(), "plugin-cache")
)

// Supervisor supervises concurrently running workers.
type Supervisor struct {
	RunService                  otf.RunService
	ConfigurationVersionService otf.ConfigurationVersionService
	StateVersionService         otf.StateVersionService

	otf.JobSelector

	// concurrency is the max number of concurrent workers
	concurrency int

	logr.Logger

	AgentID string

	Spooler

	*Terminator

	environmentVariables []string
}

// NewSupervisor is the constructor for Supervisor
func NewSupervisor(spooler Spooler, cvs otf.ConfigurationVersionService, svs otf.StateVersionService, rs otf.RunService, ps otf.PlanService, as otf.ApplyService, logger logr.Logger, concurrency int) *Supervisor {
	s := &Supervisor{
		Spooler:             spooler,
		RunService:          rs,
		StateVersionService: svs,
		JobSelector: otf.JobSelector{
			PlanService:  ps,
			ApplyService: as,
		},
		ConfigurationVersionService: cvs,
		Logger:                      logger,
		AgentID:                     DefaultID,
		concurrency:                 concurrency,
		Terminator:                  NewTerminator(),
	}

	if err := os.MkdirAll(PluginCacheDir, 0755); err != nil {
		panic(fmt.Sprintf("cannot create plugin cache dir: %s: %s", PluginCacheDir, err.Error()))
	}
	s.environmentVariables = append(os.Environ(), fmt.Sprint("TF_PLUGIN_CACHE_DIR=", PluginCacheDir))

	return s
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
