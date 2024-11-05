package runner

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/tokens"
)

const (
	AgentTokenKind resource.Kind = "at"

	defaultJobTokenExpiry = 60 * time.Minute
)

type (
	// agentToken represents the authentication token for an agent.
	// NOTE: the cryptographic token itself is not retained.
	agentToken struct {
		resource.ID `jsonapi:"primary,agent_tokens"`

		CreatedAt   time.Time
		AgentPoolID resource.ID `jsonapi:"attribute" json:"agent_pool_id"`
		Description string      `jsonapi:"attribute" json:"description"`
	}

	CreateAgentTokenOptions struct {
		Description string `json:"description" schema:"description,required"`
	}
)

func (a *agentToken) LogValue() slog.Value {
	attrs := []slog.Attr{
		slog.String("id", a.ID.String()),
		slog.String("agent_pool_id", string(a.AgentPoolID.String())),
		slog.String("description", a.Description),
	}
	return slog.GroupValue(attrs...)
}

type tokenFactory struct {
	tokens *tokens.Service
}

// createJobToken constructs a job token
func (f *tokenFactory) createJobToken(jobID resource.ID) ([]byte, error) {
	expiry := internal.CurrentTimestamp(nil).Add(defaultJobTokenExpiry)
	return f.tokens.NewToken(jobID, tokens.WithExpiry(expiry))
}

// NewAgentToken constructs a token for an agent, returning both the
// representation of the token, and the cryptographic token itself.
func (f *tokenFactory) NewAgentToken(poolID resource.ID, opts CreateAgentTokenOptions) (*agentToken, []byte, error) {
	if opts.Description == "" {
		return nil, nil, fmt.Errorf("description cannot be an empty string")
	}
	at := agentToken{
		ID:          resource.NewID(AgentTokenKind),
		CreatedAt:   internal.CurrentTimestamp(nil),
		Description: opts.Description,
		AgentPoolID: poolID,
	}
	token, err := f.tokens.NewToken(at.ID)
	if err != nil {
		return nil, nil, err
	}
	return &at, token, nil
}
