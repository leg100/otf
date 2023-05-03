package policy

import (
	"context"

	"github.com/leg100/otf"
	"github.com/leg100/otf/logr"
	"github.com/leg100/otf/rbac"
)

type (
	PolicyService = Service

	Service interface {
		GetPolicy(ctx context.Context, workspaceID string) (otf.WorkspacePolicy, error)
		SetPermission(ctx context.Context, workspaceID, team string, role rbac.Role) error
		UnsetPermission(ctx context.Context, workspaceID, team string) error

		// PolicyTx wraps the policy service with the given transaction.
		PolicyTx(ctx context.Context, tx otf.DB) Service

		// Implements authorizer, policing access to workspaces
		otf.Authorizer
	}

	service struct {
		logr.Logger
		db *pgdb
	}

	Options struct {
		otf.DB
		logr.Logger
	}
)

func NewService(opts Options) *service {
	return &service{
		Logger: opts.Logger,
		db:     &pgdb{opts.DB},
	}
}

func (s *service) PolicyTx(ctx context.Context, tx otf.DB) Service {
	ss := *s
	ss.db = &pgdb{tx}
	return &ss
}

// GetPolicy retrieves a workspace policy.
//
// NOTE: no authz protects this endpoint because it's used in the process of making
// authz decisions.
func (s *service) GetPolicy(ctx context.Context, workspaceID string) (otf.WorkspacePolicy, error) {
	return s.db.getWorkspacePolicy(ctx, workspaceID)
}

func (s *service) SetPermission(ctx context.Context, workspaceID, team string, role rbac.Role) error {
	subject, err := s.CanAccess(ctx, rbac.SetWorkspacePermissionAction, workspaceID)
	if err != nil {
		return err
	}

	if err := s.db.setWorkspacePermission(ctx, workspaceID, team, role); err != nil {
		s.Error(err, "setting workspace permission", "subject", subject, "workspace", workspaceID)
		return err
	}

	s.V(0).Info("set workspace permission", "team", team, "role", role, "subject", subject, "workspace", workspaceID)

	// TODO: publish event

	return nil
}

func (s *service) UnsetPermission(ctx context.Context, workspaceID, team string) error {
	subject, err := s.CanAccess(ctx, rbac.UnsetWorkspacePermissionAction, workspaceID)
	if err != nil {
		s.Error(err, "unsetting workspace permission", "team", team, "subject", subject, "workspace", workspaceID)
		return err
	}

	s.V(0).Info("unset workspace permission", "team", team, "subject", subject, "workspace", workspaceID)
	// TODO: publish event
	return s.db.unsetWorkspacePermission(ctx, workspaceID, team)
}

// CanAccess determines whether the subject (in the ctx) is permitted to carry
// out the specified action on the workspace with the given id.
func (s *service) CanAccess(ctx context.Context, action rbac.Action, workspaceID string) (otf.Subject, error) {
	subj, err := otf.SubjectFromContext(ctx)
	if err != nil {
		return nil, err
	}
	policy, err := s.db.getWorkspacePolicy(ctx, workspaceID)
	if err != nil {
		return nil, otf.ErrResourceNotFound
	}
	if subj.CanAccessWorkspace(action, policy) {
		return subj, nil
	}
	s.Error(nil, "unauthorized action", "workspace", workspaceID, "organization", policy.Organization, "action", action, "subject", subj)
	return nil, otf.ErrAccessNotPermitted
}
