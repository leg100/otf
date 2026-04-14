package trigger

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/sql"
)

type pgdb struct {
	*sql.DB
}

func (db *pgdb) create(ctx context.Context, trigger *trigger) error {
	_, err := db.Exec(ctx, `
INSERT INTO run_triggers (
    run_id,
    created_at,
    workspace_id,
    sourceable_workspace_id
) VALUES (
    $1,
    $2,
    $3,
    $4
)`,
		trigger.ID,
		trigger.CreatedAt,
		trigger.WorkspaceID,
		trigger.SourceableWorkspaceID,
	)
	if err != nil {
		return err
	}
	return nil
}

func (db *pgdb) listByWorkspaceID(ctx context.Context, workspaceID resource.TfeID) ([]*trigger, error) {
	rows := db.Query(ctx, `
SELECT *
FROM run_triggers
WHERE run_triggers.workspace_id = $1
`, workspaceID)
	return sql.CollectRows(rows, pgx.RowToAddrOfStructByName[trigger])
}

func (db *pgdb) listBySourceableWorkspaceID(ctx context.Context, sourceableWorkspaceID resource.TfeID) ([]*trigger, error) {
	rows := db.Query(ctx, `
SELECT *
FROM run_triggers
WHERE run_triggers.sourceable_workspace_id = $1
`, sourceableWorkspaceID)
	return sql.CollectRows(rows, pgx.RowToAddrOfStructByName[trigger])
}

func (db *pgdb) get(ctx context.Context, triggerID resource.ID) (*trigger, error) {
	row := db.Query(ctx, `
SELECT *
FROM run_triggers
WHERE run_trigger_id = $1
`, triggerID)
	return sql.CollectExactlyOneRow(row, pgx.RowToAddrOfStructByName[trigger])
}

func (db *pgdb) delete(ctx context.Context, triggerID resource.ID) error {
	_, err := db.Exec(ctx, `
DELETE
FROM run_triggers
WHERE run_trigger_id = $1
`, triggerID)
	return err
}
