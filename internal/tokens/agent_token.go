package tokens

import (
	"context"
	"fmt"
	"time"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/rbac"
	"github.com/lestrrat-go/jwx/v2/jwk"
)

type (
	// AgentToken represents the authentication token for an external agent.
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

	NewAgentTokenOptions struct {
		CreateAgentTokenOptions
		key jwk.Key // key for signing new token
	}

	agentTokenService interface {
		CreateAgentToken(ctx context.Context, options CreateAgentTokenOptions) ([]byte, error)
		GetAgentToken(ctx context.Context, id string) (*AgentToken, error)
		ListAgentTokens(ctx context.Context, organization string) ([]*AgentToken, error)
		DeleteAgentToken(ctx context.Context, id string) (*AgentToken, error)
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
		ID:           internal.NewID("at"),
		CreatedAt:    internal.CurrentTimestamp(nil),
		Description:  opts.Description,
		Organization: opts.Organization,
	}
	token, err := NewToken(NewTokenOptions{
		key:     opts.key,
		Subject: at.ID,
		Kind:    agentTokenKind,
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

// AgentFromContext retrieves an agent token from a context
func AgentFromContext(ctx context.Context) (*AgentToken, error) {
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

func (a *service) GetAgentToken(ctx context.Context, tokenID string) (*AgentToken, error) {
	at, err := a.db.getAgentTokenByID(ctx, tokenID)
	if err != nil {
		a.Error(err, "retrieving agent token", "token", "******")
		return nil, err
	}
	a.V(9).Info("retrieved agent token", "organization", at.Organization, "id", at.ID)
	return at, nil
}

func (a *service) CreateAgentToken(ctx context.Context, opts CreateAgentTokenOptions) ([]byte, error) {
	subject, err := a.organization.CanAccess(ctx, rbac.CreateAgentTokenAction, opts.Organization)
	if err != nil {
		return nil, err
	}

	at, token, err := NewAgentToken(NewAgentTokenOptions{
		CreateAgentTokenOptions: opts,
		key:                     a.key,
	})
	if err != nil {
		return nil, err
	}
	if err := a.db.createAgentToken(ctx, at); err != nil {
		a.Error(err, "creating agent token", "organization", opts.Organization, "id", at.ID, "subject", subject)
		return nil, err
	}
	a.V(0).Info("created agent token", "organization", opts.Organization, "id", at.ID, "subject", subject)
	return token, nil
}

func (a *service) ListAgentTokens(ctx context.Context, organization string) ([]*AgentToken, error) {
	subject, err := a.organization.CanAccess(ctx, rbac.ListAgentTokensAction, organization)
	if err != nil {
		return nil, err
	}

	tokens, err := a.db.listAgentTokens(ctx, organization)
	if err != nil {
		a.Error(err, "listing agent tokens", "organization", organization, "subject", subject)
		return nil, err
	}
	a.V(9).Info("listed agent tokens", "organization", organization, "subject", subject)
	return tokens, nil
}

func (a *service) DeleteAgentToken(ctx context.Context, id string) (*AgentToken, error) {
	// retrieve agent token first in order to get organization for authorization
	at, err := a.db.getAgentTokenByID(ctx, id)
	if err != nil {
		// we can't reveal any info because all we have is the
		// authentication token which is sensitive.
		a.Error(err, "retrieving agent token", "token", "******")
		return nil, err
	}

	subject, err := a.organization.CanAccess(ctx, rbac.DeleteAgentTokenAction, at.Organization)
	if err != nil {
		return nil, err
	}

	if err := a.db.deleteAgentToken(ctx, id); err != nil {
		a.Error(err, "deleting agent token", "agent token", at, "subject", subject)
		return nil, err
	}
	a.V(0).Info("deleted agent token", "agent token", at, "subject", subject)
	return at, nil
}
