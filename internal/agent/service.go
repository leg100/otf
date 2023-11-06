package agent

import (
	"context"

	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/logr"
	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/pubsub"
	"github.com/leg100/otf/internal/rbac"
	"github.com/leg100/otf/internal/sql"
	"github.com/leg100/otf/internal/sql/pggen"
	"github.com/leg100/otf/internal/tfeapi"
)

type (
	service struct {
		logr.Logger
		*db
		tfeapi *tfe
		*registrar
		// Subscriber for receiving stream of job and agent events
		pubsub.Subscriber

		organization internal.Authorizer
	}

	ServiceOptions struct {
		logr.Logger
		*sql.DB
		*tfeapi.Responder
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
	svc.registrar = &registrar{
		service: svc,
	}
	return svc
}

func (s *service) AddHandlers(r *mux.Router) {
	s.tfeapi.addHandlers(r)
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
		if err := s.db.createAgent(ctx, agent); err != nil {
			return nil, err
		}
		// TODO: reallocate jobs in db
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

//	func (s *service) createJob(ctx context.Context, run *otfrun.Run, agentID string) (*Job, error) {
//		//return newJob(run, agentID), nil
//		return nil, nil
//	}

func (s *service) ping(ctx context.Context, agentID string) error {
	return nil
}

func (s *service) getAllocatedJobs(ctx context.Context, agentID string) ([]*Job, error) {
	// 1. subscribe to pubsub for jobs newly allocated to agent
	// 2. get jobs from db allocated to agent
	sub, err := s.Subscribe(ctx, "get-allocated-jobs-"+agentID)
	if err != nil {
		return nil, err
	}
	allocated, err := s.db.getAllocatedJobs(ctx, agentID)
	if err != nil {
		return nil, err
	}
	if len(allocated) > 0 {
		return allocated, nil
	}
	for event := range sub {
		if job, ok := event.Payload.(*Job); ok {
			return []*Job{job}, nil
		}
	}
	return nil, nil
}

func (s *service) listJobs(ctx context.Context) ([]*Job, error) {
	return nil, nil
}

func (s *service) reallocateJob(ctx context.Context, spec JobSpec) error {
	return nil
}

func (s *service) updateJobStatus(ctx context.Context, jobID string, status JobStatus) error {
	return nil
}

func (s *service) allocateJob(ctx context.Context, job *Job) error {
	// TODO: perform DB update of job
	return nil
}
