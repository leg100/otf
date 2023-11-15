package agent

import "context"

type fakeService struct {
	at    *agentToken
	token []byte

	Service
}

func (f *fakeService) CreateAgentToken(context.Context, CreateAgentTokenOptions) ([]byte, error) {
	return f.token, nil
}

func (f *fakeService) ListAgentTokens(context.Context, string) ([]*agentToken, error) {
	return []*agentToken{f.at}, nil
}

func (f *fakeService) DeleteAgentToken(context.Context, string) (*agentToken, error) {
	return f.at, nil
}
