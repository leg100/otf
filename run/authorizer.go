package run

import (
	"context"

	"github.com/leg100/otf"
	"github.com/leg100/otf/rbac"
)

// authorizer authorizes access to a run
type authorizer struct {
	db        *pgdb
	workspace otf.Authorizer
}

func (a *authorizer) CanAccessRun(ctx context.Context, action rbac.Action, runID string) (otf.Subject, error) {
	run, err := a.db.GetRun(ctx, runID)
	if err != nil {
		return nil, err
	}
	return a.workspace.CanAccess(ctx, action, run.WorkspaceID)
}
