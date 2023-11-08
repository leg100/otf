package agent

import (
	"context"
	"time"
)

var (
	pingTimeout            = 60 * time.Second
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
	*service
	// frequency with which the manager will check agents.
	interval time.Duration
}

// Start the manager. Every interval the status of agents is checked,
// updating their status as necessary.
//
// Should be invoked in a go routine.
func (a *manager) Start(ctx context.Context) error {
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
		// update agent status from unknown to errored if it has still failed to
		// ping within a further 2 hours.
		if time.Since(agent.LastPingAt) > pingTimeout+(2*time.Hour) {
			// update agent status to errored.
			return a.updateAgentStatus(ctx, agent.ID, AgentErrored)
		}
	case AgentErrored, AgentExited:
		// purge agent from database once a further 3 hours has passed for
		// agents in a terminal state.
		if time.Since(agent.LastPingAt) > pingTimeout+(3*time.Hour) {
			// remove agent from db
			return a.deleteAgent(ctx, agent.ID)
		}
	}
	return nil
}
