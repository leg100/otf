package sql

import (
	"context"

	"github.com/jackc/pgx/v4"
)

type workspaceUpdater struct {
	q  *DBQuerier
	id string
}

func newWorkspaceUpdater(tx pgx.Tx, id string) *workspaceUpdater {
	return &workspaceUpdater{
		q:  NewQuerier(tx),
		id: id,
	}
}

func (u *workspaceUpdater) ToggleLock(ctx context.Context, lock bool) error {
	_, err := u.q.UpdateWorkspaceLockByID(ctx, &lock, &u.id)
	return err
}

func (u *workspaceUpdater) UpdateName(ctx context.Context, name string) error {
	_, err := u.q.UpdateWorkspaceNameByID(ctx, &name, &u.id)
	return err
}

func (u *workspaceUpdater) UpdateAllowDestroyPlan(ctx context.Context, allow bool) error {
	_, err := u.q.UpdateWorkspaceAllowDestroyPlanByID(ctx, &allow, &u.id)
	return err
}
