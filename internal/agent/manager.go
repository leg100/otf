package agent

import (
	"context"
	"time"
)

var pingTimeout = 60 * time.Second

// manager updates the state of agents.
type manager struct {
	// service for retrieving agents and updating their state.
	service
	// frequency with which the manager will check agents.
	interval time.Duration
}

// Start the manager. Should be invoked in a go routine.
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
				newStatus, remove := a.check(agent)
				if remove {
					if err := a.deleteAgent(ctx, agent.ID); err != nil {
						return err
					}
				} else if newStatus != "" {
					if err := a.updateAgentStatus(ctx, agent.ID, newStatus); err != nil {
						return err
					}
				}
			}
		case <-ctx.Done():
			return nil
		}
	}
}

func (a *manager) check(agent *Agent) (updateTo AgentStatus, remove bool) {
	switch agent.Status {
	case AgentIdle, AgentBusy:
		if time.Since(agent.LastPingAt) > pingTimeout {
			// update agent status to unknown.
			return AgentUnknown, false
		}
	case AgentUnknown:
		// if pingTimeout + 2 hours has elapsed
		if time.Since(agent.LastPingAt) > pingTimeout+2*time.Hour {
			// update agent status to errored.
			return AgentErrored, false
		}
	case AgentErrored, AgentExited:
		// if pingTimeout + 3 hours has elapsed
		if time.Since(agent.LastPingAt.Add(3*time.Hour)) > pingTimeout {
			// remove agent
			return "", true
		}
	}
	return "", false
}
