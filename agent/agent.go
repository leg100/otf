/*
Package agent provides a daemon capable of running remote operations on behalf of a user.
*/
package agent

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/leg100/ots"
)

const (
	DefaultDataDir = "~/.ots-agent"
	DefaultID      = "agent-001"
)

// Agent runs remote operations
type Agent struct {
	// ID uniquely identifies the agent.
	ID string

	logr.Logger

	// DataDir stores artefacts relating to runs, i.e. downloaded plugins,
	// modules (?), configuration versions, state, etc.
	DataDir string

	// ServerAddr is the address (<host>:<port>) of the OTS server to connect
	// to.
	ServerAddr string

	ConfigurationVersionService ots.ConfigurationVersionService
	StateVersionService         ots.StateVersionService

	Spooler

	*Supervisor
}

// NewAgent is the constructor for an Agent
func NewAgent(logger logr.Logger, cvs ots.ConfigurationVersionService, svs ots.StateVersionService, rs ots.RunService, es ots.EventService) (*Agent, error) {
	logger = logger.WithValues("component", "agent")

	spooler, err := NewSpooler(rs, es, logger)
	if err != nil {
		return nil, err
	}

	return &Agent{
		Logger:  logger,
		Spooler: spooler,
		Supervisor: NewSupervisor(
			spooler,
			cvs,
			svs,
			rs,
			logger, DefaultConcurrency),
	}, nil
}

// Start starts the agent daemon
func (a *Agent) Start(ctx context.Context) {
	// start spooler in background
	go a.Spooler.Start(ctx)

	a.Supervisor.Start(ctx)

}
