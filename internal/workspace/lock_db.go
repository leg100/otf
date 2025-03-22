package workspace

import (
	"context"

	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/sql"
)

// toggleLock toggles the workspace lock state in the DB.
func (db *pgdb) toggleLock(ctx context.Context, workspaceID resource.TfeID, togglefn func(*Workspace) error) (*Workspace, error) {
	return sql.Updater(
		ctx,
		db.DB,
		func(ctx context.Context, conn sql.Connection) (*Workspace, error) {
			return db.forUpdate(ctx, conn, workspaceID)
		},
		func(ctx context.Context, ws *Workspace) error {
			return togglefn(ws)
		},
		func(ctx context.Context, conn sql.Connection, ws *Workspace) error {
			var (
				runID  *resource.TfeID
				userID *resource.TfeID
			)
			if ws.Locked() {
				switch ws.Lock.Kind() {
				case resource.RunKind:
					runID = ws.Lock
				case resource.UserKind:
					userID = ws.Lock
				default:
					return ErrWorkspaceInvalidLock
				}
			}
			_, err := db.Conn(ctx).Exec(ctx, `
UPDATE workspaces
SET
    lock_user_id = $1,
    lock_run_id = $2
WHERE workspace_id = $3
`,
				userID,
				runID,
				workspaceID,
			)
			return err
		})
}
