// Code generated by pggen. DO NOT EDIT.

package sql

import (
	"context"
	"fmt"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"time"
)

const insertPlanSQL = `INSERT INTO plans (
    plan_id,
    created_at,
    updated_at,
    status,
    run_id
) VALUES (
    $1,
    NOW(),
    NOW(),
    $2,
    $3
)
RETURNING *;`

type InsertPlanParams struct {
	ID     *string
	Status *string
	RunID  *string
}

type InsertPlanRow struct {
	PlanID               *string   `json:"plan_id"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
	ResourceAdditions    int32     `json:"resource_additions"`
	ResourceChanges      int32     `json:"resource_changes"`
	ResourceDestructions int32     `json:"resource_destructions"`
	Status               *string   `json:"status"`
	StatusTimestamps     *string   `json:"status_timestamps"`
	PlanBin              []byte    `json:"plan_bin"`
	PlanJson             []byte    `json:"plan_json"`
	RunID                *string   `json:"run_id"`
}

func (s InsertPlanRow) GetPlanID() *string { return s.PlanID }
func (s InsertPlanRow) GetCreatedAt() time.Time { return s.CreatedAt }
func (s InsertPlanRow) GetUpdatedAt() time.Time { return s.UpdatedAt }
func (s InsertPlanRow) GetResourceAdditions() int32 { return s.ResourceAdditions }
func (s InsertPlanRow) GetResourceChanges() int32 { return s.ResourceChanges }
func (s InsertPlanRow) GetResourceDestructions() int32 { return s.ResourceDestructions }
func (s InsertPlanRow) GetStatus() *string { return s.Status }
func (s InsertPlanRow) GetStatusTimestamps() *string { return s.StatusTimestamps }
func (s InsertPlanRow) GetPlanBin() []byte { return s.PlanBin }
func (s InsertPlanRow) GetPlanJson() []byte { return s.PlanJson }
func (s InsertPlanRow) GetRunID() *string { return s.RunID }


// InsertPlan implements Querier.InsertPlan.
func (q *DBQuerier) InsertPlan(ctx context.Context, params InsertPlanParams) (InsertPlanRow, error) {
	ctx = context.WithValue(ctx, "pggen_query_name", "InsertPlan")
	row := q.conn.QueryRow(ctx, insertPlanSQL, params.ID, params.Status, params.RunID)
	var item InsertPlanRow
	if err := row.Scan(&item.PlanID, &item.CreatedAt, &item.UpdatedAt, &item.ResourceAdditions, &item.ResourceChanges, &item.ResourceDestructions, &item.Status, &item.StatusTimestamps, &item.PlanBin, &item.PlanJson, &item.RunID); err != nil {
		return item, fmt.Errorf("query InsertPlan: %w", err)
	}
	return item, nil
}

// InsertPlanBatch implements Querier.InsertPlanBatch.
func (q *DBQuerier) InsertPlanBatch(batch genericBatch, params InsertPlanParams) {
	batch.Queue(insertPlanSQL, params.ID, params.Status, params.RunID)
}

// InsertPlanScan implements Querier.InsertPlanScan.
func (q *DBQuerier) InsertPlanScan(results pgx.BatchResults) (InsertPlanRow, error) {
	row := results.QueryRow()
	var item InsertPlanRow
	if err := row.Scan(&item.PlanID, &item.CreatedAt, &item.UpdatedAt, &item.ResourceAdditions, &item.ResourceChanges, &item.ResourceDestructions, &item.Status, &item.StatusTimestamps, &item.PlanBin, &item.PlanJson, &item.RunID); err != nil {
		return item, fmt.Errorf("scan InsertPlanBatch row: %w", err)
	}
	return item, nil
}

const insertPlanStatusTimestampSQL = `INSERT INTO plan_status_timestamps (
    plan_id,
    status,
    timestamp
) VALUES (
    $1,
    $2,
    NOW()
)
RETURNING *;`

type InsertPlanStatusTimestampRow struct {
	PlanID    *string   `json:"plan_id"`
	Status    *string   `json:"status"`
	Timestamp time.Time `json:"timestamp"`
}

func (s InsertPlanStatusTimestampRow) GetPlanID() *string { return s.PlanID }
func (s InsertPlanStatusTimestampRow) GetStatus() *string { return s.Status }
func (s InsertPlanStatusTimestampRow) GetTimestamp() time.Time { return s.Timestamp }


// InsertPlanStatusTimestamp implements Querier.InsertPlanStatusTimestamp.
func (q *DBQuerier) InsertPlanStatusTimestamp(ctx context.Context, id *string, status *string) (InsertPlanStatusTimestampRow, error) {
	ctx = context.WithValue(ctx, "pggen_query_name", "InsertPlanStatusTimestamp")
	row := q.conn.QueryRow(ctx, insertPlanStatusTimestampSQL, id, status)
	var item InsertPlanStatusTimestampRow
	if err := row.Scan(&item.PlanID, &item.Status, &item.Timestamp); err != nil {
		return item, fmt.Errorf("query InsertPlanStatusTimestamp: %w", err)
	}
	return item, nil
}

// InsertPlanStatusTimestampBatch implements Querier.InsertPlanStatusTimestampBatch.
func (q *DBQuerier) InsertPlanStatusTimestampBatch(batch genericBatch, id *string, status *string) {
	batch.Queue(insertPlanStatusTimestampSQL, id, status)
}

// InsertPlanStatusTimestampScan implements Querier.InsertPlanStatusTimestampScan.
func (q *DBQuerier) InsertPlanStatusTimestampScan(results pgx.BatchResults) (InsertPlanStatusTimestampRow, error) {
	row := results.QueryRow()
	var item InsertPlanStatusTimestampRow
	if err := row.Scan(&item.PlanID, &item.Status, &item.Timestamp); err != nil {
		return item, fmt.Errorf("scan InsertPlanStatusTimestampBatch row: %w", err)
	}
	return item, nil
}

const updatePlanStatusSQL = `UPDATE plans
SET
    status = $1,
    updated_at = NOW()
WHERE plan_id = $2
RETURNING *;`

type UpdatePlanStatusRow struct {
	PlanID               *string   `json:"plan_id"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
	ResourceAdditions    int32     `json:"resource_additions"`
	ResourceChanges      int32     `json:"resource_changes"`
	ResourceDestructions int32     `json:"resource_destructions"`
	Status               *string   `json:"status"`
	StatusTimestamps     *string   `json:"status_timestamps"`
	PlanBin              []byte    `json:"plan_bin"`
	PlanJson             []byte    `json:"plan_json"`
	RunID                *string   `json:"run_id"`
}

func (s UpdatePlanStatusRow) GetPlanID() *string { return s.PlanID }
func (s UpdatePlanStatusRow) GetCreatedAt() time.Time { return s.CreatedAt }
func (s UpdatePlanStatusRow) GetUpdatedAt() time.Time { return s.UpdatedAt }
func (s UpdatePlanStatusRow) GetResourceAdditions() int32 { return s.ResourceAdditions }
func (s UpdatePlanStatusRow) GetResourceChanges() int32 { return s.ResourceChanges }
func (s UpdatePlanStatusRow) GetResourceDestructions() int32 { return s.ResourceDestructions }
func (s UpdatePlanStatusRow) GetStatus() *string { return s.Status }
func (s UpdatePlanStatusRow) GetStatusTimestamps() *string { return s.StatusTimestamps }
func (s UpdatePlanStatusRow) GetPlanBin() []byte { return s.PlanBin }
func (s UpdatePlanStatusRow) GetPlanJson() []byte { return s.PlanJson }
func (s UpdatePlanStatusRow) GetRunID() *string { return s.RunID }


// UpdatePlanStatus implements Querier.UpdatePlanStatus.
func (q *DBQuerier) UpdatePlanStatus(ctx context.Context, status *string, id *string) (UpdatePlanStatusRow, error) {
	ctx = context.WithValue(ctx, "pggen_query_name", "UpdatePlanStatus")
	row := q.conn.QueryRow(ctx, updatePlanStatusSQL, status, id)
	var item UpdatePlanStatusRow
	if err := row.Scan(&item.PlanID, &item.CreatedAt, &item.UpdatedAt, &item.ResourceAdditions, &item.ResourceChanges, &item.ResourceDestructions, &item.Status, &item.StatusTimestamps, &item.PlanBin, &item.PlanJson, &item.RunID); err != nil {
		return item, fmt.Errorf("query UpdatePlanStatus: %w", err)
	}
	return item, nil
}

// UpdatePlanStatusBatch implements Querier.UpdatePlanStatusBatch.
func (q *DBQuerier) UpdatePlanStatusBatch(batch genericBatch, status *string, id *string) {
	batch.Queue(updatePlanStatusSQL, status, id)
}

// UpdatePlanStatusScan implements Querier.UpdatePlanStatusScan.
func (q *DBQuerier) UpdatePlanStatusScan(results pgx.BatchResults) (UpdatePlanStatusRow, error) {
	row := results.QueryRow()
	var item UpdatePlanStatusRow
	if err := row.Scan(&item.PlanID, &item.CreatedAt, &item.UpdatedAt, &item.ResourceAdditions, &item.ResourceChanges, &item.ResourceDestructions, &item.Status, &item.StatusTimestamps, &item.PlanBin, &item.PlanJson, &item.RunID); err != nil {
		return item, fmt.Errorf("scan UpdatePlanStatusBatch row: %w", err)
	}
	return item, nil
}

const getPlanBinByRunIDSQL = `SELECT plan_bin
FROM plans
WHERE run_id = $1
;`

// GetPlanBinByRunID implements Querier.GetPlanBinByRunID.
func (q *DBQuerier) GetPlanBinByRunID(ctx context.Context, runID *string) ([]byte, error) {
	ctx = context.WithValue(ctx, "pggen_query_name", "GetPlanBinByRunID")
	row := q.conn.QueryRow(ctx, getPlanBinByRunIDSQL, runID)
	item := []byte{}
	if err := row.Scan(&item); err != nil {
		return item, fmt.Errorf("query GetPlanBinByRunID: %w", err)
	}
	return item, nil
}

// GetPlanBinByRunIDBatch implements Querier.GetPlanBinByRunIDBatch.
func (q *DBQuerier) GetPlanBinByRunIDBatch(batch genericBatch, runID *string) {
	batch.Queue(getPlanBinByRunIDSQL, runID)
}

// GetPlanBinByRunIDScan implements Querier.GetPlanBinByRunIDScan.
func (q *DBQuerier) GetPlanBinByRunIDScan(results pgx.BatchResults) ([]byte, error) {
	row := results.QueryRow()
	item := []byte{}
	if err := row.Scan(&item); err != nil {
		return item, fmt.Errorf("scan GetPlanBinByRunIDBatch row: %w", err)
	}
	return item, nil
}

const getPlanJSONByRunIDSQL = `SELECT plan_json
FROM plans
WHERE run_id = $1
;`

// GetPlanJSONByRunID implements Querier.GetPlanJSONByRunID.
func (q *DBQuerier) GetPlanJSONByRunID(ctx context.Context, runID *string) ([]byte, error) {
	ctx = context.WithValue(ctx, "pggen_query_name", "GetPlanJSONByRunID")
	row := q.conn.QueryRow(ctx, getPlanJSONByRunIDSQL, runID)
	item := []byte{}
	if err := row.Scan(&item); err != nil {
		return item, fmt.Errorf("query GetPlanJSONByRunID: %w", err)
	}
	return item, nil
}

// GetPlanJSONByRunIDBatch implements Querier.GetPlanJSONByRunIDBatch.
func (q *DBQuerier) GetPlanJSONByRunIDBatch(batch genericBatch, runID *string) {
	batch.Queue(getPlanJSONByRunIDSQL, runID)
}

// GetPlanJSONByRunIDScan implements Querier.GetPlanJSONByRunIDScan.
func (q *DBQuerier) GetPlanJSONByRunIDScan(results pgx.BatchResults) ([]byte, error) {
	row := results.QueryRow()
	item := []byte{}
	if err := row.Scan(&item); err != nil {
		return item, fmt.Errorf("scan GetPlanJSONByRunIDBatch row: %w", err)
	}
	return item, nil
}

const putPlanBinByRunIDSQL = `UPDATE plans
SET plan_bin = $1
WHERE run_id = $2
;`

// PutPlanBinByRunID implements Querier.PutPlanBinByRunID.
func (q *DBQuerier) PutPlanBinByRunID(ctx context.Context, planBin []byte, runID *string) (pgconn.CommandTag, error) {
	ctx = context.WithValue(ctx, "pggen_query_name", "PutPlanBinByRunID")
	cmdTag, err := q.conn.Exec(ctx, putPlanBinByRunIDSQL, planBin, runID)
	if err != nil {
		return cmdTag, fmt.Errorf("exec query PutPlanBinByRunID: %w", err)
	}
	return cmdTag, err
}

// PutPlanBinByRunIDBatch implements Querier.PutPlanBinByRunIDBatch.
func (q *DBQuerier) PutPlanBinByRunIDBatch(batch genericBatch, planBin []byte, runID *string) {
	batch.Queue(putPlanBinByRunIDSQL, planBin, runID)
}

// PutPlanBinByRunIDScan implements Querier.PutPlanBinByRunIDScan.
func (q *DBQuerier) PutPlanBinByRunIDScan(results pgx.BatchResults) (pgconn.CommandTag, error) {
	cmdTag, err := results.Exec()
	if err != nil {
		return cmdTag, fmt.Errorf("exec PutPlanBinByRunIDBatch: %w", err)
	}
	return cmdTag, err
}

const putPlanJSONByRunIDSQL = `UPDATE plans
SET plan_json = $1
WHERE run_id = $2
;`

// PutPlanJSONByRunID implements Querier.PutPlanJSONByRunID.
func (q *DBQuerier) PutPlanJSONByRunID(ctx context.Context, planJson []byte, runID *string) (pgconn.CommandTag, error) {
	ctx = context.WithValue(ctx, "pggen_query_name", "PutPlanJSONByRunID")
	cmdTag, err := q.conn.Exec(ctx, putPlanJSONByRunIDSQL, planJson, runID)
	if err != nil {
		return cmdTag, fmt.Errorf("exec query PutPlanJSONByRunID: %w", err)
	}
	return cmdTag, err
}

// PutPlanJSONByRunIDBatch implements Querier.PutPlanJSONByRunIDBatch.
func (q *DBQuerier) PutPlanJSONByRunIDBatch(batch genericBatch, planJson []byte, runID *string) {
	batch.Queue(putPlanJSONByRunIDSQL, planJson, runID)
}

// PutPlanJSONByRunIDScan implements Querier.PutPlanJSONByRunIDScan.
func (q *DBQuerier) PutPlanJSONByRunIDScan(results pgx.BatchResults) (pgconn.CommandTag, error) {
	cmdTag, err := results.Exec()
	if err != nil {
		return cmdTag, fmt.Errorf("exec PutPlanJSONByRunIDBatch: %w", err)
	}
	return cmdTag, err
}

const updatePlanResourcesSQL = `UPDATE plans
SET
    resource_additions = $1,
    resource_changes = $2,
    resource_destructions = $3
WHERE run_id = $4
;`

type UpdatePlanResourcesParams struct {
	ResourceAdditions    int32
	ResourceChanges      int32
	ResourceDestructions int32
	RunID                *string
}

// UpdatePlanResources implements Querier.UpdatePlanResources.
func (q *DBQuerier) UpdatePlanResources(ctx context.Context, params UpdatePlanResourcesParams) (pgconn.CommandTag, error) {
	ctx = context.WithValue(ctx, "pggen_query_name", "UpdatePlanResources")
	cmdTag, err := q.conn.Exec(ctx, updatePlanResourcesSQL, params.ResourceAdditions, params.ResourceChanges, params.ResourceDestructions, params.RunID)
	if err != nil {
		return cmdTag, fmt.Errorf("exec query UpdatePlanResources: %w", err)
	}
	return cmdTag, err
}

// UpdatePlanResourcesBatch implements Querier.UpdatePlanResourcesBatch.
func (q *DBQuerier) UpdatePlanResourcesBatch(batch genericBatch, params UpdatePlanResourcesParams) {
	batch.Queue(updatePlanResourcesSQL, params.ResourceAdditions, params.ResourceChanges, params.ResourceDestructions, params.RunID)
}

// UpdatePlanResourcesScan implements Querier.UpdatePlanResourcesScan.
func (q *DBQuerier) UpdatePlanResourcesScan(results pgx.BatchResults) (pgconn.CommandTag, error) {
	cmdTag, err := results.Exec()
	if err != nil {
		return cmdTag, fmt.Errorf("exec UpdatePlanResourcesBatch: %w", err)
	}
	return cmdTag, err
}
