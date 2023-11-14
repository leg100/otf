package agent

import (
	"context"

	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/hooks"
	"github.com/leg100/otf/internal/http/html"
	"github.com/leg100/otf/internal/logr"
	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/pubsub"
	"github.com/leg100/otf/internal/rbac"
	"github.com/leg100/otf/internal/run"
	"github.com/leg100/otf/internal/sql"
	"github.com/leg100/otf/internal/sql/pggen"
	"github.com/leg100/otf/internal/tfeapi"
	"github.com/leg100/otf/internal/tokens"
)

type (
	AgentService = Service

	Service interface {
		NewAllocator(pubsub.Subscriber) *allocator
		NewManager() *manager

		registerAgent(ctx context.Context, opts registerAgentOptions) (*Agent, error)
		getAgentJobs(ctx context.Context, agentID string) ([]*Job, error)
		updateAgentStatus(ctx context.Context, agentID string, status AgentStatus) error

		CreateAgentToken(ctx context.Context, opts CreateAgentTokenOptions) ([]byte, error)
		GetAgentToken(ctx context.Context, tokenID string) (*AgentToken, error)
		ListAgentTokens(ctx context.Context, organization string) ([]*AgentToken, error)
		DeleteAgentToken(ctx context.Context, id string) (*AgentToken, error)

		createJobToken(ctx context.Context, spec JobSpec) ([]byte, error)
		updateJobStatus(ctx context.Context, spec JobSpec, status JobStatus) error
	}

	service struct {
		logr.Logger
		pubsub.Subscriber
		run.RunService

		organization internal.Authorizer

		tfeapi *tfe
		api    *api
		web    *webHandlers

		*db
		*registrar
		*tokenFactory
	}

	ServiceOptions struct {
		logr.Logger
		*sql.DB
		*pubsub.Broker
		html.Renderer
		*tfeapi.Responder
		run.RunService
		tokens.TokensService
	}

	tokenFactory struct {
		tokens.TokensService
	}
)

func NewService(opts ServiceOptions) *service {
	svc := &service{
		Logger:       opts.Logger,
		db:           &db{DB: opts.DB},
		organization: &organization.Authorizer{Logger: opts.Logger},
		tokenFactory: &tokenFactory{
			TokensService: opts.TokensService,
		},
	}
	svc.tfeapi = &tfe{
		service:   svc,
		Responder: opts.Responder,
	}
	svc.api = &api{
		service:   svc,
		Responder: opts.Responder,
	}
	svc.web = &webHandlers{
		Renderer: opts.Renderer,
		Service:  svc,
	}
	svc.registrar = &registrar{
		service: svc,
	}
	// permit broker to transform database trigger events into agent pool events
	opts.Broker.Register("agent_pools", pubsub.GetterFunc(func(ctx context.Context, poolID string, action pubsub.DBAction) (any, error) {
		if action == pubsub.DeleteDBAction {
			return &Agent{ID: poolID}, nil
		}
		return svc.getPool(ctx, poolID)
	}))
	// permit broker to transform database trigger events into agent events
	opts.Broker.Register("agents", pubsub.GetterFunc(func(ctx context.Context, agentID string, action pubsub.DBAction) (any, error) {
		if action == pubsub.DeleteDBAction {
			return &Agent{ID: agentID}, nil
		}
		return svc.getAgent(ctx, agentID)
	}))
	// permit broker to transform database trigger events into job events
	opts.Broker.Register("jobs", pubsub.GetterFunc(func(ctx context.Context, jobspecString string, action pubsub.DBAction) (any, error) {
		spec, err := NewJobSpecFromString(jobspecString)
		if err != nil {
			return nil, err
		}
		if action == pubsub.DeleteDBAction {
			return &Job{JobSpec: spec}, nil
		}
		return svc.getJob(ctx, spec)
	}))
	// create jobs when a plan or apply is enqueued
	opts.AfterEnqueuePlan(svc.createJob)
	opts.AfterEnqueueApply(svc.createJob)
	// relay cancel signal from run service to agent.
	opts.AfterCancelSignal(svc.relaySignal(cancelSignal))
	// relay force-cancel signal from run service to agent.
	opts.AfterForceCancelSignal(svc.relaySignal(forceCancelSignal))
	// Register with auth middleware the agent token kind and a means of
	// retrieving an AgentToken corresponding to token's subject.
	opts.TokensService.RegisterKind(AgentTokenKind, func(ctx context.Context, tokenID string) (internal.Subject, error) {
		return svc.GetAgentToken(ctx, tokenID)
	})
	// Register with auth middleware the job token and a means of
	// retrieving Job corresponding to token.
	opts.TokensService.RegisterKind(JobTokenKind, func(ctx context.Context, jobspecString string) (internal.Subject, error) {
		spec, err := NewJobSpecFromString(jobspecString)
		if err != nil {
			return nil, err
		}
		return svc.getJob(ctx, spec)
	})
	return svc
}

func (s *service) AddHandlers(r *mux.Router) {
	s.tfeapi.addHandlers(r)
	s.api.addHandlers(r)
	s.web.addHandlers(r)
}

func (s *service) NewAllocator(subscriber pubsub.Subscriber) *allocator {
	return &allocator{
		Subscriber: subscriber,
		service:    s,
	}
}

func (s *service) NewManager() *manager {
	return &manager{
		service:  s,
		interval: defaultManagerInterval,
	}
}

func (s *service) createPool(ctx context.Context, opts createPoolOptions) (*Pool, error) {
	subject, err := s.organization.CanAccess(ctx, rbac.CreateRunAction, opts.Organization)
	if err != nil {
		return nil, err
	}
	pool, err := newPool(opts)
	if err != nil {
		s.Error(err, "creating agent pool", "subject", subject)
		return nil, err
	}
	if err := s.db.createPool(ctx, pool); err != nil {
		return nil, err
	}
	s.V(0).Info("created agent pool", "subject", subject, "pool", pool)
	return pool, nil
}

func (s *service) updatePool(ctx context.Context, poolID string, opts updatePoolOptions) (*Pool, error) {
	var (
		subject       internal.Subject
		before, after Pool
	)
	err := s.db.Lock(ctx, "agent_pools, agent_pool_allowed_workspaces", func(ctx context.Context, q pggen.Querier) (err error) {
		pool, err := s.db.getPool(ctx, poolID)
		if err != nil {
			return err
		}
		subject, err = s.organization.CanAccess(ctx, rbac.CreateRunAction, pool.Organization)
		if err != nil {
			return err
		}
		before = *pool
		after = *pool
		if err := after.update(opts); err != nil {
			return err
		}
		if err := s.db.updatePool(ctx, &after); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		s.Error(err, "updating agent pool", "agent_pool_id", poolID, "subject", subject)
		return nil, err
	}
	s.V(0).Info("updated agent pool", "subject", subject, "before", &before, "after", &after)
	return &after, nil
}

func (s *service) getPool(ctx context.Context, poolID string) (*Pool, error) {
	pool, err := s.db.getPool(ctx, poolID)
	if err != nil {
		s.Error(err, "retrieving agent pool", "agent_pool_id", poolID)
		return nil, err
	}
	subject, err := s.organization.CanAccess(ctx, rbac.CreateRunAction, pool.Organization)
	if err != nil {
		return nil, err
	}
	s.V(9).Info("retrieved agent pool", "subject", subject, "organization", pool.Organization)
	return pool, nil
}

func (s *service) listPools(ctx context.Context, opts listPoolOptions) ([]*Pool, error) {
	// TODO: handle authz with and without org
	subject, err := s.organization.CanAccess(ctx, rbac.CreateRunAction, "")
	if err != nil {
		return nil, err
	}
	pools, err := s.db.listPools(ctx, opts)
	if err != nil {
		s.Error(err, "listing agent pools", "subject", subject)
		return nil, err
	}
	s.V(9).Info("listed agent pools", "subject", subject, "count", len(pools))
	return pools, nil
}

func (s *service) deletePool(ctx context.Context, poolID string) error {
	var subject internal.Subject
	err := s.db.Tx(ctx, func(ctx context.Context, q pggen.Querier) error {
		organization, err := s.db.deletePool(ctx, poolID)
		if err != nil {
			return err
		}
		subject, err = s.organization.CanAccess(ctx, rbac.CreateRunAction, organization)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		s.Error(err, "deleting agent pool", "agent_pool_id", poolID, "subject", subject)
		return err
	}
	s.V(9).Info("deleted agent pool", "subject", subject)
	return nil
}

func (s *service) registerAgent(ctx context.Context, opts registerAgentOptions) (*Agent, error) {
	agent, err := func() (*Agent, error) {
		agent, err := s.register(ctx, opts)
		if err != nil {
			return nil, err
		}
		err = s.db.Tx(ctx, func(ctx context.Context, q pggen.Querier) error {
			if err := s.db.createAgent(ctx, agent); err != nil {
				return err
			}
			// an agent when registering can optionally send a list of current
			// jobs, which are jobs that were created by a previous agent before
			// it terminated.
			//
			// This has no purpose currently because jobs are part of the agent
			// process and when an agent terminates it terminates the jobs too.
			// But the project intends to introduce further methods of creating
			// jobs, via docker and kubernetes etc, and this behaviour will come
			// in useful then.
			for _, spec := range opts.CurrentJobs {
				return s.db.updateJob(ctx, spec, func(job *Job) error {
					return job.reallocate(agent.ID)
				})
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
		return agent, nil
	}()
	if err != nil {
		s.Error(err, "registering agent")
		return nil, err
	}
	s.V(0).Info("registered agent", "agent", agent)
	return agent, nil
}

func (s *service) getAgent(ctx context.Context, agentID string) (*Agent, error) {
	return nil, nil
}

func (s *service) updateAgentStatus(ctx context.Context, agentID string, status AgentStatus) error {
	return nil
}

func (s *service) listAgents(ctx context.Context) ([]*Agent, error) {
	return s.db.listAgents(ctx)
}

func (s *service) listAgentsByOrganization(ctx context.Context, organization string) ([]*Agent, error) {
	return s.db.listAgentsByOrganization(ctx, organization)
}

func (s *service) deleteAgent(ctx context.Context, agentID string) error {
	return nil
}

func (s *service) createJob(ctx context.Context, run *run.Run) error {
	job := newJob(run)
	if err := s.db.createJob(ctx, job); err != nil {
		return err
	}
	return nil
}

func (s *service) relaySignal(sig signal) hooks.Listener[*run.Run] {
	return func(ctx context.Context, run *run.Run) error {
		spec := JobSpec{RunID: run.ID, Phase: run.Phase()}
		return s.db.updateJob(ctx, spec, func(job *Job) error {
			return job.setSignal(sig)
		})
	}
}

// getAgentJobs returns jobs that either:
// (a) have JobAllocated status
// (b) have JobRunning status and a non-null signal
//
// getAgentJobs is intended to be called by an agent in order to retrieve jobs to
// run and jobs to cancel.
func (s *service) getAgentJobs(ctx context.Context, agentID string) ([]*Job, error) {
	sub, err := s.Subscribe(ctx, "get-agent-jobs-"+agentID)
	if err != nil {
		return nil, err
	}
	jobs, err := s.db.getAllocatedAndSignaledJobs(ctx, agentID)
	if err != nil {
		return nil, err
	}
	if len(jobs) > 0 {
		// return existing jobs
		return jobs, nil
	}
	// wait for a job matching criteria to arrive:
	for event := range sub {
		job, ok := event.Payload.(*Job)
		if !ok {
			continue
		}
		if job.AgentID == nil || *job.AgentID != agentID {
			continue
		}
		switch job.Status {
		case JobAllocated:
			return []*Job{job}, nil
		case JobRunning:
			if job.signal != nil {
				return []*Job{job}, nil
			}
		}
	}
	return nil, nil
}

func (s *service) getJob(ctx context.Context, spec JobSpec) (*Job, error) {
	return s.db.getJob(ctx, spec)
}

func (s *service) listJobs(ctx context.Context) ([]*Job, error) {
	return s.db.listJobs(ctx)
}

func (s *service) allocateJob(ctx context.Context, spec JobSpec, agentID string) error {
	return s.db.updateJob(ctx, spec, func(job *Job) error {
		return job.allocate(agentID)
	})
}

func (s *service) reallocateJob(ctx context.Context, spec JobSpec, agentID string) error {
	return s.db.updateJob(ctx, spec, func(job *Job) error {
		return job.reallocate(agentID)
	})
}

func (s *service) updateJobStatus(ctx context.Context, spec JobSpec, status JobStatus) error {
	return s.db.updateJob(ctx, spec, func(job *Job) error {
		if err := job.updateStatus(status); err != nil {
			return err
		}
		// update corresponding run phase too
		var err error
		switch status {
		case JobRunning:
			_, err = s.RunService.StartPhase(ctx, spec.RunID, spec.Phase, run.PhaseStartOptions{})
		case JobFinished, JobErrored:
			_, err = s.RunService.FinishPhase(ctx, spec.RunID, spec.Phase, run.PhaseFinishOptions{
				Errored: status == JobErrored,
			})
		case JobCanceled:
			// set immediate=true to skip sending a signal and looping back
			// round again
			_, err = s.RunService.Cancel(ctx, spec.RunID, true)
		}
		return err
	})
}

// agent tokens

func (a *service) CreateAgentToken(ctx context.Context, opts CreateAgentTokenOptions) ([]byte, error) {
	subject, err := a.organization.CanAccess(ctx, rbac.CreateAgentTokenAction, opts.Organization)
	if err != nil {
		return nil, err
	}

	at, token, err := a.NewAgentToken(opts)
	if err != nil {
		return nil, err
	}
	if err := a.db.createAgentToken(ctx, at); err != nil {
		a.Error(err, "creating agent token", "organization", opts.Organization, "id", at.ID, "subject", subject)
		return nil, err
	}
	a.V(0).Info("created agent token", "organization", opts.Organization, "id", at.ID, "subject", subject)
	return token, nil
}

func (a *service) GetAgentToken(ctx context.Context, tokenID string) (*AgentToken, error) {
	at, err := a.db.getAgentTokenByID(ctx, tokenID)
	if err != nil {
		a.Error(err, "retrieving agent token", "token", "******")
		return nil, err
	}
	a.V(9).Info("retrieved agent token", "organization", at.Organization, "id", at.ID)
	return at, nil
}

func (a *service) ListAgentTokens(ctx context.Context, organization string) ([]*AgentToken, error) {
	subject, err := a.organization.CanAccess(ctx, rbac.ListAgentTokensAction, organization)
	if err != nil {
		return nil, err
	}

	tokens, err := a.db.listAgentTokens(ctx, organization)
	if err != nil {
		a.Error(err, "listing agent tokens", "organization", organization, "subject", subject)
		return nil, err
	}
	a.V(9).Info("listed agent tokens", "organization", organization, "subject", subject)
	return tokens, nil
}

func (a *service) DeleteAgentToken(ctx context.Context, id string) (*AgentToken, error) {
	// retrieve agent token first in order to get organization for authorization
	at, err := a.db.getAgentTokenByID(ctx, id)
	if err != nil {
		// we can't reveal any info because all we have is the
		// authentication token which is sensitive.
		a.Error(err, "retrieving agent token", "token", "******")
		return nil, err
	}

	subject, err := a.organization.CanAccess(ctx, rbac.DeleteAgentTokenAction, at.Organization)
	if err != nil {
		return nil, err
	}

	if err := a.db.deleteAgentToken(ctx, id); err != nil {
		a.Error(err, "deleting agent token", "agent token", at, "subject", subject)
		return nil, err
	}
	a.V(0).Info("deleted agent token", "agent token", at, "subject", subject)
	return at, nil
}
