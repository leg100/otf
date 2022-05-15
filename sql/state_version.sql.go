// Code generated by pggen. DO NOT EDIT.

package sql

import (
	"context"
	"fmt"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"time"
)

const insertStateVersionSQL = `INSERT INTO state_versions (
    state_version_id,
    created_at,
    updated_at,
    serial,
    state,
    run_id
) VALUES (
    $1,
    current_timestamp,
    current_timestamp,
    $2,
    $3,
    $4
)
RETURNING created_at, updated_at
;`

type InsertStateVersionParams struct {
	ID     string
	Serial int32
	State  []byte
	RunID  string
}

type InsertStateVersionRow struct {
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// InsertStateVersion implements Querier.InsertStateVersion.
func (q *DBQuerier) InsertStateVersion(ctx context.Context, params InsertStateVersionParams) (InsertStateVersionRow, error) {
	ctx = context.WithValue(ctx, "pggen_query_name", "InsertStateVersion")
	row := q.conn.QueryRow(ctx, insertStateVersionSQL, params.ID, params.Serial, params.State, params.RunID)
	var item InsertStateVersionRow
	if err := row.Scan(&item.CreatedAt, &item.UpdatedAt); err != nil {
		return item, fmt.Errorf("query InsertStateVersion: %w", err)
	}
	return item, nil
}

// InsertStateVersionBatch implements Querier.InsertStateVersionBatch.
func (q *DBQuerier) InsertStateVersionBatch(batch genericBatch, params InsertStateVersionParams) {
	batch.Queue(insertStateVersionSQL, params.ID, params.Serial, params.State, params.RunID)
}

// InsertStateVersionScan implements Querier.InsertStateVersionScan.
func (q *DBQuerier) InsertStateVersionScan(results pgx.BatchResults) (InsertStateVersionRow, error) {
	row := results.QueryRow()
	var item InsertStateVersionRow
	if err := row.Scan(&item.CreatedAt, &item.UpdatedAt); err != nil {
		return item, fmt.Errorf("scan InsertStateVersionBatch row: %w", err)
	}
	return item, nil
}

const findStateVersionsByWorkspaceNameSQL = `SELECT
    state_versions.*,
    array_remove(array_agg(state_version_outputs), NULL) AS state_version_outputs,
    count(*) OVER() AS full_count
FROM state_versions
JOIN (runs JOIN workspaces USING (workspace_id)) USING (run_id)
JOIN organizations USING (organization_id)
LEFT JOIN state_version_outputs USING (state_version_id)
WHERE workspaces.name = $1
AND organizations.name = $2
GROUP BY state_versions.state_version_id
LIMIT $3
OFFSET $4
;`

type FindStateVersionsByWorkspaceNameParams struct {
	WorkspaceName    string
	OrganizationName string
	Limit            int
	Offset           int
}

type FindStateVersionsByWorkspaceNameRow struct {
	StateVersionID      *string               `json:"state_version_id"`
	CreatedAt           time.Time             `json:"created_at"`
	UpdatedAt           time.Time             `json:"updated_at"`
	Serial              *int32                `json:"serial"`
	VcsCommitSha        *string               `json:"vcs_commit_sha"`
	VcsCommitUrl        *string               `json:"vcs_commit_url"`
	State               []byte                `json:"state"`
	RunID               *string               `json:"run_id"`
	StateVersionOutputs []StateVersionOutputs `json:"state_version_outputs"`
	FullCount           *int                  `json:"full_count"`
}

// FindStateVersionsByWorkspaceName implements Querier.FindStateVersionsByWorkspaceName.
func (q *DBQuerier) FindStateVersionsByWorkspaceName(ctx context.Context, params FindStateVersionsByWorkspaceNameParams) ([]FindStateVersionsByWorkspaceNameRow, error) {
	ctx = context.WithValue(ctx, "pggen_query_name", "FindStateVersionsByWorkspaceName")
	rows, err := q.conn.Query(ctx, findStateVersionsByWorkspaceNameSQL, params.WorkspaceName, params.OrganizationName, params.Limit, params.Offset)
	if err != nil {
		return nil, fmt.Errorf("query FindStateVersionsByWorkspaceName: %w", err)
	}
	defer rows.Close()
	items := []FindStateVersionsByWorkspaceNameRow{}
	stateVersionOutputsArray := q.types.newStateVersionOutputsArray()
	for rows.Next() {
		var item FindStateVersionsByWorkspaceNameRow
		if err := rows.Scan(&item.StateVersionID, &item.CreatedAt, &item.UpdatedAt, &item.Serial, &item.VcsCommitSha, &item.VcsCommitUrl, &item.State, &item.RunID, stateVersionOutputsArray, &item.FullCount); err != nil {
			return nil, fmt.Errorf("scan FindStateVersionsByWorkspaceName row: %w", err)
		}
		if err := stateVersionOutputsArray.AssignTo(&item.StateVersionOutputs); err != nil {
			return nil, fmt.Errorf("assign FindStateVersionsByWorkspaceName row: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("close FindStateVersionsByWorkspaceName rows: %w", err)
	}
	return items, err
}

// FindStateVersionsByWorkspaceNameBatch implements Querier.FindStateVersionsByWorkspaceNameBatch.
func (q *DBQuerier) FindStateVersionsByWorkspaceNameBatch(batch genericBatch, params FindStateVersionsByWorkspaceNameParams) {
	batch.Queue(findStateVersionsByWorkspaceNameSQL, params.WorkspaceName, params.OrganizationName, params.Limit, params.Offset)
}

// FindStateVersionsByWorkspaceNameScan implements Querier.FindStateVersionsByWorkspaceNameScan.
func (q *DBQuerier) FindStateVersionsByWorkspaceNameScan(results pgx.BatchResults) ([]FindStateVersionsByWorkspaceNameRow, error) {
	rows, err := results.Query()
	if err != nil {
		return nil, fmt.Errorf("query FindStateVersionsByWorkspaceNameBatch: %w", err)
	}
	defer rows.Close()
	items := []FindStateVersionsByWorkspaceNameRow{}
	stateVersionOutputsArray := q.types.newStateVersionOutputsArray()
	for rows.Next() {
		var item FindStateVersionsByWorkspaceNameRow
		if err := rows.Scan(&item.StateVersionID, &item.CreatedAt, &item.UpdatedAt, &item.Serial, &item.VcsCommitSha, &item.VcsCommitUrl, &item.State, &item.RunID, stateVersionOutputsArray, &item.FullCount); err != nil {
			return nil, fmt.Errorf("scan FindStateVersionsByWorkspaceNameBatch row: %w", err)
		}
		if err := stateVersionOutputsArray.AssignTo(&item.StateVersionOutputs); err != nil {
			return nil, fmt.Errorf("assign FindStateVersionsByWorkspaceName row: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("close FindStateVersionsByWorkspaceNameBatch rows: %w", err)
	}
	return items, err
}

const findStateVersionByIDSQL = `SELECT
    state_versions.*,
    array_remove(array_agg(state_version_outputs), NULL) AS state_version_outputs
FROM state_versions
JOIN (runs JOIN workspaces USING (workspace_id)) USING (run_id)
LEFT JOIN state_version_outputs USING (state_version_id)
WHERE state_versions.state_version_id = $1
GROUP BY state_versions.state_version_id
;`

type FindStateVersionByIDRow struct {
	StateVersionID      *string               `json:"state_version_id"`
	CreatedAt           time.Time             `json:"created_at"`
	UpdatedAt           time.Time             `json:"updated_at"`
	Serial              *int32                `json:"serial"`
	VcsCommitSha        *string               `json:"vcs_commit_sha"`
	VcsCommitUrl        *string               `json:"vcs_commit_url"`
	State               []byte                `json:"state"`
	RunID               *string               `json:"run_id"`
	StateVersionOutputs []StateVersionOutputs `json:"state_version_outputs"`
}

// FindStateVersionByID implements Querier.FindStateVersionByID.
func (q *DBQuerier) FindStateVersionByID(ctx context.Context, id string) (FindStateVersionByIDRow, error) {
	ctx = context.WithValue(ctx, "pggen_query_name", "FindStateVersionByID")
	row := q.conn.QueryRow(ctx, findStateVersionByIDSQL, id)
	var item FindStateVersionByIDRow
	stateVersionOutputsArray := q.types.newStateVersionOutputsArray()
	if err := row.Scan(&item.StateVersionID, &item.CreatedAt, &item.UpdatedAt, &item.Serial, &item.VcsCommitSha, &item.VcsCommitUrl, &item.State, &item.RunID, stateVersionOutputsArray); err != nil {
		return item, fmt.Errorf("query FindStateVersionByID: %w", err)
	}
	if err := stateVersionOutputsArray.AssignTo(&item.StateVersionOutputs); err != nil {
		return item, fmt.Errorf("assign FindStateVersionByID row: %w", err)
	}
	return item, nil
}

// FindStateVersionByIDBatch implements Querier.FindStateVersionByIDBatch.
func (q *DBQuerier) FindStateVersionByIDBatch(batch genericBatch, id string) {
	batch.Queue(findStateVersionByIDSQL, id)
}

// FindStateVersionByIDScan implements Querier.FindStateVersionByIDScan.
func (q *DBQuerier) FindStateVersionByIDScan(results pgx.BatchResults) (FindStateVersionByIDRow, error) {
	row := results.QueryRow()
	var item FindStateVersionByIDRow
	stateVersionOutputsArray := q.types.newStateVersionOutputsArray()
	if err := row.Scan(&item.StateVersionID, &item.CreatedAt, &item.UpdatedAt, &item.Serial, &item.VcsCommitSha, &item.VcsCommitUrl, &item.State, &item.RunID, stateVersionOutputsArray); err != nil {
		return item, fmt.Errorf("scan FindStateVersionByIDBatch row: %w", err)
	}
	if err := stateVersionOutputsArray.AssignTo(&item.StateVersionOutputs); err != nil {
		return item, fmt.Errorf("assign FindStateVersionByID row: %w", err)
	}
	return item, nil
}

const findStateVersionLatestByWorkspaceIDSQL = `SELECT
    state_versions.*,
    array_remove(array_agg(state_version_outputs), NULL) AS state_version_outputs
FROM state_versions
JOIN (runs JOIN workspaces USING (workspace_id)) USING (run_id)
LEFT JOIN state_version_outputs USING (state_version_id)
WHERE workspaces.workspace_id = $1
GROUP BY state_versions.state_version_id
ORDER BY state_versions.serial DESC, state_versions.created_at DESC
;`

type FindStateVersionLatestByWorkspaceIDRow struct {
	StateVersionID      *string               `json:"state_version_id"`
	CreatedAt           time.Time             `json:"created_at"`
	UpdatedAt           time.Time             `json:"updated_at"`
	Serial              *int32                `json:"serial"`
	VcsCommitSha        *string               `json:"vcs_commit_sha"`
	VcsCommitUrl        *string               `json:"vcs_commit_url"`
	State               []byte                `json:"state"`
	RunID               *string               `json:"run_id"`
	StateVersionOutputs []StateVersionOutputs `json:"state_version_outputs"`
}

// FindStateVersionLatestByWorkspaceID implements Querier.FindStateVersionLatestByWorkspaceID.
func (q *DBQuerier) FindStateVersionLatestByWorkspaceID(ctx context.Context, workspaceID string) (FindStateVersionLatestByWorkspaceIDRow, error) {
	ctx = context.WithValue(ctx, "pggen_query_name", "FindStateVersionLatestByWorkspaceID")
	row := q.conn.QueryRow(ctx, findStateVersionLatestByWorkspaceIDSQL, workspaceID)
	var item FindStateVersionLatestByWorkspaceIDRow
	stateVersionOutputsArray := q.types.newStateVersionOutputsArray()
	if err := row.Scan(&item.StateVersionID, &item.CreatedAt, &item.UpdatedAt, &item.Serial, &item.VcsCommitSha, &item.VcsCommitUrl, &item.State, &item.RunID, stateVersionOutputsArray); err != nil {
		return item, fmt.Errorf("query FindStateVersionLatestByWorkspaceID: %w", err)
	}
	if err := stateVersionOutputsArray.AssignTo(&item.StateVersionOutputs); err != nil {
		return item, fmt.Errorf("assign FindStateVersionLatestByWorkspaceID row: %w", err)
	}
	return item, nil
}

// FindStateVersionLatestByWorkspaceIDBatch implements Querier.FindStateVersionLatestByWorkspaceIDBatch.
func (q *DBQuerier) FindStateVersionLatestByWorkspaceIDBatch(batch genericBatch, workspaceID string) {
	batch.Queue(findStateVersionLatestByWorkspaceIDSQL, workspaceID)
}

// FindStateVersionLatestByWorkspaceIDScan implements Querier.FindStateVersionLatestByWorkspaceIDScan.
func (q *DBQuerier) FindStateVersionLatestByWorkspaceIDScan(results pgx.BatchResults) (FindStateVersionLatestByWorkspaceIDRow, error) {
	row := results.QueryRow()
	var item FindStateVersionLatestByWorkspaceIDRow
	stateVersionOutputsArray := q.types.newStateVersionOutputsArray()
	if err := row.Scan(&item.StateVersionID, &item.CreatedAt, &item.UpdatedAt, &item.Serial, &item.VcsCommitSha, &item.VcsCommitUrl, &item.State, &item.RunID, stateVersionOutputsArray); err != nil {
		return item, fmt.Errorf("scan FindStateVersionLatestByWorkspaceIDBatch row: %w", err)
	}
	if err := stateVersionOutputsArray.AssignTo(&item.StateVersionOutputs); err != nil {
		return item, fmt.Errorf("assign FindStateVersionLatestByWorkspaceID row: %w", err)
	}
	return item, nil
}

const deleteStateVersionByIDSQL = `DELETE
FROM state_versions
WHERE state_version_id = $1
;`

// DeleteStateVersionByID implements Querier.DeleteStateVersionByID.
func (q *DBQuerier) DeleteStateVersionByID(ctx context.Context, stateVersionID string) (pgconn.CommandTag, error) {
	ctx = context.WithValue(ctx, "pggen_query_name", "DeleteStateVersionByID")
	cmdTag, err := q.conn.Exec(ctx, deleteStateVersionByIDSQL, stateVersionID)
	if err != nil {
		return cmdTag, fmt.Errorf("exec query DeleteStateVersionByID: %w", err)
	}
	return cmdTag, err
}

// DeleteStateVersionByIDBatch implements Querier.DeleteStateVersionByIDBatch.
func (q *DBQuerier) DeleteStateVersionByIDBatch(batch genericBatch, stateVersionID string) {
	batch.Queue(deleteStateVersionByIDSQL, stateVersionID)
}

// DeleteStateVersionByIDScan implements Querier.DeleteStateVersionByIDScan.
func (q *DBQuerier) DeleteStateVersionByIDScan(results pgx.BatchResults) (pgconn.CommandTag, error) {
	cmdTag, err := results.Exec()
	if err != nil {
		return cmdTag, fmt.Errorf("exec DeleteStateVersionByIDBatch: %w", err)
	}
	return cmdTag, err
}
