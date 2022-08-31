package otf

import (
	"context"
	"fmt"
	"time"

	"github.com/leg100/otf/sql/pggen"
)

// AgentToken is an authentication token for an agent.
type AgentToken struct {
	id        string
	createdAt time.Time
	token     string
	// Token belongs to an organization
	organizationID string
}

func (t *AgentToken) ID() string             { return t.id }
func (t *AgentToken) Token() string          { return t.token }
func (t *AgentToken) CreatedAt() time.Time   { return t.createdAt }
func (t *AgentToken) OrganizationID() string { return t.organizationID }

// AgentTokenService shares same method signatures as AgentTokenStore
type AgentTokenService AgentTokenStore

// AgentTokenStore persists agent authentication tokens.
type AgentTokenStore interface {
	CreateAgentToken(ctx context.Context, token *Token) error
	GetAgentToken(ctx context.Context, id string) (*Token, error)
	DeleteAgentToken(ctx context.Context, id string) error
}

func NewAgentToken(organizationID string) (*AgentToken, error) {
	t, err := GenerateToken()
	if err != nil {
		return nil, fmt.Errorf("generating token: %w", err)
	}
	token := AgentToken{
		id:             NewID("at"),
		createdAt:      CurrentTimestamp(),
		token:          t,
		organizationID: organizationID,
	}
	return &token, nil
}

// UnmarshalAgentTokenDBResult unmarshals a row from the database.
func UnmarshalAgentTokenDBResult(typ pggen.FindAgentTokenRow) (*AgentToken, error) {
	token := AgentToken{
		id:             typ.TokenID.String,
		createdAt:      typ.CreatedAt.Time,
		token:          typ.Token.String,
		organizationID: typ.OrganizationID.String,
	}
	return &token, nil
}
