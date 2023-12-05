package agent

import "context"

type fakeService struct {
	pool                   *Pool
	createAgentPoolOptions CreateAgentPoolOptions
	at                     *agentToken
	token                  []byte
	status                 AgentStatus
	deletedAgentID         string
	job                    *Job

	Service
}

func (f *fakeService) CreateAgentPool(ctx context.Context, opts CreateAgentPoolOptions) (*Pool, error) {
	f.createAgentPoolOptions = opts
	return f.pool, nil
}

func (f *fakeService) listAllAgentPools(ctx context.Context) ([]*Pool, error) {
	return []*Pool{f.pool}, nil
}

func (f *fakeService) listAgentPoolsByOrganization(context.Context, string, listPoolOptions) ([]*Pool, error) {
	return []*Pool{f.pool}, nil
}

func (f *fakeService) CreateAgentToken(context.Context, string, CreateAgentTokenOptions) (*agentToken, []byte, error) {
	return f.at, f.token, nil
}

func (f *fakeService) ListAgentTokens(context.Context, string) ([]*agentToken, error) {
	return []*agentToken{f.at}, nil
}

func (f *fakeService) DeleteAgentToken(context.Context, string) (*agentToken, error) {
	return f.at, nil
}

func (f *fakeService) updateAgentStatus(ctx context.Context, agentID string, status AgentStatus) error {
	f.status = status
	return nil
}

func (f *fakeService) deleteAgent(ctx context.Context, agentID string) error {
	f.deletedAgentID = agentID
	return nil
}

func (f *fakeService) allocateJob(ctx context.Context, spec JobSpec, agentID string) (*Job, error) {
	if err := f.job.allocate(agentID); err != nil {
		return nil, err
	}
	return f.job, nil
}

func (f *fakeService) reallocateJob(ctx context.Context, spec JobSpec, agentID string) (*Job, error) {
	if err := f.job.reallocate(agentID); err != nil {
		return nil, err
	}
	return f.job, nil
}
