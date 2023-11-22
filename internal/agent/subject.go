package agent

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/rbac"
)

type (
	// unregisteredServer is a server agent that is yet to register
	unregisteredServerAgent struct {
		internal.Subject
	}

	// unregisteredPoolAgent is a pool agent that is yet to register
	unregisteredPoolAgent struct {
		pool         *Pool
		agentTokenID string

		internal.Subject
	}

	// serverAgent is a registered server agent (otfd) for the purposes of
	// authorization and auditing.
	serverAgent struct {
		*Agent
	}

	// poolAgent is a registered pool agent (otf-agent) for the purposes of
	// authorization and auditing.
	poolAgent struct {
		*unregisteredPoolAgent
		agent *Agent
	}
)

// poolAgentFromContext retrieves an pool agent subject from a context
func poolAgentFromContext(ctx context.Context) (*poolAgent, error) {
	subj, err := internal.SubjectFromContext(ctx)
	if err != nil {
		return nil, err
	}
	agent, ok := subj.(*poolAgent)
	if !ok {
		return nil, fmt.Errorf("subject found in context but it is not a pool agent")
	}
	return agent, nil
}

func (a *poolAgent) LogValue() slog.Value {
	attrs := append(
		a.agent.LogValue().Group(),
		slog.String("agent_token_id", a.agentTokenID),
	)
	return slog.GroupValue(attrs...)
}

func (a *poolAgent) String() string { return a.agent.ID }

func (a *poolAgent) IsSiteAdmin() bool   { return true }
func (a *poolAgent) IsOwner(string) bool { return true }

func (a *poolAgent) Organizations() []string { return nil }

func (*poolAgent) CanAccessSite(action rbac.Action) bool {
	return false
}

func (*poolAgent) CanAccessTeam(rbac.Action, string) bool {
	return false
}

func (a *poolAgent) CanAccessOrganization(action rbac.Action, name string) bool {
	return a.pool.Organization == name
}

func (a *poolAgent) CanAccessWorkspace(action rbac.Action, policy internal.WorkspacePolicy) bool {
	return a.pool.Organization == policy.Organization
}

func (a *serverAgent) String() string      { return a.ID }
func (a *serverAgent) IsSiteAdmin() bool   { return true }
func (a *serverAgent) IsOwner(string) bool { return true }

func (a *serverAgent) Organizations() []string {
	// an agent is not a member of organizations (although its agent pool is).
	return nil
}

func (*serverAgent) CanAccessSite(action rbac.Action) bool {
	// agent cannot carry out site-level actions
	return false
}

func (*serverAgent) CanAccessTeam(rbac.Action, string) bool {
	// agent cannot carry out team-level actions
	return false
}

func (a *serverAgent) CanAccessOrganization(action rbac.Action, name string) bool {
	// only a server-based agent can authenticate as an Agent, and if that is
	// so, then it can carry out all organization-based actions.
	//
	// TODO: permit only those actions that an agent needs to carry out (get
	// agent jobs, etc).
	return a.IsServer()
}

func (a *serverAgent) CanAccessWorkspace(action rbac.Action, policy internal.WorkspacePolicy) bool {
	// only a server-based agent can authenticate as an Agent, and if that is
	// so, then it can carry out all workspace-based actions.
	//
	// TODO: permit only those actions that an agent needs to carry out (get
	// agent jobs, etc).
	return a.IsServer()
}
