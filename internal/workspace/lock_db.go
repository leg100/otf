package workspace

import (
	"context"

	"github.com/jackc/pgtype"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/sql"
	"github.com/leg100/otf/internal/sql/pggen"
)

// toggleLock toggles the workspace lock state in the DB.
func (db *pgdb) toggleLock(ctx context.Context, workspaceID string, togglefn func(*Workspace) error) (*Workspace, error) {
	var ws *Workspace
	err := db.tx(ctx, func(tx *pgdb) error {
		// retrieve workspace
		result, err := tx.FindWorkspaceByIDForUpdate(ctx, sql.String(workspaceID))
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
		params := pggen.UpdateWorkspaceLockByIDParams{
			WorkspaceID: pgtype.Text{String: ws.ID, Status: pgtype.Present},
		}
		if ws.lock == nil {
			params.RunID = pgtype.Text{Status: pgtype.Null}
			params.Username = pgtype.Text{Status: pgtype.Null}
		} else if ws.LockKind == RunLock {
			params.RunID = pgtype.Text{String: ws.lock.id, Status: pgtype.Present}
			params.Username = pgtype.Text{Status: pgtype.Null}
		} else if ws.LockKind == UserLock {
			params.Username = pgtype.Text{String: ws.lock.id, Status: pgtype.Present}
			params.RunID = pgtype.Text{Status: pgtype.Null}
		} else {
			return internal.ErrWorkspaceInvalidLock
		}
		_, err = tx.UpdateWorkspaceLockByID(ctx, params)
		if err != nil {
			return sql.Error(err)
		}
		return nil
	})
	// return ws with new lock
	return ws, err
}
