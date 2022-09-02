package otf

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgtype"
)

// AgentToken is an authentication token for an agent.
type AgentToken struct {
	id               string
	createdAt        time.Time
	token            string
	description      string
	organizationName string
}

func (t *AgentToken) ID() string               { return t.id }
func (t *AgentToken) String() string           { return t.id }
func (t *AgentToken) Token() string            { return t.token }
func (t *AgentToken) CreatedAt() time.Time     { return t.createdAt }
func (t *AgentToken) Description() string      { return t.description }
func (t *AgentToken) OrganizationName() string { return t.organizationName }

// CanAccess implements the Subject interface - an agent can only acccess its
// organization resources.
func (t *AgentToken) CanAccess(organizationName *string) bool {
	if organizationName == nil {
		return false
	}
	return t.organizationName == *organizationName
}

type AgentTokenCreateOptions struct {
	OrganizationName string
	Description      string
}

func NewAgentToken(opts AgentTokenCreateOptions) (*AgentToken, error) {
	if opts.OrganizationName == "" {
		return nil, fmt.Errorf("organization name cannot be an empty string")
	}
	if opts.Description == "" {
		return nil, fmt.Errorf("description cannot be an empty string")
	}
	t, err := GenerateToken()
	if err != nil {
		return nil, fmt.Errorf("generating token: %w", err)
	}
	token := AgentToken{
		id:               NewID("at"),
		createdAt:        CurrentTimestamp(),
		token:            t,
		description:      opts.Description,
		organizationName: opts.OrganizationName,
	}
	return &token, nil
}

type AgentTokenRow struct {
	TokenID          pgtype.Text        `json:"token_id"`
	Token            pgtype.Text        `json:"token"`
	CreatedAt        pgtype.Timestamptz `json:"created_at"`
	Description      pgtype.Text        `json:"description"`
	OrganizationName pgtype.Text        `json:"organization_name"`
}

// UnmarshalAgentTokenDBResult unmarshals a row from the database.
func UnmarshalAgentTokenDBResult(row AgentTokenRow) *AgentToken {
	return &AgentToken{
		id:               row.TokenID.String,
		createdAt:        row.CreatedAt.Time,
		token:            row.Token.String,
		description:      row.Description.String,
		organizationName: row.OrganizationName.String,
	}
}

// AgentTokenService provides access to agent tokens
type AgentTokenService interface {
	CreateAgentToken(ctx context.Context, opts AgentTokenCreateOptions) (*AgentToken, error)
	// GetAgentToken retrieves agent token using its cryptographic
	// authentication token.
	GetAgentToken(ctx context.Context, token string) (*AgentToken, error)
	ListAgentTokens(ctx context.Context, organizationName string) ([]*AgentToken, error)
	DeleteAgentToken(ctx context.Context, id string) error
}

// AgentTokenStore persists agent authentication tokens.
type AgentTokenStore interface {
	CreateAgentToken(ctx context.Context, at *AgentToken) error
	// GetAgentToken retrieves agent token using its cryptographic
	// authentication token.
	GetAgentToken(ctx context.Context, token string) (*AgentToken, error)
	ListAgentTokens(ctx context.Context, organizationName string) ([]*AgentToken, error)
	DeleteAgentToken(ctx context.Context, id string) error
}
