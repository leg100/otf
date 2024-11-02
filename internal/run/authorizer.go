package run

import (
	"context"

	"github.com/leg100/otf/internal/authz"
	"github.com/leg100/otf/internal/rbac"
)

// authorizer authorizes access to a run
type authorizer struct {
	db        *pgdb
	workspace authz.Authorizer
}

func (a *authorizer) CanAccess(ctx context.Context, action rbac.Action, runID string) (authz.Subject, error) {
	run, err := a.db.GetRun(ctx, runID)
	if err != nil {
		return nil, err
	}
	return a.workspace.CanAccess(ctx, action, run.WorkspaceID)
}
