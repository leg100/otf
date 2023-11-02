package agent

import (
	"context"

	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/logr"
	"github.com/leg100/otf/internal/organization"
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

func (s *service) listPools(ctx context.Context, organization string, opts listPoolOptions) ([]*Pool, error) {
	subject, err := s.organization.CanAccess(ctx, rbac.CreateRunAction, organization)
	if err != nil {
		return nil, err
	}
	pools, err := s.db.listPools(ctx, organization, opts)
	if err != nil {
		s.Error(err, "listing agent pools", "subject", subject, "organization", organization)
		return nil, err
	}
	s.V(9).Info("listed agent pools", "subject", subject, "organization", organization, "count", len(pools))
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

//func (s *service) createJob(ctx context.Context, run *otfrun.Run, agentID string) (*Job, error) {
//	//return newJob(run, agentID), nil
//	return nil, nil
//}
//
//func (s *service) ping(ctx context.Context, agentID string) error {
//	return nil
//}
