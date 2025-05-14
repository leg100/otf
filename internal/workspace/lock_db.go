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
		func(ctx context.Context) (*Workspace, error) {
			return db.forUpdate(ctx, workspaceID)
		},
		func(ctx context.Context, ws *Workspace) error {
			return togglefn(ws)
		},
		func(ctx context.Context, ws *Workspace) error {
			var (
				runID, username resource.ID
			)
			if ws.Locked() {
				switch ws.Lock.Kind() {
				case resource.RunKind:
					runID = ws.Lock
				case resource.UserKind:
					username = ws.Lock
				default:
					return ErrWorkspaceInvalidLock
				}
			}
			_, err := db.Exec(ctx, `
UPDATE workspaces
SET
    lock_username = $1,
    lock_run_id = $2
WHERE workspace_id = $3
`,
				username,
				runID,
				workspaceID,
			)
			return err
		})
}
