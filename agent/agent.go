/*
Package agent provides a daemon capable of running remote operations on behalf of a user.
*/
package agent

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/leg100/otf"
	"golang.org/x/sync/errgroup"
)

const (
	DefaultDataDir = "~/.otf-agent"
	DefaultID      = "agent-001"
)

// Agent runs remote operations
type Agent struct {
	// ID uniquely identifies the agent.
	ID string

	// DataDir stores artefacts relating to runs, i.e. downloaded plugins,
	// modules (?), configuration versions, state, etc.
	DataDir string

	// ServerAddr is the address (<host>:<port>) of the OTF server to connect
	// to.
	ServerAddr string

	Spooler

	*Supervisor
}

// Config configures the agent.
type Config struct {
	// Organization if non-nil restricts the agent to processing runs belonging
	// to the specified organization.
	Organization *string
	// External indicates whether the agent is running as a separate process,
	// otf-agent, thus whether it handles runs for workspaces in remote mode
	// (external=false) or workspaces in agent mode (external=true).
	External bool
}

// NewAgent is the constructor for an Agent
func NewAgent(logger logr.Logger, app otf.Application, cfg Config) (*Agent, error) {
	spooler := NewSpooler(app, logger, cfg)

	supervisor := NewSupervisor(
		spooler,
		app,
		logger, DefaultConcurrency)

	return &Agent{
		Spooler:    spooler,
		Supervisor: supervisor,
	}, nil
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
