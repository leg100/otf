package agenttoken

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgtype"
	"github.com/leg100/otf/http/jsonapi"
	"github.com/leg100/otf/rbac"
)

// AgentToken is an long-lived authentication token for an external agent.
type AgentToken struct {
	id           string
	createdAt    time.Time
	token        *string
	description  string
	organization string
}

func (t *AgentToken) ID() string           { return t.id }
func (t *AgentToken) String() string       { return t.id }
func (t *AgentToken) Token() *string       { return t.token }
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

func (t *AgentToken) CanAccessWorkspace(action rbac.Action, policy *WorkspacePolicy) bool {
	// agent can access anything within its organization
	return t.organization == policy.Organization
}

type CreateAgentTokenOptions struct {
	Organization string `schema:"organization_name,required"`
	Description  string `schema:"description,required"`
}

func NewAgentToken(opts CreateAgentTokenOptions) (*AgentToken, error) {
	if opts.Organization == "" {
		return nil, fmt.Errorf("organization name cannot be an empty string")
	}
	if opts.Description == "" {
		return nil, fmt.Errorf("description cannot be an empty string")
	}
	t, err := GenerateAuthToken("agent")
	if err != nil {
		return nil, fmt.Errorf("generating token: %w", err)
	}
	token := AgentToken{
		id:           NewID("at"),
		createdAt:    CurrentTimestamp(),
		token:        &t,
		description:  opts.Description,
		organization: opts.Organization,
	}
	return &token, nil
}

func UnmarshalAgentTokenJSONAPI(dto *jsonapi.AgentToken) *AgentToken {
	at := &AgentToken{
		id:           dto.ID,
		organization: dto.Organization,
	}
	if dto.Token != nil {
		at.token = dto.Token
	}
	return at
}

type AgentTokenRow struct {
	TokenID          pgtype.Text        `json:"token_id"`
	Token            pgtype.Text        `json:"token"`
	CreatedAt        pgtype.Timestamptz `json:"created_at"`
	Description      pgtype.Text        `json:"description"`
	OrganizationName pgtype.Text        `json:"organization_name"`
}

// UnmarshalAgentTokenResult unmarshals a row from the database.
func UnmarshalAgentTokenResult(row AgentTokenRow) *AgentToken {
	return &AgentToken{
		id:           row.TokenID.String,
		createdAt:    row.CreatedAt.Time,
		token:        String(row.Token.String),
		description:  row.Description.String,
		organization: row.OrganizationName.String,
	}
}

// AgentTokenService provides access to agent tokens
type AgentTokenService interface {
	CreateAgentToken(ctx context.Context, opts CreateAgentTokenOptions) (*AgentToken, error)
	// GetAgentToken retrieves AgentToken using its cryptographic
	// authentication token.
	GetAgentToken(ctx context.Context, token string) (*AgentToken, error)
	ListAgentTokens(ctx context.Context, organization string) ([]*AgentToken, error)
	DeleteAgentToken(ctx context.Context, id string) (*AgentToken, error)
}

// AgentTokenStore persists agent authentication tokens.
type AgentTokenStore interface {
	CreateAgentToken(ctx context.Context, at *AgentToken) error
	// GetAgentTokenByID retrieves agent token using its ID.
	GetAgentTokenByID(ctx context.Context, id string) (*AgentToken, error)
	// GetAgentTokenByToken retrieves agent token using its cryptographic
	// authentication token.
	GetAgentTokenByToken(ctx context.Context, token string) (*AgentToken, error)
	ListAgentTokens(ctx context.Context, organization string) ([]*AgentToken, error)
	DeleteAgentToken(ctx context.Context, id string) error
}
