// Package agent contains code related to agents
package agent

import (
	"context"
	"log/slog"
	"net"
	"time"

	"github.com/leg100/otf/internal"
)

type AgentStatus string

const (
	AgentIdle    AgentStatus = "idle"
	AgentBusy    AgentStatus = "busy"
	AgentExited  AgentStatus = "exited"
	AgentErrored AgentStatus = "errored"
	AgentUnknown AgentStatus = "unknown"
)

// Agent represents the state of an agent.
type Agent struct {
	// Unique system-wide ID
	ID string
	// Optional name
	Name *string
	// Current status of agent
	Status AgentStatus
	// Number of jobs it can handle at once
	Concurrency int
	// Whether it is built into otfd (true) or is a separate otf-agent process
	// (false)
	Server bool
	// Last time a ping was received from the agent
	LastPingAt time.Time
	// IP address of agent
	IPAddress net.IP
	// ID of agent' pool. Only set if Server is false.
	AgentPoolID *string
}

func (a *Agent) LogValue() slog.Value {
	attrs := []slog.Attr{
		slog.String("id", a.ID),
		slog.Bool("server", a.Server),
		slog.String("status", string(a.Status)),
		slog.String("ip_address", a.IPAddress.String()),
	}
	if a.AgentPoolID != nil {
		attrs = append(attrs, slog.String("pool_id", *a.AgentPoolID))
	}
	if a.Name != nil {
		attrs = append(attrs, slog.String("name", *a.Name))
	}
	return slog.GroupValue(attrs...)
}

type registerAgentOptions struct {
	// Descriptive name. Optional.
	Name *string
	// Number of jobs the agent can handle at any one time.
	Concurrency int
	// IPAddress of agent. Optional.
	IPAddress net.IP `json:"-"`
	// ID of agent's pool. If unset then the agent is assumed to be a server
	// agent (which does not belong to a pool).
	AgentPoolID *string
	// CurrentJobs are those jobs the agent has discovered leftover from a
	// previous agent. Not currently used but may be made use of in later
	// versions.
	CurrentJobs []JobSpec
}

// registrar registers new agents.
type registrar struct {
	*service
}

func (f *registrar) register(ctx context.Context, opts registerAgentOptions) (*Agent, error) {
	agent := &Agent{
		ID:          internal.NewID("agent"),
		Name:        opts.Name,
		Concurrency: opts.Concurrency,
		AgentPoolID: opts.AgentPoolID,
		Server:      opts.AgentPoolID == nil,
		Status:      AgentIdle,
		LastPingAt:  internal.CurrentTimestamp(nil),
	}
	if opts.IPAddress != nil {
		agent.IPAddress = opts.IPAddress
	} else {
		// IP address not provided: try to get local IP address used for
		// outbound comms, and if that fails, use 127.0.0.1
		ip, err := internal.GetOutboundIP()
		if err != nil {
			ip = net.IPv4(127, 0, 0, 1)
		}
		agent.IPAddress = ip
	}

	return agent, nil
}
