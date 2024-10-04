// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: run.sql

package sqlc

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const countRuns = `-- name: CountRuns :one
SELECT count(*)
FROM runs
JOIN workspaces USING(workspace_id)
JOIN (configuration_versions LEFT JOIN ingress_attributes ia USING (configuration_version_id)) USING (configuration_version_id)
WHERE
    workspaces.organization_name LIKE ANY($1::text[])
AND workspaces.workspace_id      LIKE ANY($2::text[])
AND workspaces.name              LIKE ANY($3::text[])
AND runs.source                  LIKE ANY($4::text[])
AND runs.status                  LIKE ANY($5::text[])
AND runs.plan_only::text         LIKE ANY($6::text[])
AND (($7::text IS NULL) OR ia.commit_sha = $7)
AND (($8::text IS NULL) OR ia.sender_username = $8)
`

type CountRunsParams struct {
	OrganizationNames []pgtype.Text
	WorkspaceIds      []pgtype.Text
	WorkspaceNames    []pgtype.Text
	Sources           []pgtype.Text
	Statuses          []pgtype.Text
	PlanOnly          []pgtype.Text
	CommitSHA         pgtype.Text
	VCSUsername       pgtype.Text
}

func (q *Queries) CountRuns(ctx context.Context, arg CountRunsParams) (int64, error) {
	row := q.db.QueryRow(ctx, countRuns,
		arg.OrganizationNames,
		arg.WorkspaceIds,
		arg.WorkspaceNames,
		arg.Sources,
		arg.Statuses,
		arg.PlanOnly,
		arg.CommitSHA,
		arg.VCSUsername,
	)
	var count int64
	err := row.Scan(&count)
	return count, err
}

const deleteRunByID = `-- name: DeleteRunByID :one
DELETE
FROM runs
WHERE run_id = $1
RETURNING run_id
`

func (q *Queries) DeleteRunByID(ctx context.Context, runID pgtype.Text) (pgtype.Text, error) {
	row := q.db.QueryRow(ctx, deleteRunByID, runID)
	var run_id pgtype.Text
	err := row.Scan(&run_id)
	return run_id, err
}

const findRunByID = `-- name: FindRunByID :one
SELECT
    runs.run_id,
    runs.created_at,
    runs.cancel_signaled_at,
    runs.is_destroy,
    runs.position_in_queue,
    runs.refresh,
    runs.refresh_only,
    runs.source,
    runs.status,
    plans.status      AS plan_status,
    applies.status      AS apply_status,
    runs.replace_addrs,
    runs.target_addrs,
    runs.auto_apply,
    plans.resource_report::"report" AS plan_resource_report,
    plans.output_report::"report" AS plan_output_report,
    applies.resource_report::"report" AS apply_resource_report,
    runs.configuration_version_id,
    runs.workspace_id,
    runs.plan_only,
    runs.created_by,
    runs.terraform_version,
    runs.allow_empty_apply,
    workspaces.execution_mode AS execution_mode,
    CASE WHEN workspaces.latest_run_id = runs.run_id THEN true
         ELSE false
    END AS latest,
    workspaces.organization_name,
    organizations.cost_estimation_enabled,
    rst.run_status_timestamps,
    pst.plan_status_timestamps,
    ast.apply_status_timestamps,
    rv.run_variables,
    ia::"ingress_attributes" AS ingress_attributes
FROM runs
JOIN plans USING (run_id)
JOIN applies USING (run_id)
JOIN workspaces ON runs.workspace_id = workspaces.workspace_id
JOIN organizations ON workspaces.organization_name = organizations.name
JOIN (configuration_versions cv LEFT JOIN ingress_attributes ia USING (configuration_version_id)) USING (configuration_version_id)
LEFT JOIN (
    SELECT
        run_id,
        array_agg(rv.*)::run_variables[] AS run_variables
    FROM run_variables rv
    GROUP BY run_id
) AS rv ON rv.run_id = runs.run_id
LEFT JOIN (
    SELECT
        run_id,
        array_agg(rst.*)::run_status_timestamps[] AS run_status_timestamps
    FROM run_status_timestamps rst
    GROUP BY run_id
) AS rst ON rst.run_id = runs.run_id
LEFT JOIN (
    SELECT
        run_id,
        array_agg(pst.*)::phase_status_timestamps[] AS plan_status_timestamps
    FROM phase_status_timestamps pst
    WHERE pst.phase = 'plan'
    GROUP BY run_id
) AS pst ON pst.run_id = runs.run_id
LEFT JOIN (
    SELECT
        run_id,
        array_agg(ast.*)::phase_status_timestamps[] AS apply_status_timestamps
    FROM phase_status_timestamps ast
    WHERE ast.phase = 'apply'
    GROUP BY run_id
) AS ast ON ast.run_id = runs.run_id
WHERE runs.run_id = $1
`

type FindRunByIDRow struct {
	RunID                  pgtype.Text
	CreatedAt              pgtype.Timestamptz
	CancelSignaledAt       pgtype.Timestamptz
	IsDestroy              pgtype.Bool
	PositionInQueue        pgtype.Int4
	Refresh                pgtype.Bool
	RefreshOnly            pgtype.Bool
	Source                 pgtype.Text
	Status                 pgtype.Text
	PlanStatus             pgtype.Text
	ApplyStatus            pgtype.Text
	ReplaceAddrs           []pgtype.Text
	TargetAddrs            []pgtype.Text
	AutoApply              pgtype.Bool
	PlanResourceReport     *Report
	PlanOutputReport       *Report
	ApplyResourceReport    *Report
	ConfigurationVersionID pgtype.Text
	WorkspaceID            pgtype.Text
	PlanOnly               pgtype.Bool
	CreatedBy              pgtype.Text
	TerraformVersion       pgtype.Text
	AllowEmptyApply        pgtype.Bool
	ExecutionMode          pgtype.Text
	Latest                 pgtype.Bool
	OrganizationName       pgtype.Text
	CostEstimationEnabled  pgtype.Bool
	RunStatusTimestamps    []RunStatusTimestamp
	PlanStatusTimestamps   []PhaseStatusTimestamp
	ApplyStatusTimestamps  []PhaseStatusTimestamp
	RunVariables           []RunVariable
	IngressAttributes      *IngressAttribute
}

func (q *Queries) FindRunByID(ctx context.Context, runID pgtype.Text) (FindRunByIDRow, error) {
	row := q.db.QueryRow(ctx, findRunByID, runID)
	var i FindRunByIDRow
	err := row.Scan(
		&i.RunID,
		&i.CreatedAt,
		&i.CancelSignaledAt,
		&i.IsDestroy,
		&i.PositionInQueue,
		&i.Refresh,
		&i.RefreshOnly,
		&i.Source,
		&i.Status,
		&i.PlanStatus,
		&i.ApplyStatus,
		&i.ReplaceAddrs,
		&i.TargetAddrs,
		&i.AutoApply,
		&i.PlanResourceReport,
		&i.PlanOutputReport,
		&i.ApplyResourceReport,
		&i.ConfigurationVersionID,
		&i.WorkspaceID,
		&i.PlanOnly,
		&i.CreatedBy,
		&i.TerraformVersion,
		&i.AllowEmptyApply,
		&i.ExecutionMode,
		&i.Latest,
		&i.OrganizationName,
		&i.CostEstimationEnabled,
		&i.RunStatusTimestamps,
		&i.PlanStatusTimestamps,
		&i.ApplyStatusTimestamps,
		&i.RunVariables,
		&i.IngressAttributes,
	)
	return i, err
}

const findRunByIDForUpdate = `-- name: FindRunByIDForUpdate :one
SELECT
    runs.run_id,
    runs.created_at,
    runs.cancel_signaled_at,
    runs.is_destroy,
    runs.position_in_queue,
    runs.refresh,
    runs.refresh_only,
    runs.source,
    runs.status,
    plans.status        AS plan_status,
    applies.status      AS apply_status,
    runs.replace_addrs,
    runs.target_addrs,
    runs.auto_apply,
    plans.resource_report::"report" AS plan_resource_report,
    plans.output_report::"report" AS plan_output_report,
    applies.resource_report::"report" AS apply_resource_report,
    runs.configuration_version_id,
    runs.workspace_id,
    runs.plan_only,
    runs.created_by,
    runs.terraform_version,
    runs.allow_empty_apply,
    workspaces.execution_mode AS execution_mode,
    CASE WHEN workspaces.latest_run_id = runs.run_id THEN true
         ELSE false
    END AS latest,
    workspaces.organization_name,
    organizations.cost_estimation_enabled,
    rst.run_status_timestamps,
    pst.plan_status_timestamps,
    ast.apply_status_timestamps,
    rv.run_variables,
    ia::"ingress_attributes" AS ingress_attributes
FROM runs
JOIN plans USING (run_id)
JOIN applies USING (run_id)
JOIN workspaces ON runs.workspace_id = workspaces.workspace_id
JOIN organizations ON workspaces.organization_name = organizations.name
JOIN (configuration_versions cv LEFT JOIN ingress_attributes ia USING (configuration_version_id)) USING (configuration_version_id)
LEFT JOIN (
    SELECT
        run_id,
        array_agg(rv.*)::run_variables[] AS run_variables
    FROM run_variables rv
    GROUP BY run_id
) AS rv ON rv.run_id = runs.run_id
LEFT JOIN (
    SELECT
        run_id,
        array_agg(rst.*)::run_status_timestamps[] AS run_status_timestamps
    FROM run_status_timestamps rst
    GROUP BY run_id
) AS rst ON rst.run_id = runs.run_id
LEFT JOIN (
    SELECT
        run_id,
        array_agg(pst.*)::phase_status_timestamps[] AS plan_status_timestamps
    FROM phase_status_timestamps pst
    WHERE pst.phase = 'plan'
    GROUP BY run_id
) AS pst ON pst.run_id = runs.run_id
LEFT JOIN (
    SELECT
        run_id,
        array_agg(ast.*)::phase_status_timestamps[] AS apply_status_timestamps
    FROM phase_status_timestamps ast
    WHERE ast.phase = 'apply'
    GROUP BY run_id
) AS ast ON ast.run_id = runs.run_id
WHERE runs.run_id = $1
FOR UPDATE OF runs, plans, applies
`

type FindRunByIDForUpdateRow struct {
	RunID                  pgtype.Text
	CreatedAt              pgtype.Timestamptz
	CancelSignaledAt       pgtype.Timestamptz
	IsDestroy              pgtype.Bool
	PositionInQueue        pgtype.Int4
	Refresh                pgtype.Bool
	RefreshOnly            pgtype.Bool
	Source                 pgtype.Text
	Status                 pgtype.Text
	PlanStatus             pgtype.Text
	ApplyStatus            pgtype.Text
	ReplaceAddrs           []pgtype.Text
	TargetAddrs            []pgtype.Text
	AutoApply              pgtype.Bool
	PlanResourceReport     *Report
	PlanOutputReport       *Report
	ApplyResourceReport    *Report
	ConfigurationVersionID pgtype.Text
	WorkspaceID            pgtype.Text
	PlanOnly               pgtype.Bool
	CreatedBy              pgtype.Text
	TerraformVersion       pgtype.Text
	AllowEmptyApply        pgtype.Bool
	ExecutionMode          pgtype.Text
	Latest                 pgtype.Bool
	OrganizationName       pgtype.Text
	CostEstimationEnabled  pgtype.Bool
	RunStatusTimestamps    []RunStatusTimestamp
	PlanStatusTimestamps   []PhaseStatusTimestamp
	ApplyStatusTimestamps  []PhaseStatusTimestamp
	RunVariables           []RunVariable
	IngressAttributes      *IngressAttribute
}

func (q *Queries) FindRunByIDForUpdate(ctx context.Context, runID pgtype.Text) (FindRunByIDForUpdateRow, error) {
	row := q.db.QueryRow(ctx, findRunByIDForUpdate, runID)
	var i FindRunByIDForUpdateRow
	err := row.Scan(
		&i.RunID,
		&i.CreatedAt,
		&i.CancelSignaledAt,
		&i.IsDestroy,
		&i.PositionInQueue,
		&i.Refresh,
		&i.RefreshOnly,
		&i.Source,
		&i.Status,
		&i.PlanStatus,
		&i.ApplyStatus,
		&i.ReplaceAddrs,
		&i.TargetAddrs,
		&i.AutoApply,
		&i.PlanResourceReport,
		&i.PlanOutputReport,
		&i.ApplyResourceReport,
		&i.ConfigurationVersionID,
		&i.WorkspaceID,
		&i.PlanOnly,
		&i.CreatedBy,
		&i.TerraformVersion,
		&i.AllowEmptyApply,
		&i.ExecutionMode,
		&i.Latest,
		&i.OrganizationName,
		&i.CostEstimationEnabled,
		&i.RunStatusTimestamps,
		&i.PlanStatusTimestamps,
		&i.ApplyStatusTimestamps,
		&i.RunVariables,
		&i.IngressAttributes,
	)
	return i, err
}

const findRuns = `-- name: FindRuns :many
SELECT
    runs.run_id,
    runs.created_at,
    runs.cancel_signaled_at,
    runs.is_destroy,
    runs.position_in_queue,
    runs.refresh,
    runs.refresh_only,
    runs.source,
    runs.status,
    plans.status      AS plan_status,
    applies.status      AS apply_status,
    runs.replace_addrs,
    runs.target_addrs,
    runs.auto_apply,
    plans.resource_report::"report" AS plan_resource_report,
    plans.output_report::"report" AS plan_output_report,
    applies.resource_report::"report" AS apply_resource_report,
    runs.configuration_version_id,
    runs.workspace_id,
    runs.plan_only,
    runs.created_by,
    runs.terraform_version,
    runs.allow_empty_apply,
    workspaces.execution_mode AS execution_mode,
    CASE WHEN workspaces.latest_run_id = runs.run_id THEN true
         ELSE false
    END AS latest,
    workspaces.organization_name,
    organizations.cost_estimation_enabled,
    rst.run_status_timestamps,
    pst.plan_status_timestamps,
    ast.apply_status_timestamps,
    rv.run_variables,
    ia::"ingress_attributes" AS ingress_attributes
FROM runs
JOIN plans USING (run_id)
JOIN applies USING (run_id)
JOIN workspaces ON runs.workspace_id = workspaces.workspace_id
JOIN organizations ON workspaces.organization_name = organizations.name
JOIN (configuration_versions cv LEFT JOIN ingress_attributes ia USING (configuration_version_id)) USING (configuration_version_id)
LEFT JOIN (
    SELECT
        run_id,
        array_agg(rv.*)::run_variables[] AS run_variables
    FROM run_variables rv
    GROUP BY run_id
) AS rv ON rv.run_id = runs.run_id
LEFT JOIN (
    SELECT
        run_id,
        array_agg(rst.*)::run_status_timestamps[] AS run_status_timestamps
    FROM run_status_timestamps rst
    GROUP BY run_id
) AS rst ON rst.run_id = runs.run_id
LEFT JOIN (
    SELECT
        run_id,
        array_agg(pst.*)::phase_status_timestamps[] AS plan_status_timestamps
    FROM phase_status_timestamps pst
    WHERE pst.phase = 'plan'
    GROUP BY run_id
) AS pst ON pst.run_id = runs.run_id
LEFT JOIN (
    SELECT
        run_id,
        array_agg(ast.*)::phase_status_timestamps[] AS apply_status_timestamps
    FROM phase_status_timestamps ast
    WHERE ast.phase = 'apply'
    GROUP BY run_id
) AS ast ON ast.run_id = runs.run_id
WHERE
    workspaces.organization_name LIKE ANY($1::text[])
AND workspaces.workspace_id      LIKE ANY($2::text[])
AND workspaces.name              LIKE ANY($3::text[])
AND runs.source                  LIKE ANY($4::text[])
AND runs.status                  LIKE ANY($5::text[])
AND runs.plan_only::text         LIKE ANY($6::text[])
AND (($7::text IS NULL) OR ia.commit_sha = $7)
AND (($8::text IS NULL) OR ia.sender_username = $8)
ORDER BY runs.created_at DESC
LIMIT $10::int
OFFSET $9::int
`

type FindRunsParams struct {
	OrganizationNames []pgtype.Text
	WorkspaceIds      []pgtype.Text
	WorkspaceNames    []pgtype.Text
	Sources           []pgtype.Text
	Statuses          []pgtype.Text
	PlanOnly          []pgtype.Text
	CommitSHA         pgtype.Text
	VCSUsername       pgtype.Text
	Offset            pgtype.Int4
	Limit             pgtype.Int4
}

type FindRunsRow struct {
	RunID                  pgtype.Text
	CreatedAt              pgtype.Timestamptz
	CancelSignaledAt       pgtype.Timestamptz
	IsDestroy              pgtype.Bool
	PositionInQueue        pgtype.Int4
	Refresh                pgtype.Bool
	RefreshOnly            pgtype.Bool
	Source                 pgtype.Text
	Status                 pgtype.Text
	PlanStatus             pgtype.Text
	ApplyStatus            pgtype.Text
	ReplaceAddrs           []pgtype.Text
	TargetAddrs            []pgtype.Text
	AutoApply              pgtype.Bool
	PlanResourceReport     *Report
	PlanOutputReport       *Report
	ApplyResourceReport    *Report
	ConfigurationVersionID pgtype.Text
	WorkspaceID            pgtype.Text
	PlanOnly               pgtype.Bool
	CreatedBy              pgtype.Text
	TerraformVersion       pgtype.Text
	AllowEmptyApply        pgtype.Bool
	ExecutionMode          pgtype.Text
	Latest                 pgtype.Bool
	OrganizationName       pgtype.Text
	CostEstimationEnabled  pgtype.Bool
	RunStatusTimestamps    []RunStatusTimestamp
	PlanStatusTimestamps   []PhaseStatusTimestamp
	ApplyStatusTimestamps  []PhaseStatusTimestamp
	RunVariables           []RunVariable
	IngressAttributes      *IngressAttribute
}

func (q *Queries) FindRuns(ctx context.Context, arg FindRunsParams) ([]FindRunsRow, error) {
	rows, err := q.db.Query(ctx, findRuns,
		arg.OrganizationNames,
		arg.WorkspaceIds,
		arg.WorkspaceNames,
		arg.Sources,
		arg.Statuses,
		arg.PlanOnly,
		arg.CommitSHA,
		arg.VCSUsername,
		arg.Offset,
		arg.Limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []FindRunsRow
	for rows.Next() {
		var i FindRunsRow
		if err := rows.Scan(
			&i.RunID,
			&i.CreatedAt,
			&i.CancelSignaledAt,
			&i.IsDestroy,
			&i.PositionInQueue,
			&i.Refresh,
			&i.RefreshOnly,
			&i.Source,
			&i.Status,
			&i.PlanStatus,
			&i.ApplyStatus,
			&i.ReplaceAddrs,
			&i.TargetAddrs,
			&i.AutoApply,
			&i.PlanResourceReport,
			&i.PlanOutputReport,
			&i.ApplyResourceReport,
			&i.ConfigurationVersionID,
			&i.WorkspaceID,
			&i.PlanOnly,
			&i.CreatedBy,
			&i.TerraformVersion,
			&i.AllowEmptyApply,
			&i.ExecutionMode,
			&i.Latest,
			&i.OrganizationName,
			&i.CostEstimationEnabled,
			&i.RunStatusTimestamps,
			&i.PlanStatusTimestamps,
			&i.ApplyStatusTimestamps,
			&i.RunVariables,
			&i.IngressAttributes,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getLockFileByID = `-- name: GetLockFileByID :one
SELECT lock_file
FROM runs
WHERE run_id = $1
`

func (q *Queries) GetLockFileByID(ctx context.Context, runID pgtype.Text) ([]byte, error) {
	row := q.db.QueryRow(ctx, getLockFileByID, runID)
	var lock_file []byte
	err := row.Scan(&lock_file)
	return lock_file, err
}

const insertRun = `-- name: InsertRun :exec
INSERT INTO runs (
    run_id,
    created_at,
    is_destroy,
    position_in_queue,
    refresh,
    refresh_only,
    source,
    status,
    replace_addrs,
    target_addrs,
    auto_apply,
    plan_only,
    configuration_version_id,
    workspace_id,
    created_by,
    terraform_version,
    allow_empty_apply
) VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6,
    $7,
    $8,
    $9,
    $10,
    $11,
    $12,
    $13,
    $14,
    $15,
    $16,
    $17
)
`

type InsertRunParams struct {
	ID                     pgtype.Text
	CreatedAt              pgtype.Timestamptz
	IsDestroy              pgtype.Bool
	PositionInQueue        pgtype.Int4
	Refresh                pgtype.Bool
	RefreshOnly            pgtype.Bool
	Source                 pgtype.Text
	Status                 pgtype.Text
	ReplaceAddrs           []pgtype.Text
	TargetAddrs            []pgtype.Text
	AutoApply              pgtype.Bool
	PlanOnly               pgtype.Bool
	ConfigurationVersionID pgtype.Text
	WorkspaceID            pgtype.Text
	CreatedBy              pgtype.Text
	TerraformVersion       pgtype.Text
	AllowEmptyApply        pgtype.Bool
}

func (q *Queries) InsertRun(ctx context.Context, arg InsertRunParams) error {
	_, err := q.db.Exec(ctx, insertRun,
		arg.ID,
		arg.CreatedAt,
		arg.IsDestroy,
		arg.PositionInQueue,
		arg.Refresh,
		arg.RefreshOnly,
		arg.Source,
		arg.Status,
		arg.ReplaceAddrs,
		arg.TargetAddrs,
		arg.AutoApply,
		arg.PlanOnly,
		arg.ConfigurationVersionID,
		arg.WorkspaceID,
		arg.CreatedBy,
		arg.TerraformVersion,
		arg.AllowEmptyApply,
	)
	return err
}

const insertRunStatusTimestamp = `-- name: InsertRunStatusTimestamp :exec
INSERT INTO run_status_timestamps (
    run_id,
    status,
    timestamp
) VALUES (
    $1,
    $2,
    $3
)
`

type InsertRunStatusTimestampParams struct {
	ID        pgtype.Text
	Status    pgtype.Text
	Timestamp pgtype.Timestamptz
}

func (q *Queries) InsertRunStatusTimestamp(ctx context.Context, arg InsertRunStatusTimestampParams) error {
	_, err := q.db.Exec(ctx, insertRunStatusTimestamp, arg.ID, arg.Status, arg.Timestamp)
	return err
}

const insertRunVariable = `-- name: InsertRunVariable :exec
INSERT INTO run_variables (
    run_id,
    key,
    value
) VALUES (
    $1,
    $2,
    $3
)
`

type InsertRunVariableParams struct {
	RunID pgtype.Text
	Key   pgtype.Text
	Value pgtype.Text
}

func (q *Queries) InsertRunVariable(ctx context.Context, arg InsertRunVariableParams) error {
	_, err := q.db.Exec(ctx, insertRunVariable, arg.RunID, arg.Key, arg.Value)
	return err
}

const putLockFile = `-- name: PutLockFile :one
UPDATE runs
SET lock_file = $1
WHERE run_id = $2
RETURNING run_id
`

type PutLockFileParams struct {
	LockFile []byte
	RunID    pgtype.Text
}

func (q *Queries) PutLockFile(ctx context.Context, arg PutLockFileParams) (pgtype.Text, error) {
	row := q.db.QueryRow(ctx, putLockFile, arg.LockFile, arg.RunID)
	var run_id pgtype.Text
	err := row.Scan(&run_id)
	return run_id, err
}

const updateCancelSignaledAt = `-- name: UpdateCancelSignaledAt :one
UPDATE runs
SET
    cancel_signaled_at = $1
WHERE run_id = $2
RETURNING run_id
`

type UpdateCancelSignaledAtParams struct {
	CancelSignaledAt pgtype.Timestamptz
	ID               pgtype.Text
}

func (q *Queries) UpdateCancelSignaledAt(ctx context.Context, arg UpdateCancelSignaledAtParams) (pgtype.Text, error) {
	row := q.db.QueryRow(ctx, updateCancelSignaledAt, arg.CancelSignaledAt, arg.ID)
	var run_id pgtype.Text
	err := row.Scan(&run_id)
	return run_id, err
}

const updateRunStatus = `-- name: UpdateRunStatus :one
UPDATE runs
SET
    status = $1
WHERE run_id = $2
RETURNING run_id
`

type UpdateRunStatusParams struct {
	Status pgtype.Text
	ID     pgtype.Text
}

func (q *Queries) UpdateRunStatus(ctx context.Context, arg UpdateRunStatusParams) (pgtype.Text, error) {
	row := q.db.QueryRow(ctx, updateRunStatus, arg.Status, arg.ID)
	var run_id pgtype.Text
	err := row.Scan(&run_id)
	return run_id, err
}
