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

// NewAgentOptions are optional arguments to the NewAgent constructor
type NewAgentOptions struct {
	// Organization if non-nil restricts the agent to processing runs belonging
	// to the specified organization.
	Organization *string
	// Mode the agent is operating in: local or remote
	Mode AgentMode
}

type AgentMode string

const (
	InternalAgentMode AgentMode = "internal"
	ExternalAgentMode AgentMode = "external"
)

// NewAgent is the constructor for an Agent
func NewAgent(ctx context.Context, logger logr.Logger, app otf.Application, opts NewAgentOptions) (*Agent, error) {
	if opts.Mode != InternalAgentMode && opts.Mode != ExternalAgentMode {
		return nil, fmt.Errorf("invalid agent mode: %s", opts.Mode)
	}

	spooler := NewSpooler(ctx, app, logger, opts)

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
