/*
Package agent provides a daemon capable of running remote operations on behalf of a user.
*/
package agent

import (
	"context"

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
}

// NewAgent is the constructor for an Agent
func NewAgent(ctx context.Context, logger logr.Logger, app otf.Application, opts NewAgentOptions) (*Agent, error) {
	spooler, err := NewSpooler(ctx, app, app, logger, opts)
	if err != nil {
		return nil, err
	}

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
		return a.Spooler.Start(ctx)
	})

	g.Go(func() error {
		return a.Supervisor.Start(ctx)
	})

	return g.Wait()
}
