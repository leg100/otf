// Code generated by pggen. DO NOT EDIT.

package pggen

import (
	"context"
	"fmt"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
)

const insertLatestTerraformVersionSQL = `INSERT INTO latest_terraform_version (
    version,
    checkpoint
) VALUES (
    $1,
    current_timestamp
);`

// InsertLatestTerraformVersion implements Querier.InsertLatestTerraformVersion.
func (q *DBQuerier) InsertLatestTerraformVersion(ctx context.Context, version pgtype.Text) (pgconn.CommandTag, error) {
	ctx = context.WithValue(ctx, "pggen_query_name", "InsertLatestTerraformVersion")
	cmdTag, err := q.conn.Exec(ctx, insertLatestTerraformVersionSQL, version)
	if err != nil {
		return cmdTag, fmt.Errorf("exec query InsertLatestTerraformVersion: %w", err)
	}
	return cmdTag, err
}

// InsertLatestTerraformVersionBatch implements Querier.InsertLatestTerraformVersionBatch.
func (q *DBQuerier) InsertLatestTerraformVersionBatch(batch genericBatch, version pgtype.Text) {
	batch.Queue(insertLatestTerraformVersionSQL, version)
}

// InsertLatestTerraformVersionScan implements Querier.InsertLatestTerraformVersionScan.
func (q *DBQuerier) InsertLatestTerraformVersionScan(results pgx.BatchResults) (pgconn.CommandTag, error) {
	cmdTag, err := results.Exec()
	if err != nil {
		return cmdTag, fmt.Errorf("exec InsertLatestTerraformVersionBatch: %w", err)
	}
	return cmdTag, err
}

const updateLatestTerraformVersionSQL = `UPDATE latest_terraform_version
SET version = $1,
    checkpoint = current_timestamp;`

// UpdateLatestTerraformVersion implements Querier.UpdateLatestTerraformVersion.
func (q *DBQuerier) UpdateLatestTerraformVersion(ctx context.Context, version pgtype.Text) (pgconn.CommandTag, error) {
	ctx = context.WithValue(ctx, "pggen_query_name", "UpdateLatestTerraformVersion")
	cmdTag, err := q.conn.Exec(ctx, updateLatestTerraformVersionSQL, version)
	if err != nil {
		return cmdTag, fmt.Errorf("exec query UpdateLatestTerraformVersion: %w", err)
	}
	return cmdTag, err
}

// UpdateLatestTerraformVersionBatch implements Querier.UpdateLatestTerraformVersionBatch.
func (q *DBQuerier) UpdateLatestTerraformVersionBatch(batch genericBatch, version pgtype.Text) {
	batch.Queue(updateLatestTerraformVersionSQL, version)
}

// UpdateLatestTerraformVersionScan implements Querier.UpdateLatestTerraformVersionScan.
func (q *DBQuerier) UpdateLatestTerraformVersionScan(results pgx.BatchResults) (pgconn.CommandTag, error) {
	cmdTag, err := results.Exec()
	if err != nil {
		return cmdTag, fmt.Errorf("exec UpdateLatestTerraformVersionBatch: %w", err)
	}
	return cmdTag, err
}

const findLatestTerraformVersionSQL = `SELECT *
FROM latest_terraform_version;`

type FindLatestTerraformVersionRow struct {
	Version    pgtype.Text        `json:"version"`
	Checkpoint pgtype.Timestamptz `json:"checkpoint"`
}

// FindLatestTerraformVersion implements Querier.FindLatestTerraformVersion.
func (q *DBQuerier) FindLatestTerraformVersion(ctx context.Context) ([]FindLatestTerraformVersionRow, error) {
	ctx = context.WithValue(ctx, "pggen_query_name", "FindLatestTerraformVersion")
	rows, err := q.conn.Query(ctx, findLatestTerraformVersionSQL)
	if err != nil {
		return nil, fmt.Errorf("query FindLatestTerraformVersion: %w", err)
	}
	defer rows.Close()
	items := []FindLatestTerraformVersionRow{}
	for rows.Next() {
		var item FindLatestTerraformVersionRow
		if err := rows.Scan(&item.Version, &item.Checkpoint); err != nil {
			return nil, fmt.Errorf("scan FindLatestTerraformVersion row: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("close FindLatestTerraformVersion rows: %w", err)
	}
	return items, err
}

// FindLatestTerraformVersionBatch implements Querier.FindLatestTerraformVersionBatch.
func (q *DBQuerier) FindLatestTerraformVersionBatch(batch genericBatch) {
	batch.Queue(findLatestTerraformVersionSQL)
}

// FindLatestTerraformVersionScan implements Querier.FindLatestTerraformVersionScan.
func (q *DBQuerier) FindLatestTerraformVersionScan(results pgx.BatchResults) ([]FindLatestTerraformVersionRow, error) {
	rows, err := results.Query()
	if err != nil {
		return nil, fmt.Errorf("query FindLatestTerraformVersionBatch: %w", err)
	}
	defer rows.Close()
	items := []FindLatestTerraformVersionRow{}
	for rows.Next() {
		var item FindLatestTerraformVersionRow
		if err := rows.Scan(&item.Version, &item.Checkpoint); err != nil {
			return nil, fmt.Errorf("scan FindLatestTerraformVersionBatch row: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("close FindLatestTerraformVersionBatch rows: %w", err)
	}
	return items, err
}
