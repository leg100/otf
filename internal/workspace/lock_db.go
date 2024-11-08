package workspace

import (
	"context"

	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/sql"
	"github.com/leg100/otf/internal/sql/sqlc"
)

// toggleLock toggles the workspace lock state in the DB.
func (db *pgdb) toggleLock(ctx context.Context, workspaceID resource.ID, togglefn func(*Workspace) error) (*Workspace, error) {
	var ws *Workspace
	err := db.Tx(ctx, func(ctx context.Context, q *sqlc.Queries) error {
		// retrieve workspace
		result, err := q.FindWorkspaceByIDForUpdate(ctx, sql.ID(workspaceID))
		if err != nil {
			return err
		}
		ws, err = pgresult(result).toWorkspace()
		if err != nil {
			return err
		}
		if err := togglefn(ws); err != nil {
			return err
		}
		// persist to db
		params := sqlc.UpdateWorkspaceLockByIDParams{
			WorkspaceID: sql.ID(ws.ID),
		}
		if ws.Locked() {
			switch ws.Lock.Kind {
			case resource.RunKind:
				params.RunID = sql.IDPtr(ws.Lock)
			case resource.UserKind:
				params.UserID = sql.IDPtr(ws.Lock)
			default:
				return ErrWorkspaceInvalidLock
			}
		}
		if err := q.UpdateWorkspaceLockByID(ctx, params); err != nil {
			return err
		}
		return nil
	})
	return ws, sql.Error(err)
}
