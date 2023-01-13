/*
Package agent provides a daemon capable of running remote operations on behalf of a user.
*/
package agent

import (
	"context"
	"errors"
	"fmt"
	"os/exec"

	"github.com/go-logr/logr"
	"github.com/leg100/otf"
	"github.com/spf13/pflag"
	"golang.org/x/sync/errgroup"
)

const (
	DefaultID = "agent-001"
)

// Agent processes runs.
type Agent struct {
	Spooler
	*Supervisor
}

// Config configures the agent.
type Config struct {
	Organization *string // only process runs belonging to org
	External     bool    // dedicated agent (true) or otfd (false)
	Concurrency  int     // number of workers
	Sandbox      bool    // isolate privileged ops within sandbox
	Debug        bool    // toggle debug mode
}

// NewAgent is the constructor for an Agent
func NewAgent(logger logr.Logger, app otf.Application, cfg Config) (*Agent, error) {
	if cfg.Sandbox {
		if _, err := exec.LookPath("bwrap"); errors.Is(err, exec.ErrNotFound) {
			return nil, fmt.Errorf("sandbox requires bubblewrap: %w", err)
		}
		logger.Info("sandbox mode enabled")
	}

	// Enable debug mode if log level = trace
	if logger.GetSink().Enabled(2) {
		logger.V(0).Info("enabled debug mode")
		cfg.Debug = true
	}

	spooler := NewSpooler(app, logger, cfg)
	return &Agent{
		Spooler:    spooler,
		Supervisor: NewSupervisor(spooler, app, logger, cfg),
	}, nil
}

func NewConfigFromFlags(flags *pflag.FlagSet) *Config {
	cfg := Config{}
	flags.BoolVar(&cfg.Sandbox, "sandbox", false, "Isolate terraform apply within sandbox for additional security")
	flags.IntVar(&cfg.Concurrency, "concurrency", DefaultConcurrency, "Number of runs that can be processed concurrently")
	return &cfg
}

// Start starts the agent daemon
func (a *Agent) Start(ctx context.Context) error {
	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		if err := a.Spooler.Start(ctx); err != nil {
			return fmt.Errorf("spooler terminated: %w", err)
		}
		return nil
	})

	g.Go(func() error {
		if err := a.Supervisor.Start(ctx); err != nil {
			return fmt.Errorf("supervisor terminated: %w", err)
		}
		return nil
	})

	return g.Wait()
}
