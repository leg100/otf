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
	at                     *agentToken
	token                  []byte
	status                 RunnerStatus
	job                    *Job
}

func (f *fakeService) CreateAgentPool(ctx context.Context, opts CreateAgentPoolOptions) (*Pool, error) {
	f.createAgentPoolOptions = opts
	return f.pool, nil
}

func (f *fakeService) updateAgentPool(ctx context.Context, poolID resource.ID, opts updatePoolOptions) (*Pool, error) {
	return nil, nil
}

func (f *fakeService) listAgentPoolsByOrganization(context.Context, resource.OrganizationName, listPoolOptions) ([]*Pool, error) {
	return []*Pool{f.pool}, nil
}

func (f *fakeService) GetAgentPool(context.Context, resource.ID) (*Pool, error) {
	return f.pool, nil
}

func (f *fakeService) deleteAgentPool(ctx context.Context, poolID resource.ID) (*Pool, error) {
	return nil, nil
}

func (f *fakeService) CreateAgentToken(context.Context, resource.ID, CreateAgentTokenOptions) (*agentToken, []byte, error) {
	return f.at, f.token, nil
}

func (f *fakeService) ListAgentTokens(context.Context, resource.ID) ([]*agentToken, error) {
	return []*agentToken{f.at}, nil
}

func (f *fakeService) GetAgentToken(context.Context, resource.ID) (*agentToken, error) {
	return f.at, nil
}

func (f *fakeService) DeleteAgentToken(context.Context, resource.ID) (*agentToken, error) {
	return f.at, nil
}

func (f *fakeService) listJobs(ctx context.Context) ([]*Job, error) {
	return nil, nil
}

func (f *fakeService) allocateJob(ctx context.Context, jobID resource.ID, agentID resource.ID) (*Job, error) {
	if err := f.job.allocate(agentID); err != nil {
		return nil, err
	}
	return f.job, nil
}

func (f *fakeService) reallocateJob(ctx context.Context, jobID resource.ID, agentID resource.ID) (*Job, error) {
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

func (f *fakeService) updateStatus(ctx context.Context, runnerID resource.ID, status RunnerStatus) error {
	f.status = status
	return nil
}

func (f *fakeService) listRunners(ctx context.Context) ([]*RunnerMeta, error) {
	return nil, nil
}

func (f *fakeService) listRunnersByOrganization(ctx context.Context, organization organization.Name) ([]*RunnerMeta, error) {
	return nil, nil
}

func (f *fakeService) listRunnersByPool(ctx context.Context, poolID resource.ID) ([]*RunnerMeta, error) {
	return nil, nil
}

func (f *fakeService) listServerRunners(ctx context.Context) ([]*RunnerMeta, error) {
	return nil, nil
}

func (f *fakeService) deleteRunner(ctx context.Context, runnerID resource.ID) error {
	return nil
}
