package auth

import (
	"fmt"
	"time"

	"github.com/leg100/otf"
	"github.com/leg100/otf/rbac"
)

// AgentToken is an long-lived authentication token for an external agent.
type AgentToken struct {
	id           string
	createdAt    time.Time
	token        string
	description  string
	organization string
}

func (t *AgentToken) ID() string           { return t.id }
func (t *AgentToken) String() string       { return t.id }
func (t *AgentToken) Token() string        { return t.token }
func (t *AgentToken) CreatedAt() time.Time { return t.createdAt }
func (t *AgentToken) Description() string  { return t.description }
func (t *AgentToken) Organization() string { return t.organization }

func (*AgentToken) CanAccessSite(action rbac.Action) bool {
	// agent cannot carry out site-level actions
	return false
}

func (t *AgentToken) CanAccessOrganization(action rbac.Action, name string) bool {
	return t.organization == name
}

func (t *AgentToken) CanAccessWorkspace(action rbac.Action, policy *otf.WorkspacePolicy) bool {
	// agent can access anything within its organization
	return t.organization == policy.Organization
}

func newAgentToken(opts otf.CreateAgentTokenOptions) (*AgentToken, error) {
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
		id:           otf.NewID("at"),
		createdAt:    otf.CurrentTimestamp(),
		token:        t,
		description:  opts.Description,
		organization: opts.Organization,
	}
	return &token, nil
}
