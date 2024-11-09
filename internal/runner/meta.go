package runner

import (
	"context"
	"errors"
	"log/slog"
	"net/netip"
	"time"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/authz"
	"github.com/leg100/otf/internal/rbac"
	"github.com/leg100/otf/internal/resource"
)

// RunnerMeta is information about a runner.
type RunnerMeta struct {
	ID resource.ID `jsonapi:"primary,runners"`
	// Optional name
	Name string `jsonapi:"attribute" json:"name"`
	// Version of runner
	Version string `jsonapi:"attribute" json:"version"`
	// Current status of runner
	Status RunnerStatus `jsonapi:"attribute" json:"status"`
	// Max number of jobs runner can execute
	MaxJobs int `jsonapi:"attribute" json:"max_jobs"`
	// Current number of jobs allocated to runner.
	CurrentJobs int `jsonapi:"attribute" json:"current_jobs"`
	// Last time a ping was received from the runner.
	LastPingAt time.Time `jsonapi:"attribute" json:"last-ping-at"`
	// Last time the status was updated
	LastStatusAt time.Time `jsonapi:"attribute" json:"last-status-at"`
	// IP address of runner.
	IPAddress netip.Addr `jsonapi:"attribute" json:"ip-address"`
	// Info about the runner's agent pool. Non-nil if agent runner; nil if server
	// runner.
	AgentPool *RunnerMetaAgentPool `jsonapi:"attribute" json:"agent-pool"`
}

type RunnerMetaAgentPool struct {
	// ID of agent's pool.
	ID resource.ID `json:"id"`
	// Name of agent's pool
	Name string `json:"name"`
	// Agent pool's organization.
	OrganizationName string `json:"organization-name"`
}

type registerOptions struct {
	// Descriptive name. Optional.
	Name string `json:"name"`
	// Version of agent.
	Version string `json:"version"`
	// Number of jobs the agent can handle at any one time.
	Concurrency int `json:"concurrency"`
	// IPAddress of agent. Optional. Not sent over the wire; instead the server
	// handler is responsible for determining client's IP address.
	IPAddress *netip.Addr `json:"-"`
	// ID of agent's pool. Only set if runner is an agent.
	AgentPoolID *string `json:"-"`
	// CurrentJobs are those jobs the agent has discovered leftover from a
	// previous agent. Not currently used but may be made use of in later
	// versions.
	CurrentJobs []JobSpec `json:"current-jobs,omitempty"`
}

func (m *RunnerMeta) register(opts registerOptions) error {
	if m.ID != resource.EmptyID {
		return errors.New("runner has already registered")
	}
	m.ID = resource.NewID(resource.RunnerKind)
	m.Name = opts.Name
	m.Version = opts.Version
	m.MaxJobs = opts.Concurrency

	if err := m.setStatus(RunnerIdle, true); err != nil {
		return err
	}
	if opts.IPAddress != nil {
		m.IPAddress = *opts.IPAddress
	} else {
		// IP address not provided: try to get local IP address used for
		// outbound comms, and if that fails, use localhost
		ip, err := internal.GetOutboundIP()
		if err != nil {
			ip = netip.IPv6Loopback()
		}
		m.IPAddress = ip
	}

	return nil
}

func (m *RunnerMeta) setStatus(status RunnerStatus, ping bool) error {
	// the agent fsm is as follows:
	//
	// idle -> any
	// busy -> any
	// unknown -> any
	// errored (final state)
	// exited (final state)
	switch m.Status {
	case RunnerErrored, RunnerExited:
		return internal.ErrConflict
	}
	m.Status = status
	now := internal.CurrentTimestamp(nil)
	m.LastStatusAt = now
	// also update ping time if requested
	if ping {
		m.LastPingAt = now
	}
	return nil
}

func (m *RunnerMeta) IsAgent() bool {
	return m.AgentPool != nil
}

func (m *RunnerMeta) LogValue() slog.Value {
	attrs := []slog.Attr{
		slog.String("id", m.ID.String()),
		slog.Bool("agent", m.IsAgent()),
		slog.String("status", string(m.Status)),
		slog.String("ip_address", m.IPAddress.String()),
	}
	if m.AgentPool != nil {
		attrs = append(attrs, slog.Any("pool", m.AgentPool))
	}
	if m.Name != "" {
		attrs = append(attrs, slog.String("name", m.Name))
	}
	return slog.GroupValue(attrs...)
}

func (m *RunnerMetaAgentPool) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("id", m.ID.String()),
		slog.String("name", m.Name),
		slog.String("organization", m.OrganizationName),
	)
}

func (m *RunnerMeta) IsSiteAdmin() bool   { return true }
func (m *RunnerMeta) IsOwner(string) bool { return true }

func (m *RunnerMeta) Organizations() []string { return nil }
func (m *RunnerMeta) String() string          { return m.ID.String() }

func (*RunnerMeta) CanAccessSite(action rbac.Action) bool {
	return false
}

func (*RunnerMeta) CanAccessTeam(rbac.Action, resource.ID) bool {
	return false
}

func (m *RunnerMeta) CanAccessOrganization(action rbac.Action, name string) bool {
	// TODO: permit only those actions that an agent needs to carry out (get
	// agent jobs, etc).
	if m.AgentPool != nil {
		return m.AgentPool.OrganizationName == name
	}
	return true
}

func (m *RunnerMeta) CanAccessWorkspace(action rbac.Action, policy authz.WorkspacePolicy) bool {
	// only a server-based agent can authenticate as an Agent, and if that is
	// so, then it can carry out all workspace-based actions.
	//
	// TODO: permit only those actions that an agent needs to carry out (get
	// agent jobs, etc).
	if m.AgentPool != nil {
		return m.AgentPool.OrganizationName == policy.Organization
	}
	return true
}

func runnerFromContext(ctx context.Context) (*RunnerMeta, error) {
	subject, err := authz.SubjectFromContext(ctx)
	if err != nil {
		return nil, err
	}
	meta, ok := subject.(*RunnerMeta)
	if !ok {
		return nil, internal.ErrAccessNotPermitted
	}
	return meta, nil
}

func authorizeRunner(ctx context.Context, id resource.ID) error {
	runner, err := runnerFromContext(ctx)
	if err != nil {
		return err
	}
	if id != runner.ID {
		return internal.ErrAccessNotPermitted
	}
	return nil
}
