/*
Package agent provides a daemon capable of running remote operations on behalf of a user.
*/
package agent

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/leg100/otf"
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

// NewAgent is the constructor for an Agent
func NewAgent(logger logr.Logger, app otf.Application, sub Watcher) (*Agent, error) {

	logger = logger.WithValues("component", "agent")

	spooler, err := NewSpooler(app, sub, logger)
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
func (a *Agent) Start(ctx context.Context) {
	// start spooler in background TODO: error not handled
	go a.Spooler.Start(ctx)

	a.Supervisor.Start(ctx)
}
