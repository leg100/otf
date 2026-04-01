package runner

import (
	"context"

	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/pubsub"
	"github.com/leg100/otf/internal/resource"
)

type fakeService struct {
	pool                   *Pool
	createAgentPoolOptions CreateAgentPoolOptions
	at                     *AgentToken
	token                  []byte
	status                 RunnerStatus
	job                    *Job
}

func (f *fakeService) CreateAgentPool(ctx context.Context, opts CreateAgentPoolOptions) (*Pool, error) {
	f.createAgentPoolOptions = opts
	return f.pool, nil
}

func (f *fakeService) UpdateAgentPool(ctx context.Context, poolID resource.ID, opts UpdatePoolOptions) (*Pool, error) {
	return nil, nil
}

func (f *fakeService) ListAgentPoolsByOrganization(context.Context, organization.Name, ListPoolOptions) ([]*Pool, error) {
	return []*Pool{f.pool}, nil
}

func (f *fakeService) GetAgentPool(context.Context, resource.ID) (*Pool, error) {
	return f.pool, nil
}

func (f *fakeService) DeleteAgentPool(ctx context.Context, poolID resource.ID) (*Pool, error) {
	return nil, nil
}

func (f *fakeService) CreateAgentToken(context.Context, resource.ID, CreateAgentTokenOptions) (*AgentToken, []byte, error) {
	return f.at, f.token, nil
}

func (f *fakeService) ListAgentTokens(context.Context, resource.ID) ([]*AgentToken, error) {
	return []*AgentToken{f.at}, nil
}

func (f *fakeService) GetAgentToken(context.Context, resource.ID) (*AgentToken, error) {
	return f.at, nil
}

func (f *fakeService) DeleteAgentToken(context.Context, resource.ID) (*AgentToken, error) {
	return f.at, nil
}

func (f *fakeService) listJobs(ctx context.Context) ([]*Job, error) {
	return nil, nil
}

func (f *fakeService) allocateJob(ctx context.Context, jobID resource.ID, agentID resource.ID) (*Job, error) {
	if err := f.job.allocate(agentID.(resource.TfeID)); err != nil {
		return nil, err
	}
	return f.job, nil
}

func (f *fakeService) reallocateJob(ctx context.Context, jobID resource.ID, agentID resource.ID) (*Job, error) {
	if err := f.job.reallocate(agentID.(resource.TfeID)); err != nil {
		return nil, err
	}
	return f.job, nil
}

func (f *fakeService) GetJob(ctx context.Context, jobID resource.ID) (*Job, error) {
	return f.job, nil
}

func (f *fakeService) WatchJobs(context.Context) (<-chan pubsub.Event[*JobEvent], func()) {
	return nil, nil
}

func (f *fakeService) WatchRunners(context.Context) (<-chan pubsub.Event[*RunnerEvent], func()) {
	return nil, nil
}

func (f *fakeService) getRunner(ctx context.Context, runnerID resource.ID) (*RunnerMeta, error) {
	return nil, nil
}

func (f *fakeService) updateStatus(ctx context.Context, runnerID resource.ID, status RunnerStatus) error {
	f.status = status
	return nil
}

func (f *fakeService) ListRunners(ctx context.Context, opts ListOptions) ([]*RunnerMeta, error) {
	return nil, nil
}

func (f *fakeService) DeleteRunner(ctx context.Context, runnerID resource.ID) error {
	return nil
}
