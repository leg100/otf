package agent

import (
	"context"
	"time"

	"github.com/leg100/otf/internal"
)

var (
	pingTimeout            = 30 * time.Second
	defaultManagerInterval = 10 * time.Second
)

// ManagerLockID guarantees only one manager on a cluster is running at any
// time.
const ManagerLockID int64 = 5577006791947779413

// manager manages the state of agents.
//
// Only one manager should be running on an OTF cluster at any one time.
type manager struct {
	// service for retrieving agents and updating their state.
	Service
	// frequency with which the manager will check agents.
	interval time.Duration
	// manager identifies itself as a subject when making service calls
	internal.Subject
}

func newManager(s Service) *manager {
	return &manager{
		Service:  s,
		interval: defaultManagerInterval,
	}
}

// Start the manager. Every interval the status of agents is checked,
// updating their status as necessary.
//
// Should be invoked in a go routine.
func (a *manager) Start(ctx context.Context) error {
	ctx = internal.AddSubjectToContext(ctx, a)

	ticker := time.NewTicker(a.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			agents, err := a.listAgents(ctx)
			if err != nil {
				return err
			}
			for _, agent := range agents {
				if err := a.update(ctx, agent); err != nil {
					return err
				}
			}
		case <-ctx.Done():
			return nil
		}
	}
}

func (a *manager) update(ctx context.Context, agent *Agent) error {
	switch agent.Status {
	case AgentIdle, AgentBusy:
		// update agent status to unknown if the agent has failed to ping within
		// the timeout.
		if time.Since(agent.LastPingAt) > pingTimeout {
			return a.updateAgentStatus(ctx, agent.ID, AgentUnknown)
		}
	case AgentUnknown:
		// update agent status from unknown to errored if a further period of 5
		// minutes has elapsed.
		if time.Since(agent.LastStatusAt) > 5*time.Minute {
			// update agent status to errored.
			return a.updateAgentStatus(ctx, agent.ID, AgentErrored)
		}
	case AgentErrored, AgentExited:
		// purge agent from database once a further 1 hour has elapsed for
		// agents in a terminal state.
		if time.Since(agent.LastStatusAt) > time.Hour {
			// remove agent from db
			return a.deleteAgent(ctx, agent.ID)
		}
	}
	return nil
}
