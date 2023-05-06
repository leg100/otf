package run

import (
	"context"

	internal "github.com/leg100/otf"
	"github.com/leg100/otf/rbac"
)

// authorizer authorizes access to a run
type authorizer struct {
	db        *pgdb
	workspace internal.Authorizer
}

func (a *authorizer) CanAccess(ctx context.Context, action rbac.Action, runID string) (internal.Subject, error) {
	run, err := a.db.GetRun(ctx, runID)
	if err != nil {
		return nil, err
	}
	return a.workspace.CanAccess(ctx, action, run.WorkspaceID)
}
