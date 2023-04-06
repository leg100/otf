package auth

import (
	"fmt"
	"time"

	"github.com/leg100/otf"
	"github.com/leg100/otf/rbac"
	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jwt"
)

type (
	// AgentToken represents the authentication token for an external agent.
	// NOTE: the cryptographic token itself is not retained.
	AgentToken struct {
		ID           string
		CreatedAt    time.Time
		Description  string
		Organization string
	}

	CreateAgentTokenOptions struct {
		Organization string `schema:"organization_name,required"`
		Description  string `schema:"description,required"`
	}

	NewAgentTokenOptions struct {
		CreateAgentTokenOptions
		key jwk.Key // key for signing new token
	}
)

// NewAgentToken constructs a token for an external agent, returning both the
// representation of the token, and the cryptographic token itself.
//
// TODO(@leg100): Unit test this.
func NewAgentToken(opts NewAgentTokenOptions) (*AgentToken, []byte, error) {
	if opts.Organization == "" {
		return nil, nil, fmt.Errorf("organization name cannot be an empty string")
	}
	if opts.Description == "" {
		return nil, nil, fmt.Errorf("description cannot be an empty string")
	}
	at := AgentToken{
		ID:           otf.NewID("at"),
		CreatedAt:    otf.CurrentTimestamp(),
		Description:  opts.Description,
		Organization: opts.Organization,
	}
	token, err := jwt.NewBuilder().
		Subject(at.ID).
		Claim("kind", agentTokenKind).
		Claim("organization", opts.Organization).
		IssuedAt(time.Now()).
		Build()
	if err != nil {
		return nil, nil, err
	}
	serialized, err := jwt.Sign(token, jwt.WithKey(jwa.HS256, opts.key))
	if err != nil {
		return nil, nil, err
	}
	return &at, serialized, nil
}

func (t *AgentToken) String() string      { return t.ID }
func (t *AgentToken) IsSiteAdmin() bool   { return true }
func (t *AgentToken) IsOwner(string) bool { return true }

func (t *AgentToken) Organizations() []string { return []string{t.Organization} }

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
