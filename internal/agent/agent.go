// Package agent contains code related to agents
package agent

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"time"

	"github.com/leg100/otf/internal"
)

var (
	ErrInvalidAgentStateTransition   = errors.New("invalid agent state transition")
	ErrUnauthorizedAgentRegistration = errors.New("unauthorization agent registration")
)

type AgentStatus string

const (
	AgentIdle    AgentStatus = "idle"
	AgentBusy    AgentStatus = "busy"
	AgentExited  AgentStatus = "exited"
	AgentErrored AgentStatus = "errored"
	AgentUnknown AgentStatus = "unknown"
)

// Agent describes an agent. (The agent *process* is Daemon).
type Agent struct {
	// Unique system-wide ID
	ID string `jsonapi:"primary,agents"`
	// Optional name
	Name *string `jsonapi:"attribute" json:"name"`
	// Version of agent
	Version string `jsonapi:"attribute" json:"version"`
	// Current status of agent
	Status AgentStatus `jsonapi:"attribute" json:"status"`
	// Number of jobs it can handle at once
	Concurrency int `jsonapi:"attribute" json:"concurrency"`
	// Last time a ping was received from the agent
	LastPingAt time.Time `jsonapi:"attribute" json:"last-ping-at"`
	// Last time the status was updated
	LastStatusAt time.Time `jsonapi:"attribute" json:"last-status-at"`
	// IP address of agent
	IPAddress net.IP `jsonapi:"attribute" json:"ip-address"`
	// ID of agent' pool. If nil then the agent is assumed to be a server agent
	// (otfd).
	AgentPoolID *string `jsonapi:"attribute" json:"agent-pool-id"`
}

type registerAgentOptions struct {
	// Descriptive name. Optional.
	Name *string `json:"name"`
	// Version of agent.
	Version string `json:"version"`
	// Number of jobs the agent can handle at any one time.
	Concurrency int `json:"concurrency"`
	// IPAddress of agent. Optional. Not sent over the wire; instead the server
	// handler is responsible for determing client's IP address.
	IPAddress net.IP `json:"-"`
	// ID of agent's pool. If unset then the agent is assumed to be a server
	// agent (which does not belong to a pool).
	AgentPoolID *string `json:"-"`
	// CurrentJobs are those jobs the agent has discovered leftover from a
	// previous agent. Not currently used but may be made use of in later
	// versions.
	CurrentJobs []JobSpec `json:"current-jobs,omitempty"`
}

// registrar registers new agents.
type registrar struct {
	*service
}

func (f *registrar) register(ctx context.Context, opts registerAgentOptions) (*Agent, error) {
	agent := &Agent{
		ID:          internal.NewID("agent"),
		Name:        opts.Name,
		Version:     opts.Version,
		Concurrency: opts.Concurrency,
		AgentPoolID: opts.AgentPoolID,
		LastPingAt:  internal.CurrentTimestamp(nil),
	}
	if err := agent.setStatus(AgentIdle, true); err != nil {
		return nil, err
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

func (a *Agent) setStatus(status AgentStatus, ping bool) error {
	// the agent fsm is as follows:
	//
	// idle -> any
	// busy -> any
	// unknown -> any
	// errored (final state)
	// exited (final state)
	switch a.Status {
	case AgentErrored, AgentExited:
		return internal.ErrConflict
	}
	a.Status = status
	now := internal.CurrentTimestamp(nil)
	a.LastStatusAt = now
	// also update ping time if requested
	if ping {
		a.LastPingAt = now
	}
	return nil
}

// IsServer determines whether the agent is part of the server process (otfd) or
// a separate process (otf-agent).
func (a *Agent) IsServer() bool { return a.AgentPoolID == nil }

func (a *Agent) LogValue() slog.Value {
	attrs := []slog.Attr{
		slog.String("id", a.ID),
		slog.Bool("server", a.IsServer()),
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
