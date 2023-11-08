package agent

import (
	"context"

	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/hooks"
	"github.com/leg100/otf/internal/logr"
	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/pubsub"
	"github.com/leg100/otf/internal/rbac"
	"github.com/leg100/otf/internal/run"
	"github.com/leg100/otf/internal/sql"
	"github.com/leg100/otf/internal/sql/pggen"
	"github.com/leg100/otf/internal/tfeapi"
)

type (
	AgentService = Service

	Service interface {
		NewAllocator(pubsub.Subscriber) *allocator
		NewManager() *manager
	}

	service struct {
		logr.Logger
		pubsub.Subscriber
		run.RunService

		organization internal.Authorizer

		*db
		tfeapi *tfe
		api    *api
		*registrar
	}

	ServiceOptions struct {
		logr.Logger
		*sql.DB
		*tfeapi.Responder
		run.RunService
	}
)

func NewService(opts ServiceOptions) *service {
	svc := &service{
		Logger:       opts.Logger,
		db:           &db{DB: opts.DB},
		organization: &organization.Authorizer{Logger: opts.Logger},
	}
	svc.tfeapi = &tfe{
		service:   svc,
		Responder: opts.Responder,
	}
	svc.api = &api{
		service:   svc,
		Responder: opts.Responder,
	}
	svc.registrar = &registrar{
		service: svc,
	}
	// create jobs when a plan or apply is enqueued
	opts.AfterEnqueuePlan(svc.createJob)
	opts.AfterEnqueueApply(svc.createJob)
	// relay cancel signal from run service to agent.
	opts.AfterCancelSignal(svc.relaySignal(cancelSignal))
	// relay force-cancel signal from run service to agent.
	opts.AfterForceCancelSignal(svc.relaySignal(forceCancelSignal))
	return svc
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

func (s *service) AddHandlers(r *mux.Router) {
	s.tfeapi.addHandlers(r)
	s.api.addHandlers(r)
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
// getAgentJobs is intended to be called by an agent in order to receive jobs to
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
		switch job.Status {
		case JobAllocated:
			if *job.AgentID == agentID {
				return []*Job{job}, nil
			}
		case JobRunning:
			if *job.AgentID == agentID && job.signal != nil {
				return []*Job{job}, nil
			}
		}
	}
	return nil, nil
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
