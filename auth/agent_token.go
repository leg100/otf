package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/leg100/otf"
	"github.com/leg100/otf/rbac"
)

type (
	// AgentToken is an long-lived authentication token for an external agent.
	AgentToken struct {
		ID           string
		CreatedAt    time.Time
		Token        string
		Description  string
		Organization string
	}

	CreateAgentTokenOptions struct {
		Organization string `schema:"organization_name,required"`
		Description  string `schema:"description,required"`
	}

	// AgentTokenService provides access to agent tokens
	AgentTokenService interface {
		GetAgentToken(ctx context.Context, token string) (AgentToken, error)
	}
)

func NewAgentToken(opts CreateAgentTokenOptions) (*AgentToken, error) {
	if opts.Organization == "" {
		return nil, fmt.Errorf("organization name cannot be an empty string")
	}
	if opts.Description == "" {
		return nil, fmt.Errorf("description cannot be an empty string")
	}
	t, err := otf.GenerateAuthToken("agent")
	if err != nil {
		return nil, fmt.Errorf("generating token: %w", err)
	}
	token := AgentToken{
		ID:           otf.NewID("at"),
		CreatedAt:    otf.CurrentTimestamp(),
		Token:        t,
		Description:  opts.Description,
		Organization: opts.Organization,
	}
	return &token, nil
}

func (t *AgentToken) String() string      { return t.ID }
func (t *AgentToken) IsSiteAdmin() bool   { return true }
func (t *AgentToken) IsOwner(string) bool { return true }

func (*AgentToken) CanAccessSite(action rbac.Action) bool {
	// agent cannot carry out site-level actions
	return false
}

func (t *AgentToken) CanAccessOrganization(action rbac.Action, name string) bool {
	return t.Organization == name
}

func (t *AgentToken) CanAccessWorkspace(action rbac.Action, policy otf.WorkspacePolicy) bool {
	// agent can access anything within its organization
	return t.Organization == policy.Organization
}
