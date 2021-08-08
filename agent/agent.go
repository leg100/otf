/*
Package agent provides a daemon capable of running remote operations on behalf of a user.
*/
package agent

import (
	"context"
	"os"
	"time"

	"github.com/go-logr/logr"
	"github.com/leg100/go-tfe"
	"github.com/leg100/ots"
)

const (
	DefaultDataDir = "~/.ots-agent"
	DefaultID      = "agent-001"
)

// Agent processes jobs
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
}

func NewAgent(logger logr.Logger, cvs ots.ConfigurationVersionService, svs ots.StateVersionService, ps ots.PlanService, rs ots.RunService) *Agent {
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
	}
}

// Poller polls the daemon for queued runs and launches jobs accordingly.
func (a *Agent) Poller(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(time.Second):
		}

		runs, err := a.RunService.GetQueued(tfe.RunListOptions{})
		if err != nil {
			a.Error(err, "unable to poll daemon")
			continue
		}
		if len(runs.Items) == 0 {
			continue
		}

		a.Info("job received", "run", runs.Items[0].ID, "status", runs.Items[0].Status)

		path, err := os.MkdirTemp("", "ots-plan")
		if err != nil {
			a.Error(err, "unable to create temp path")
			continue
		}

		switch runs.Items[0].Status {
		case tfe.RunPlanQueued:
			if err := a.Plan(ctx, runs.Items[0], path); err != nil {
				a.Error(err, "unable to process run")

				_, err := a.RunService.UpdatePlanStatus(runs.Items[0].ID, tfe.PlanErrored)
				if err != nil {
					a.Error(err, "unable to update plan status")
				}
			}
		case tfe.RunApplyQueued:
			if err := a.Apply(ctx, runs.Items[0], path); err != nil {
				a.Error(err, "unable to process run")

				_, err := a.RunService.UpdateApplyStatus(runs.Items[0].ID, tfe.ApplyErrored)
				if err != nil {
					a.Error(err, "unable to update apply status")
				}
			}
		default:
			a.Error(nil, "unexpected run status", "status", runs.Items[0].Status)
		}
	}
}
