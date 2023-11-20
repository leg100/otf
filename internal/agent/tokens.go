package agent

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/tokens"
)

const (
	AgentTokenKind tokens.Kind = "agent_token"
	JobTokenKind   tokens.Kind = "job_token"

	defaultJobTokenExpiry = 60 * time.Minute
)

type (
	// agentToken represents the authentication token for an agent.
	// NOTE: the cryptographic token itself is not retained.
	agentToken struct {
		ID          string `jsonapi:"primary,agent_tokens"`
		CreatedAt   time.Time
		AgentPoolID string `jsonapi:"attribute" json:"agent_pool_id"`
		Description string `jsonapi:"attribute" json:"description"`
	}

	CreateAgentTokenOptions struct {
		Description string `json:"description" schema:"description,required"`
	}
)

func (a *agentToken) LogValue() slog.Value {
	attrs := []slog.Attr{
		slog.String("id", a.ID),
		slog.String("agent_pool_id", string(a.AgentPoolID)),
		slog.String("description", a.Description),
	}
	return slog.GroupValue(attrs...)
}

type tokenFactory struct {
	tokens.TokensService
}

// createJobToken constructs a job token
func (f *tokenFactory) createJobToken(spec JobSpec) ([]byte, error) {
	expiry := internal.CurrentTimestamp(nil).Add(defaultJobTokenExpiry)
	return f.NewToken(tokens.NewTokenOptions{
		Subject: spec.String(),
		Kind:    JobTokenKind,
		Expiry:  &expiry,
	})
}

// NewAgentToken constructs a token for an agent, returning both the
// representation of the token, and the cryptographic token itself.
func (f *tokenFactory) NewAgentToken(poolID string, opts CreateAgentTokenOptions) (*agentToken, []byte, error) {
	if poolID == "" {
		return nil, nil, fmt.Errorf("agent pool ID cannot be an empty string")
	}
	if opts.Description == "" {
		return nil, nil, fmt.Errorf("description cannot be an empty string")
	}
	at := agentToken{
		ID:          internal.NewID("at"),
		CreatedAt:   internal.CurrentTimestamp(nil),
		Description: opts.Description,
		AgentPoolID: poolID,
	}
	token, err := f.NewToken(tokens.NewTokenOptions{
		Subject: at.ID,
		Kind:    AgentTokenKind,
		Claims: map[string]string{
			"agent_pool_id": poolID,
		},
	})
	if err != nil {
		return nil, nil, err
	}
	return &at, token, nil
}
