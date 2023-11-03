// Package agent contains code related to agents
package agent

import (
	"context"
	"log/slog"
	"net"
	"time"

	"github.com/leg100/otf/internal"
)

type Status string

const (
	AgentIdle    Status = "idle"
	AgentBusy    Status = "busy"
	AgentExited  Status = "exited"
	AgentErrored Status = "errored"
)

type Agent struct {
	// Unique system-wide ID
	ID string
	// Optional name
	Name *string
	// Current status of agent
	Status Status
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

func (a *Agent) isActive() bool {
	switch a.Status {
	case AgentIdle, AgentBusy:
		return true
	default:
		return false
	}
}

type registerAgentOptions struct {
	Name        *string
	Concurrency int
	IPAddress   net.IP
	AgentPoolID *string
}

// registrar registers new agents.
type registrar struct {
	*service
}

func (f *registrar) register(ctx context.Context, opts registerAgentOptions) (*Agent, error) {
	return &Agent{
		ID:          internal.NewID("agent"),
		Name:        opts.Name,
		IPAddress:   opts.IPAddress,
		Concurrency: opts.Concurrency,
		AgentPoolID: opts.AgentPoolID,
		Server:      opts.AgentPoolID == nil,
		Status:      AgentIdle,
		LastPingAt:  internal.CurrentTimestamp(nil),
	}, nil
}
