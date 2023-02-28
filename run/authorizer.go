package run

import (
	"context"

	"github.com/leg100/otf"
	"github.com/leg100/otf/rbac"
)

type workspaceAuthorizer interface {
	CanAccessWorkspaceByID(ctx context.Context, action rbac.Action, workspaceID string) (otf.Subject, error)
}

// authorizer authorizes access to a run
type authorizer struct {
	db *pgdb
	workspaceAuthorizer
}

func (a *authorizer) CanAccessRun(ctx context.Context, action rbac.Action, runID string) (otf.Subject, error) {
	run, err := a.db.GetRun(ctx, runID)
	if err != nil {
		return nil, err
	}
	return a.CanAccessWorkspaceByID(ctx, action, run.workspaceID)
}
