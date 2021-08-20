/*
Package agent provides a daemon capable of running remote operations on behalf of a user.
*/
package agent

import (
	"context"
	"os"

	"github.com/go-logr/logr"
	"github.com/leg100/go-tfe"
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
	PlanService                 ots.PlanService
	RunService                  ots.RunService

	Processor

	*Spooler
}

// NewAgent is the constructor for an Agent
func NewAgent(logger logr.Logger, cvs ots.ConfigurationVersionService, svs ots.StateVersionService, ps ots.PlanService, rs ots.RunService, es ots.EventService) (*Agent, error) {
	spooler, err := NewSpooler(rs, es, logger)
	if err != nil {
		return nil, err
	}

	return &Agent{
		Logger:      logger.WithValues("component", "agent"),
		RunService:  rs,
		PlanService: ps,
		Processor: &processor{
			Logger:                      logger.WithValues("component", "agent"),
			ConfigurationVersionService: cvs,
			StateVersionService:         svs,
			PlanService:                 ps,
			RunService:                  rs,
			TerraformRunner:             &runner{},
		},
		Spooler: spooler,
	}, nil
}

// Start starts the agent daemon
func (a *Agent) Start(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case run := <-a.GetJob():
			a.handleJob(ctx, run)
		}
	}
}

func (a *Agent) handleJob(ctx context.Context, run *ots.Run) {
	a.Info("job received", "run", run.ID, "status", run.Status)

	path, err := os.MkdirTemp("", "ots-plan")
	if err != nil {
		// TODO: update run status with error
		a.Error(err, "unable to create temp path")
		return
	}

	switch run.Status {
	case tfe.RunPlanQueued:
		if err := a.Plan(ctx, run, path); err != nil {
			a.Error(err, "unable to process run")

			_, err := a.RunService.UpdatePlanStatus(run.ID, tfe.PlanErrored)
			if err != nil {
				a.Error(err, "unable to update plan status")
			}
		}
	case tfe.RunApplyQueued:
		if err := a.Apply(ctx, run, path); err != nil {
			a.Error(err, "unable to process run")

			_, err := a.RunService.UpdateApplyStatus(run.ID, tfe.ApplyErrored)
			if err != nil {
				a.Error(err, "unable to update apply status")
			}
		}
	default:
		a.Error(nil, "unexpected run status", "status", run.Status)
	}
}
