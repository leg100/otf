package agent

import (
	"context"
	"os"
	"time"

	"github.com/leg100/go-tfe"
	"github.com/leg100/ots"
	"github.com/rs/zerolog"
)

const (
	DefaultDataDir = "~/.ots-agent"
	DefaultID      = "agent-001"
)

// Agent processes jobs
type Agent struct {
	// ID uniquely identifies the agent.
	ID string

	zerolog.Logger

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

func NewAgent(logger *zerolog.Logger, cvs ots.ConfigurationVersionService, svs ots.StateVersionService, ps ots.PlanService, rs ots.RunService) *Agent {
	return &Agent{
		Logger:      logger.With().Str("component", "agent").Logger(),
		RunService:  rs,
		PlanService: ps,
		Processor: &processor{
			Logger:                      logger.With().Str("component", "agent").Logger(),
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
			break
		case <-time.After(time.Second):
		}

		runs, err := a.RunService.GetQueued(tfe.RunListOptions{})
		if err != nil {
			a.Error().Msgf("unable to poll daemon: %s", err.Error())
			continue
		}
		if len(runs.Items) == 0 {
			continue
		}

		a.Info().
			Str("run", runs.Items[0].ExternalID).
			Msg("job received")

		path, err := os.MkdirTemp("", "ots-plan")
		if err != nil {
			a.Error().Msgf("unable to create temp path: %s", err.Error())
			continue
		}

		if err := a.Process(ctx, runs.Items[0], path); err != nil {
			a.Error().Msgf("unable to process run: %s", err.Error())

			_, err := a.RunService.UpdatePlanStatus(runs.Items[0].ExternalID, tfe.PlanErrored)
			if err != nil {
				a.Error().Msgf("unable to update plan status: %s", err.Error())
			}
		}
	}
}
