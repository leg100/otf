package runner

import (
	"context"

	"github.com/leg100/otf/internal/pubsub"
)

type fakeService struct {
	pool                   *Pool
	createAgentPoolOptions CreateAgentPoolOptions
	at                     *agentToken
	token                  []byte
	status                 RunnerStatus
	job                    *Job
}

func (f *fakeService) CreateAgentPool(ctx context.Context, opts CreateAgentPoolOptions) (*Pool, error) {
	f.createAgentPoolOptions = opts
	return f.pool, nil
}

func (f *fakeService) updateAgentPool(ctx context.Context, poolID string, opts updatePoolOptions) (*Pool, error) {
	return nil, nil
}

func (f *fakeService) listAllAgentPools(ctx context.Context) ([]*Pool, error) {
	return []*Pool{f.pool}, nil
}

func (f *fakeService) listAgentPoolsByOrganization(context.Context, string, listPoolOptions) ([]*Pool, error) {
	return []*Pool{f.pool}, nil
}

func (f *fakeService) GetAgentPool(context.Context, string) (*Pool, error) {
	return f.pool, nil
}

func (f *fakeService) deleteAgentPool(ctx context.Context, poolID string) (*Pool, error) {
	return nil, nil
}

func (f *fakeService) CreateAgentToken(context.Context, string, CreateAgentTokenOptions) (*agentToken, []byte, error) {
	return f.at, f.token, nil
}

func (f *fakeService) ListAgentTokens(context.Context, string) ([]*agentToken, error) {
	return []*agentToken{f.at}, nil
}

func (f *fakeService) GetAgentToken(context.Context, string) (*agentToken, error) {
	return f.at, nil
}

func (f *fakeService) DeleteAgentToken(context.Context, string) (*agentToken, error) {
	return f.at, nil
}

func (f *fakeService) listJobs(ctx context.Context) ([]*Job, error) {
	return nil, nil
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

func (f *fakeService) WatchJobs(context.Context) (<-chan pubsub.Event[*Job], func()) {
	return nil, nil
}

func (f *fakeService) WatchRunners(context.Context) (<-chan pubsub.Event[*RunnerMeta], func()) {
	return nil, nil
}

func (f *fakeService) register(ctx context.Context, opts registerOptions) (*RunnerMeta, error) {
	return nil, nil
}

func (f *fakeService) updateStatus(ctx context.Context, runnerID string, status RunnerStatus) error {
	f.status = status
	return nil
}

func (f *fakeService) listRunners(ctx context.Context) ([]*RunnerMeta, error) {
	return nil, nil
}

func (f *fakeService) listRunnersByOrganization(ctx context.Context, organization string) ([]*RunnerMeta, error) {
	return nil, nil
}

func (f *fakeService) listRunnersByPool(ctx context.Context, poolID string) ([]*RunnerMeta, error) {
	return nil, nil
}

func (f *fakeService) listServerRunners(ctx context.Context) ([]*RunnerMeta, error) {
	return nil, nil
}

func (f *fakeService) deleteRunner(ctx context.Context, runnerID string) error {
	return nil
}
