package run

import (
	"context"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/rbac"
	"github.com/leg100/otf/internal/resource"
)

// authorizer authorizes access to a run
type authorizer struct {
	db        *pgdb
	workspace internal.Authorizer
}

func (a *authorizer) CanAccess(ctx context.Context, action rbac.Action, runID resource.ID) (internal.Subject, error) {
	run, err := a.db.GetRun(ctx, runID)
	if err != nil {
		return nil, err
	}
	return a.workspace.CanAccess(ctx, action, run.WorkspaceID)
}
