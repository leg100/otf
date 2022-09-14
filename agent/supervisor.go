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

var PluginCacheDir = filepath.Join(os.TempDir(), "plugin-cache")

// Supervisor supervises concurrently running workers.
type Supervisor struct {
	otf.Application

	// concurrency is the max number of concurrent workers
	concurrency int

	logr.Logger

	AgentID string

	Spooler

	*Terminator
	// Downloader for workers to download tf cli on demand
	otf.Downloader
	environmentVariables []string
}

// NewSupervisor is the constructor for Supervisor
func NewSupervisor(spooler Spooler, app otf.Application, logger logr.Logger, concurrency int) *Supervisor {
	s := &Supervisor{
		Spooler:     spooler,
		Application: app,
		Logger:      logger,
		AgentID:     DefaultID,
		concurrency: concurrency,
		Terminator:  NewTerminator(),
		Downloader:  NewTerraformDownloader(),
	}

	// TODO: consider moving env var setup closer to where they are used i.e.
	// the Environment obj
	if err := os.MkdirAll(PluginCacheDir, 0o755); err != nil {
		panic(fmt.Sprintf("cannot create plugin cache dir: %s: %s", PluginCacheDir, err.Error()))
	}
	s.environmentVariables = append(os.Environ(), fmt.Sprint("TF_PLUGIN_CACHE_DIR=", PluginCacheDir))
	s.environmentVariables = append(s.environmentVariables, "TF_IN_AUTOMATION=true")
	s.environmentVariables = append(s.environmentVariables, "CHECKPOINT_DISABLE=true")

	return s
}

// Start starts the supervisor's workers.
func (s *Supervisor) Start(ctx context.Context) error {
	for i := 0; i < s.concurrency; i++ {
		w := &Worker{Supervisor: s}
		go w.Start(ctx)
	}

	for {
		select {
		case cancelation := <-s.GetCancelation():
			s.Cancel(cancelation.Run.ID(), cancelation.Forceful)
		case <-ctx.Done():
			return nil
		}
	}
}
