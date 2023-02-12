package auth

import (
	"context"

	"github.com/leg100/otf"
)

type fakeAgentTokenApp struct {
	token *agentToken

	agentTokenApp
}

func (f *fakeAgentTokenApp) createAgentToken(context.Context, otf.CreateAgentTokenOptions) (*agentToken, error) {
	return f.token, nil
}

func (f *fakeAgentTokenApp) listAgentTokens(context.Context, string) ([]*agentToken, error) {
	return []*agentToken{f.token}, nil
}

func (f *fakeAgentTokenApp) deleteAgentToken(context.Context, string) (*agentToken, error) {
	return f.token, nil
}
