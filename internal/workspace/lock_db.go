package workspace

import (
	"context"

	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/sql"
)

// toggleLock toggles the workspace lock state in the DB.
func (db *pgdb) toggleLock(ctx context.Context, workspaceID resource.TfeID, togglefn func(*Workspace) error) (*Workspace, error) {
	var ws *Workspace
	err := db.Tx(ctx, func(ctx context.Context, conn sql.Connection) error {
		// retrieve workspace
		result, err := q.FindWorkspaceByIDForUpdate(ctx, conn, workspaceID)
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
		params := UpdateWorkspaceLockByIDParams{
			WorkspaceID: ws.ID,
		}
		if ws.Locked() {
			switch ws.Lock.Kind() {
			case resource.RunKind:
				params.RunID = ws.Lock
			case resource.UserKind:
				params.UserID = ws.Lock
			default:
				return ErrWorkspaceInvalidLock
			}
		}
		if err := q.UpdateWorkspaceLockByID(ctx, conn, params); err != nil {
			return err
		}
		return nil
	})
	return ws, sql.Error(err)
}
