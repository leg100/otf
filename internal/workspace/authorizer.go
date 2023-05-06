package workspace

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/leg100/otf"
	"github.com/leg100/otf/rbac"
)

// authorizer authorizes access to a workspace
type authorizer struct {
	logr.Logger

	db *pgdb
}

func (a *authorizer) CanAccess(ctx context.Context, action rbac.Action, workspaceID string) (otf.Subject, error) {
	subj, err := otf.SubjectFromContext(ctx)
	if err != nil {
		return nil, err
	}
	policy, err := a.db.GetWorkspacePolicy(ctx, workspaceID)
	if err != nil {
		return nil, otf.ErrResourceNotFound
	}
	if subj.CanAccessWorkspace(action, policy) {
		return subj, nil
	}
	a.Error(nil, "unauthorized action", "workspace", workspaceID, "organization", policy.Organization, "action", action, "subject", subj)
	return nil, otf.ErrAccessNotPermitted
}
