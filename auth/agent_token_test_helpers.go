package auth

import (
	"context"
	"testing"

	"github.com/leg100/otf"
	"github.com/stretchr/testify/require"
)

func NewTestAgentToken(t *testing.T, org string) *AgentToken {
	token, err := newAgentToken(otf.CreateAgentTokenOptions{
		Organization: org,
		Description:  "lorem ipsum...",
	})
	require.NoError(t, err)
	return token
}

type fakeAgentTokenService struct {
	token *AgentToken

	agentTokenService
}

func (f *fakeAgentTokenService) createAgentToken(context.Context, otf.CreateAgentTokenOptions) (*AgentToken, error) {
	return f.token, nil
}

func (f *fakeAgentTokenService) listAgentTokens(context.Context, string) ([]*AgentToken, error) {
	return []*AgentToken{f.token}, nil
}

func (f *fakeAgentTokenService) deleteAgentToken(context.Context, string) (*AgentToken, error) {
	return f.token, nil
}
