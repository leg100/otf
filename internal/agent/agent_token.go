package agent

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgtype"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/rbac"
	"github.com/leg100/otf/internal/tokens"
)

const AgentTokenKind tokens.Kind = "agent_token"

type (
	// AgentToken represents the authentication token for an agent.
	// NOTE: the cryptographic token itself is not retained.
	AgentToken struct {
		ID           string `jsonapi:"primary,agent_tokens"`
		CreatedAt    time.Time
		Description  string `jsonapi:"attribute" json:"description"`
		Organization string `jsonapi:"attribute" json:"organization_name"`
		AgentPoolID  string `jsonapi:"attribute" json:"agent_pool_id"`
	}

	CreateAgentTokenOptions struct {
		Organization string `json:"organization_name" schema:"organization_name,required"`
		Description  string `json:"description" schema:"description,required"`
	}
)

// NewAgentToken constructs a token for an agent, returning both the
// representation of the token, and the cryptographic token itself.
func (f *tokenFactory) NewAgentToken(opts CreateAgentTokenOptions) (*AgentToken, []byte, error) {
	if opts.Organization == "" {
		return nil, nil, fmt.Errorf("organization name cannot be an empty string")
	}
	if opts.Description == "" {
		return nil, nil, fmt.Errorf("description cannot be an empty string")
	}
	at := AgentToken{
		ID:           internal.NewID("at"),
		CreatedAt:    internal.CurrentTimestamp(nil),
		Description:  opts.Description,
		Organization: opts.Organization,
	}
	token, err := f.NewToken(tokens.NewTokenOptions{
		Subject: at.ID,
		Kind:    AgentTokenKind,
		Claims: map[string]string{
			"organization": opts.Organization,
		},
	})
	if err != nil {
		return nil, nil, err
	}
	return &at, token, nil
}

func (t *AgentToken) String() string      { return t.ID }
func (t *AgentToken) IsSiteAdmin() bool   { return true }
func (t *AgentToken) IsOwner(string) bool { return true }

func (t *AgentToken) Organizations() []string { return []string{t.Organization} }

func (*AgentToken) CanAccessSite(action rbac.Action) bool {
	// agent cannot carry out site-level actions
	return false
}

func (*AgentToken) CanAccessTeam(rbac.Action, string) bool {
	// agent cannot carry out team-level actions
	return false
}

func (t *AgentToken) CanAccessOrganization(action rbac.Action, name string) bool {
	return t.Organization == name
}

func (t *AgentToken) CanAccessWorkspace(action rbac.Action, policy internal.WorkspacePolicy) bool {
	// agent can access anything within its organization
	return t.Organization == policy.Organization
}

// AgentTokenFromContext retrieves an agent token from a context
func AgentTokenFromContext(ctx context.Context) (*AgentToken, error) {
	subj, err := internal.SubjectFromContext(ctx)
	if err != nil {
		return nil, err
	}
	agent, ok := subj.(*AgentToken)
	if !ok {
		return nil, fmt.Errorf("subject found in context but it is not an agent")
	}
	return agent, nil
}

type agentTokenRow struct {
	AgentTokenID     pgtype.Text        `json:"agent_token_id"`
	CreatedAt        pgtype.Timestamptz `json:"created_at"`
	Description      pgtype.Text        `json:"description"`
	OrganizationName pgtype.Text        `json:"organization_name"`
	AgentPoolID      pgtype.Text        `json:"agent_pool_id"`
}

func (row agentTokenRow) toAgentToken() *AgentToken {
	return &AgentToken{
		ID:           row.AgentTokenID.String,
		CreatedAt:    row.CreatedAt.Time.UTC(),
		Description:  row.Description.String,
		Organization: row.OrganizationName.String,
	}
}
