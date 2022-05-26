// Code generated by pggen. DO NOT EDIT.

package pggen

import (
	"context"
	"fmt"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
)

const insertWorkspaceLockUserSQL = `INSERT INTO workspace_locks (
    workspace_id,
    user_id
) VALUES (
    $1,
    $2
);`

// InsertWorkspaceLockUser implements Querier.InsertWorkspaceLockUser.
func (q *DBQuerier) InsertWorkspaceLockUser(ctx context.Context, workspaceID string, userID string) (pgconn.CommandTag, error) {
	ctx = context.WithValue(ctx, "pggen_query_name", "InsertWorkspaceLockUser")
	cmdTag, err := q.conn.Exec(ctx, insertWorkspaceLockUserSQL, workspaceID, userID)
	if err != nil {
		return cmdTag, fmt.Errorf("exec query InsertWorkspaceLockUser: %w", err)
	}
	return cmdTag, err
}

// InsertWorkspaceLockUserBatch implements Querier.InsertWorkspaceLockUserBatch.
func (q *DBQuerier) InsertWorkspaceLockUserBatch(batch genericBatch, workspaceID string, userID string) {
	batch.Queue(insertWorkspaceLockUserSQL, workspaceID, userID)
}

// InsertWorkspaceLockUserScan implements Querier.InsertWorkspaceLockUserScan.
func (q *DBQuerier) InsertWorkspaceLockUserScan(results pgx.BatchResults) (pgconn.CommandTag, error) {
	cmdTag, err := results.Exec()
	if err != nil {
		return cmdTag, fmt.Errorf("exec InsertWorkspaceLockUserBatch: %w", err)
	}
	return cmdTag, err
}

const insertWorkspaceLockRunSQL = `INSERT INTO workspace_locks (
    workspace_id,
    run_id
) VALUES (
    $1,
    $2
);`

// InsertWorkspaceLockRun implements Querier.InsertWorkspaceLockRun.
func (q *DBQuerier) InsertWorkspaceLockRun(ctx context.Context, workspaceID string, runID string) (pgconn.CommandTag, error) {
	ctx = context.WithValue(ctx, "pggen_query_name", "InsertWorkspaceLockRun")
	cmdTag, err := q.conn.Exec(ctx, insertWorkspaceLockRunSQL, workspaceID, runID)
	if err != nil {
		return cmdTag, fmt.Errorf("exec query InsertWorkspaceLockRun: %w", err)
	}
	return cmdTag, err
}

// InsertWorkspaceLockRunBatch implements Querier.InsertWorkspaceLockRunBatch.
func (q *DBQuerier) InsertWorkspaceLockRunBatch(batch genericBatch, workspaceID string, runID string) {
	batch.Queue(insertWorkspaceLockRunSQL, workspaceID, runID)
}

// InsertWorkspaceLockRunScan implements Querier.InsertWorkspaceLockRunScan.
func (q *DBQuerier) InsertWorkspaceLockRunScan(results pgx.BatchResults) (pgconn.CommandTag, error) {
	cmdTag, err := results.Exec()
	if err != nil {
		return cmdTag, fmt.Errorf("exec InsertWorkspaceLockRunBatch: %w", err)
	}
	return cmdTag, err
}

const findWorkspaceLockForUpdateSQL = `SELECT l.*
FROM workspace_locks l
LEFT JOIN users USING (user_id)
LEFT JOIN runs USING (run_id)
WHERE l.workspace_id = $1
FOR UPDATE of l;`

type FindWorkspaceLockForUpdateRow struct {
	WorkspaceID string `json:"workspace_id"`
	UserID      string `json:"user_id"`
	RunID       string `json:"run_id"`
}

// FindWorkspaceLockForUpdate implements Querier.FindWorkspaceLockForUpdate.
func (q *DBQuerier) FindWorkspaceLockForUpdate(ctx context.Context, workspaceID string) (FindWorkspaceLockForUpdateRow, error) {
	ctx = context.WithValue(ctx, "pggen_query_name", "FindWorkspaceLockForUpdate")
	row := q.conn.QueryRow(ctx, findWorkspaceLockForUpdateSQL, workspaceID)
	var item FindWorkspaceLockForUpdateRow
	if err := row.Scan(&item.WorkspaceID, &item.UserID, &item.RunID); err != nil {
		return item, fmt.Errorf("query FindWorkspaceLockForUpdate: %w", err)
	}
	return item, nil
}

// FindWorkspaceLockForUpdateBatch implements Querier.FindWorkspaceLockForUpdateBatch.
func (q *DBQuerier) FindWorkspaceLockForUpdateBatch(batch genericBatch, workspaceID string) {
	batch.Queue(findWorkspaceLockForUpdateSQL, workspaceID)
}

// FindWorkspaceLockForUpdateScan implements Querier.FindWorkspaceLockForUpdateScan.
func (q *DBQuerier) FindWorkspaceLockForUpdateScan(results pgx.BatchResults) (FindWorkspaceLockForUpdateRow, error) {
	row := results.QueryRow()
	var item FindWorkspaceLockForUpdateRow
	if err := row.Scan(&item.WorkspaceID, &item.UserID, &item.RunID); err != nil {
		return item, fmt.Errorf("scan FindWorkspaceLockForUpdateBatch row: %w", err)
	}
	return item, nil
}

const deleteWorkspaceLockSQL = `DELETE
FROM workspace_locks
WHERE workspace_id = $1
RETURNING workspace_id;`

// DeleteWorkspaceLock implements Querier.DeleteWorkspaceLock.
func (q *DBQuerier) DeleteWorkspaceLock(ctx context.Context, workspaceID string) (string, error) {
	ctx = context.WithValue(ctx, "pggen_query_name", "DeleteWorkspaceLock")
	row := q.conn.QueryRow(ctx, deleteWorkspaceLockSQL, workspaceID)
	var item string
	if err := row.Scan(&item); err != nil {
		return item, fmt.Errorf("query DeleteWorkspaceLock: %w", err)
	}
	return item, nil
}

// DeleteWorkspaceLockBatch implements Querier.DeleteWorkspaceLockBatch.
func (q *DBQuerier) DeleteWorkspaceLockBatch(batch genericBatch, workspaceID string) {
	batch.Queue(deleteWorkspaceLockSQL, workspaceID)
}

// DeleteWorkspaceLockScan implements Querier.DeleteWorkspaceLockScan.
func (q *DBQuerier) DeleteWorkspaceLockScan(results pgx.BatchResults) (string, error) {
	row := results.QueryRow()
	var item string
	if err := row.Scan(&item); err != nil {
		return item, fmt.Errorf("scan DeleteWorkspaceLockBatch row: %w", err)
	}
	return item, nil
}
