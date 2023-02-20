package auth

import (
	"context"

	"github.com/leg100/otf"
)

type fakeAgentTokenApp struct {
	token *AgentToken

	agentTokenApp
}

func (f *fakeAgentTokenApp) createAgentToken(context.Context, otf.CreateAgentTokenOptions) (*AgentToken, error) {
	return f.token, nil
}

func (f *fakeAgentTokenApp) listAgentTokens(context.Context, string) ([]*AgentToken, error) {
	return []*AgentToken{f.token}, nil
}

func (f *fakeAgentTokenApp) deleteAgentToken(context.Context, string) (*AgentToken, error) {
	return f.token, nil
}
