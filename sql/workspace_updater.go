package sql

import (
	"context"

	"github.com/jackc/pgx/v4"
)

type workspaceUpdater struct {
	q      *DBQuerier
	id     string
	result workspaceComposite
}

func newWorkspaceUpdater(tx pgx.Tx, id string) *workspaceUpdater {
	return &workspaceUpdater{
		q:  NewQuerier(tx),
		id: id,
	}
}

func (u *workspaceUpdater) ToggleLock(ctx context.Context, lock bool) (err error) {
	u.result, err = u.q.UpdateWorkspaceLockByID(ctx, &lock, &u.id)
	return err
}

func (u *workspaceUpdater) UpdateName(ctx context.Context, name string) (err error) {
	u.result, err = u.q.UpdateWorkspaceNameByID(ctx, &name, &u.id)
	return err
}

func (u *workspaceUpdater) UpdateAllowDestroyPlan(ctx context.Context, allow bool) (err error) {
	u.result, err = u.q.UpdateWorkspaceAllowDestroyPlanByID(ctx, &allow, &u.id)
	return err
}
