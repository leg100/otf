package workspace

import (
	"context"
	"fmt"

	"github.com/jackc/pgtype"
	"github.com/leg100/otf"
	"github.com/leg100/otf/sql"
	"github.com/leg100/otf/sql/pggen"
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
		ws, err = UnmarshalWorkspaceResult(WorkspaceResult(result))
		if err != nil {
			return err
		}
		if err := togglefn(ws); err != nil {
			return err
		}
		// persist to db
		params, err := MarshalWorkspaceLockParams(ws)
		if err != nil {
			return err
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

func MarshalWorkspaceLockParams(ws *Workspace) (pggen.UpdateWorkspaceLockByIDParams, error) {
	params := pggen.UpdateWorkspaceLockByIDParams{
		WorkspaceID: pgtype.Text{String: ws.ID(), Status: pgtype.Present},
	}
	switch state := ws.state.(type) {
	case RunLock:
		params.RunID = pgtype.Text{String: state.ID(), Status: pgtype.Present}
		params.UserID = pgtype.Text{Status: pgtype.Null}
	case UserLock:
		params.UserID = pgtype.Text{String: state.ID(), Status: pgtype.Present}
		params.RunID = pgtype.Text{Status: pgtype.Null}
	case nil:
		params.RunID = pgtype.Text{Status: pgtype.Null}
		params.UserID = pgtype.Text{Status: pgtype.Null}
	default:
		return params, otf.ErrWorkspaceInvalidLock
	}
	return params, nil
}

func unmarshalWorkspaceLock(dst *Workspace, row *WorkspaceResult) error {
	if row.UserLock == nil && row.RunLock == nil {
		dst.state = nil
	} else if row.UserLock != nil {
		dst.state = UserLock{id: row.UserLock.UserID.String, username: row.UserLock.Username.String}
	} else if row.RunLock != nil {
		dst.state = RunLock{id: row.RunLock.RunID.String}
	} else {
		return fmt.Errorf("workspace cannot be locked by both a run and a user")
	}
	return nil
}