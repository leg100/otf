package workspace

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/leg100/otf/internal/sql"
	"github.com/leg100/otf/internal/sql/sqlc"
)

// toggleLock toggles the workspace lock state in the DB.
func (db *pgdb) toggleLock(ctx context.Context, workspaceID string, togglefn func(*Workspace) error) (*Workspace, error) {
	var ws *Workspace
	err := db.Tx(ctx, func(ctx context.Context, q *sqlc.Queries) error {
		// retrieve workspace
		result, err := q.FindWorkspaceByIDForUpdate(ctx, sql.String(workspaceID))
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
			WorkspaceID: pgtype.Text{String: ws.ID},
		}
		if ws.Lock != nil {
			if ws.Lock.LockKind == RunLock {
				params.RunID = pgtype.Text{String: ws.Lock.id}
			} else if ws.Lock.LockKind == UserLock {
				params.Username = pgtype.Text{String: ws.Lock.id}
			} else {
				return ErrWorkspaceInvalidLock
			}
		}
		if err := q.UpdateWorkspaceLockByID(ctx, params); err != nil {
			return sql.Error(err)
		}
		return nil
	})
	return ws, err
}
