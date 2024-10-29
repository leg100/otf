package runner

import (
	"context"
	"log/slog"
	"net/netip"
	"time"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/rbac"
)

// runnerMeta is information about a runner.
type runnerMeta struct {
	// Unique system-wide ID
	ID string `jsonapi:"primary,agents"`
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
	// ID of agent's pool. Nil if server runner.
	AgentPoolID *string `jsonapi:"attribute" json:"agent-pool-id"`
	// Agent pool's organization
	AgentPoolOrganizationName *string

	isRemote bool
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
	// ID of agent's pool. If unset then the agent is assumed to be a server
	// agent (which does not belong to a pool).
	AgentPoolID *string `json:"-"`
	// CurrentJobs are those jobs the agent has discovered leftover from a
	// previous agent. Not currently used but may be made use of in later
	// versions.
	CurrentJobs []JobSpec `json:"current-jobs,omitempty"`
}

func register(opts registerOptions) (*runnerMeta, error) {
	m := &runnerMeta{
		ID:          internal.NewID("agent"),
		Name:        opts.Name,
		Version:     opts.Version,
		MaxJobs:     opts.Concurrency,
		AgentPoolID: opts.AgentPoolID,
	}
	if err := m.setStatus(RunnerIdle, true); err != nil {
		return nil, err
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

	return m, nil
}

func metadataFromContext(ctx context.Context) (*runnerMeta, error) {
	subject, err := internal.SubjectFromContext(ctx)
	if err != nil {
		return nil, err
	}
	meta, ok := subject.(*runnerMeta)
	if !ok {
		return nil, ErrUnauthorizedAgentRegistration
	}
	return meta, nil
}

func (m *runnerMeta) setStatus(status RunnerStatus, ping bool) error {
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

func (m *runnerMeta) LogValue() slog.Value {
	attrs := []slog.Attr{
		slog.String("id", m.ID),
		slog.Bool("remote", m.isRemote),
		slog.String("status", string(m.Status)),
		slog.String("ip_address", m.IPAddress.String()),
	}
	if m.AgentPoolID != nil {
		attrs = append(attrs, slog.String("pool_id", *m.AgentPoolID))
	}
	if m.Name != "" {
		attrs = append(attrs, slog.String("name", m.Name))
	}
	return slog.GroupValue(attrs...)
}

func (a *runnerMeta) CanAccessOrganization(action rbac.Action, name string) bool {
	// TODO: permit only those actions that an agent needs to carry out (get
	// agent jobs, etc).
	if a.isRemote {
		return *a.AgentPoolOrganizationName == name
	}
	return true
}

func (a *runnerMeta) CanAccessWorkspace(action rbac.Action, policy internal.WorkspacePolicy) bool {
	// only a server-based agent can authenticate as an Agent, and if that is
	// so, then it can carry out all workspace-based actions.
	//
	// TODO: permit only those actions that an agent needs to carry out (get
	// agent jobs, etc).
	if a.isRemote {
		return *a.AgentPoolOrganizationName == policy.Organization
	}
	return true
}
