// Package agent contains code related to agents
package agent

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"time"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/rbac"
)

// An agent implements Subject (the server-based agent that is part of the otfd
// process authenticates as an agent, whereas the otf-agent process
// authenticates as an agent pool).
var _ internal.Subject = (*Agent)(nil)

type AgentStatus string

const (
	AgentIdle    AgentStatus = "idle"
	AgentBusy    AgentStatus = "busy"
	AgentExited  AgentStatus = "exited"
	AgentErrored AgentStatus = "errored"
	AgentUnknown AgentStatus = "unknown"
)

// Agent represents an agent. (The agent process itself is the Daemon).
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

func (a *Agent) String() string      { return a.ID }
func (a *Agent) IsSiteAdmin() bool   { return true }
func (a *Agent) IsOwner(string) bool { return true }

func (a *Agent) Organizations() []string {
	// an agent is not a member of organizations (although its agent pool is).
	return nil
}

func (*Agent) CanAccessSite(action rbac.Action) bool {
	// agent cannot carry out site-level actions
	return false
}

func (*Agent) CanAccessTeam(rbac.Action, string) bool {
	// agent cannot carry out team-level actions
	return false
}

func (a *Agent) CanAccessOrganization(action rbac.Action, name string) bool {
	// only a server-based agent can authenticate as an Agent, and if that is
	// so, then it can carry out all organization-based actions.
	//
	// TODO: permit only those actions that an agent needs to carry out (get
	// agent jobs, etc).
	return a.Server
}

func (a *Agent) CanAccessWorkspace(action rbac.Action, policy internal.WorkspacePolicy) bool {
	// only a server-based agent can authenticate as an Agent, and if that is
	// so, then it can carry out all workspace-based actions.
	//
	// TODO: permit only those actions that an agent needs to carry out (get
	// agent jobs, etc).
	return a.Server
}

// AgentFromContext retrieves an agent subject from a context
func AgentFromContext(ctx context.Context) (*Agent, error) {
	subj, err := internal.SubjectFromContext(ctx)
	if err != nil {
		return nil, err
	}
	agent, ok := subj.(*Agent)
	if !ok {
		return nil, fmt.Errorf("subject found in context but it is not an agent")
	}
	return agent, nil
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
