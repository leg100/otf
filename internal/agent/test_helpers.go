package agent

import "context"

type fakeService struct {
	at    *AgentToken
	token []byte

	Service
}

func (f *fakeService) CreateAgentToken(context.Context, CreateAgentTokenOptions) ([]byte, error) {
	return f.token, nil
}

func (f *fakeService) ListAgentTokens(context.Context, string) ([]*AgentToken, error) {
	return []*AgentToken{f.at}, nil
}

func (f *fakeService) DeleteAgentToken(context.Context, string) (*AgentToken, error) {
	return f.at, nil
}
