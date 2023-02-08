package otf

import "context"

type AgentToken interface {
	Token() string

	Subject
}

type CreateAgentTokenOptions struct {
	Organization string `schema:"organization_name,required"`
	Description  string `schema:"description,required"`
}

// AgentTokenService provides access to agent tokens
type AgentTokenService interface {
	GetAgentToken(ctx context.Context, token string) (AgentToken, error)
}
