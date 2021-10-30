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

	ConfigurationVersionService otf.ConfigurationVersionService
	StateVersionService         otf.StateVersionService

	Spooler

	*Supervisor
}

// NewAgent is the constructor for an Agent
func NewAgent(logger logr.Logger, cvs otf.ConfigurationVersionService, svs otf.StateVersionService, rs otf.RunService, es otf.EventService) (*Agent, error) {
	logger = logger.WithValues("component", "agent")

	executor := otf.Executor{
		RunService:                  rs,
		StateVersionService:         svs,
		ConfigurationVersionService: cvs,
		Logger:                      logger,
		AgentID:                     DefaultID,
	}

	spooler, err := NewSpooler(rs, es, logger, &executor)
	if err != nil {
		return nil, err
	}

	return &Agent{
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
