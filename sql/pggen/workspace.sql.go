// Code generated by pggen. DO NOT EDIT.

package pggen

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
)

const insertWorkspaceSQL = `INSERT INTO workspaces (
    workspace_id,
    created_at,
    updated_at,
    allow_destroy_plan,
    auto_apply,
    can_queue_destroy_plan,
    description,
    environment,
    execution_mode,
    file_triggers_enabled,
    global_remote_state,
    migration_environment,
    name,
    queue_all_runs,
    speculative_enabled,
    source_name,
    source_url,
    structured_run_output_enabled,
    terraform_version,
    trigger_prefixes,
    working_directory,
    organization_id
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
    $17,
    $18,
    $19,
    $20,
    $21,
    $22
);`

type InsertWorkspaceParams struct {
	ID                         string
	CreatedAt                  time.Time
	UpdatedAt                  time.Time
	AllowDestroyPlan           bool
	AutoApply                  bool
	CanQueueDestroyPlan        bool
	Description                string
	Environment                string
	ExecutionMode              string
	FileTriggersEnabled        bool
	GlobalRemoteState          bool
	MigrationEnvironment       string
	Name                       string
	QueueAllRuns               bool
	SpeculativeEnabled         bool
	SourceName                 string
	SourceURL                  string
	StructuredRunOutputEnabled bool
	TerraformVersion           string
	TriggerPrefixes            []string
	WorkingDirectory           string
	OrganizationID             string
}

// InsertWorkspace implements Querier.InsertWorkspace.
func (q *DBQuerier) InsertWorkspace(ctx context.Context, params InsertWorkspaceParams) (pgconn.CommandTag, error) {
	ctx = context.WithValue(ctx, "pggen_query_name", "InsertWorkspace")
	cmdTag, err := q.conn.Exec(ctx, insertWorkspaceSQL, params.ID, params.CreatedAt, params.UpdatedAt, params.AllowDestroyPlan, params.AutoApply, params.CanQueueDestroyPlan, params.Description, params.Environment, params.ExecutionMode, params.FileTriggersEnabled, params.GlobalRemoteState, params.MigrationEnvironment, params.Name, params.QueueAllRuns, params.SpeculativeEnabled, params.SourceName, params.SourceURL, params.StructuredRunOutputEnabled, params.TerraformVersion, params.TriggerPrefixes, params.WorkingDirectory, params.OrganizationID)
	if err != nil {
		return cmdTag, fmt.Errorf("exec query InsertWorkspace: %w", err)
	}
	return cmdTag, err
}

// InsertWorkspaceBatch implements Querier.InsertWorkspaceBatch.
func (q *DBQuerier) InsertWorkspaceBatch(batch genericBatch, params InsertWorkspaceParams) {
	batch.Queue(insertWorkspaceSQL, params.ID, params.CreatedAt, params.UpdatedAt, params.AllowDestroyPlan, params.AutoApply, params.CanQueueDestroyPlan, params.Description, params.Environment, params.ExecutionMode, params.FileTriggersEnabled, params.GlobalRemoteState, params.MigrationEnvironment, params.Name, params.QueueAllRuns, params.SpeculativeEnabled, params.SourceName, params.SourceURL, params.StructuredRunOutputEnabled, params.TerraformVersion, params.TriggerPrefixes, params.WorkingDirectory, params.OrganizationID)
}

// InsertWorkspaceScan implements Querier.InsertWorkspaceScan.
func (q *DBQuerier) InsertWorkspaceScan(results pgx.BatchResults) (pgconn.CommandTag, error) {
	cmdTag, err := results.Exec()
	if err != nil {
		return cmdTag, fmt.Errorf("exec InsertWorkspaceBatch: %w", err)
	}
	return cmdTag, err
}

const findWorkspacesSQL = `SELECT
    workspaces.*,
    CASE WHEN $1 THEN (organizations.*)::"organizations" END AS organization
FROM workspaces
JOIN organizations USING (organization_id)
WHERE workspaces.name LIKE $2 || '%'
AND organizations.name = $3
LIMIT $4
OFFSET $5
;`

type FindWorkspacesParams struct {
	IncludeOrganization bool
	Prefix              string
	OrganizationName    string
	Limit               int
	Offset              int
}

type FindWorkspacesRow struct {
	WorkspaceID                string         `json:"workspace_id"`
	CreatedAt                  time.Time      `json:"created_at"`
	UpdatedAt                  time.Time      `json:"updated_at"`
	AllowDestroyPlan           bool           `json:"allow_destroy_plan"`
	AutoApply                  bool           `json:"auto_apply"`
	CanQueueDestroyPlan        bool           `json:"can_queue_destroy_plan"`
	Description                string         `json:"description"`
	Environment                string         `json:"environment"`
	ExecutionMode              string         `json:"execution_mode"`
	FileTriggersEnabled        bool           `json:"file_triggers_enabled"`
	GlobalRemoteState          bool           `json:"global_remote_state"`
	MigrationEnvironment       string         `json:"migration_environment"`
	Name                       string         `json:"name"`
	QueueAllRuns               bool           `json:"queue_all_runs"`
	SpeculativeEnabled         bool           `json:"speculative_enabled"`
	SourceName                 string         `json:"source_name"`
	SourceURL                  string         `json:"source_url"`
	StructuredRunOutputEnabled bool           `json:"structured_run_output_enabled"`
	TerraformVersion           string         `json:"terraform_version"`
	TriggerPrefixes            []string       `json:"trigger_prefixes"`
	WorkingDirectory           string         `json:"working_directory"`
	OrganizationID             string         `json:"organization_id"`
	Organization               *Organizations `json:"organization"`
}

// FindWorkspaces implements Querier.FindWorkspaces.
func (q *DBQuerier) FindWorkspaces(ctx context.Context, params FindWorkspacesParams) ([]FindWorkspacesRow, error) {
	ctx = context.WithValue(ctx, "pggen_query_name", "FindWorkspaces")
	rows, err := q.conn.Query(ctx, findWorkspacesSQL, params.IncludeOrganization, params.Prefix, params.OrganizationName, params.Limit, params.Offset)
	if err != nil {
		return nil, fmt.Errorf("query FindWorkspaces: %w", err)
	}
	defer rows.Close()
	items := []FindWorkspacesRow{}
	organizationRow := q.types.newOrganizations()
	for rows.Next() {
		var item FindWorkspacesRow
		if err := rows.Scan(&item.WorkspaceID, &item.CreatedAt, &item.UpdatedAt, &item.AllowDestroyPlan, &item.AutoApply, &item.CanQueueDestroyPlan, &item.Description, &item.Environment, &item.ExecutionMode, &item.FileTriggersEnabled, &item.GlobalRemoteState, &item.MigrationEnvironment, &item.Name, &item.QueueAllRuns, &item.SpeculativeEnabled, &item.SourceName, &item.SourceURL, &item.StructuredRunOutputEnabled, &item.TerraformVersion, &item.TriggerPrefixes, &item.WorkingDirectory, &item.OrganizationID, organizationRow); err != nil {
			return nil, fmt.Errorf("scan FindWorkspaces row: %w", err)
		}
		if err := organizationRow.AssignTo(&item.Organization); err != nil {
			return nil, fmt.Errorf("assign FindWorkspaces row: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("close FindWorkspaces rows: %w", err)
	}
	return items, err
}

// FindWorkspacesBatch implements Querier.FindWorkspacesBatch.
func (q *DBQuerier) FindWorkspacesBatch(batch genericBatch, params FindWorkspacesParams) {
	batch.Queue(findWorkspacesSQL, params.IncludeOrganization, params.Prefix, params.OrganizationName, params.Limit, params.Offset)
}

// FindWorkspacesScan implements Querier.FindWorkspacesScan.
func (q *DBQuerier) FindWorkspacesScan(results pgx.BatchResults) ([]FindWorkspacesRow, error) {
	rows, err := results.Query()
	if err != nil {
		return nil, fmt.Errorf("query FindWorkspacesBatch: %w", err)
	}
	defer rows.Close()
	items := []FindWorkspacesRow{}
	organizationRow := q.types.newOrganizations()
	for rows.Next() {
		var item FindWorkspacesRow
		if err := rows.Scan(&item.WorkspaceID, &item.CreatedAt, &item.UpdatedAt, &item.AllowDestroyPlan, &item.AutoApply, &item.CanQueueDestroyPlan, &item.Description, &item.Environment, &item.ExecutionMode, &item.FileTriggersEnabled, &item.GlobalRemoteState, &item.MigrationEnvironment, &item.Name, &item.QueueAllRuns, &item.SpeculativeEnabled, &item.SourceName, &item.SourceURL, &item.StructuredRunOutputEnabled, &item.TerraformVersion, &item.TriggerPrefixes, &item.WorkingDirectory, &item.OrganizationID, organizationRow); err != nil {
			return nil, fmt.Errorf("scan FindWorkspacesBatch row: %w", err)
		}
		if err := organizationRow.AssignTo(&item.Organization); err != nil {
			return nil, fmt.Errorf("assign FindWorkspaces row: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("close FindWorkspacesBatch rows: %w", err)
	}
	return items, err
}

const countWorkspacesSQL = `SELECT count(*)
FROM workspaces
JOIN organizations USING (organization_id)
WHERE workspaces.name LIKE $1 || '%'
AND organizations.name = $2
;`

// CountWorkspaces implements Querier.CountWorkspaces.
func (q *DBQuerier) CountWorkspaces(ctx context.Context, prefix string, organizationName string) (*int, error) {
	ctx = context.WithValue(ctx, "pggen_query_name", "CountWorkspaces")
	row := q.conn.QueryRow(ctx, countWorkspacesSQL, prefix, organizationName)
	var item int
	if err := row.Scan(&item); err != nil {
		return &item, fmt.Errorf("query CountWorkspaces: %w", err)
	}
	return &item, nil
}

// CountWorkspacesBatch implements Querier.CountWorkspacesBatch.
func (q *DBQuerier) CountWorkspacesBatch(batch genericBatch, prefix string, organizationName string) {
	batch.Queue(countWorkspacesSQL, prefix, organizationName)
}

// CountWorkspacesScan implements Querier.CountWorkspacesScan.
func (q *DBQuerier) CountWorkspacesScan(results pgx.BatchResults) (*int, error) {
	row := results.QueryRow()
	var item int
	if err := row.Scan(&item); err != nil {
		return &item, fmt.Errorf("scan CountWorkspacesBatch row: %w", err)
	}
	return &item, nil
}

const findWorkspaceIDByNameSQL = `SELECT workspaces.workspace_id
FROM workspaces
JOIN organizations USING (organization_id)
WHERE workspaces.name = $1
AND organizations.name = $2;`

// FindWorkspaceIDByName implements Querier.FindWorkspaceIDByName.
func (q *DBQuerier) FindWorkspaceIDByName(ctx context.Context, name string, organizationName string) (string, error) {
	ctx = context.WithValue(ctx, "pggen_query_name", "FindWorkspaceIDByName")
	row := q.conn.QueryRow(ctx, findWorkspaceIDByNameSQL, name, organizationName)
	var item string
	if err := row.Scan(&item); err != nil {
		return item, fmt.Errorf("query FindWorkspaceIDByName: %w", err)
	}
	return item, nil
}

// FindWorkspaceIDByNameBatch implements Querier.FindWorkspaceIDByNameBatch.
func (q *DBQuerier) FindWorkspaceIDByNameBatch(batch genericBatch, name string, organizationName string) {
	batch.Queue(findWorkspaceIDByNameSQL, name, organizationName)
}

// FindWorkspaceIDByNameScan implements Querier.FindWorkspaceIDByNameScan.
func (q *DBQuerier) FindWorkspaceIDByNameScan(results pgx.BatchResults) (string, error) {
	row := results.QueryRow()
	var item string
	if err := row.Scan(&item); err != nil {
		return item, fmt.Errorf("scan FindWorkspaceIDByNameBatch row: %w", err)
	}
	return item, nil
}

const findWorkspaceByNameSQL = `SELECT
    workspaces.*,
    CASE WHEN $1 THEN (organizations.*)::"organizations" END AS organization
FROM workspaces
JOIN organizations USING (organization_id)
WHERE workspaces.name = $2
AND organizations.name = $3;`

type FindWorkspaceByNameParams struct {
	IncludeOrganization bool
	Name                string
	OrganizationName    string
}

type FindWorkspaceByNameRow struct {
	WorkspaceID                string         `json:"workspace_id"`
	CreatedAt                  time.Time      `json:"created_at"`
	UpdatedAt                  time.Time      `json:"updated_at"`
	AllowDestroyPlan           bool           `json:"allow_destroy_plan"`
	AutoApply                  bool           `json:"auto_apply"`
	CanQueueDestroyPlan        bool           `json:"can_queue_destroy_plan"`
	Description                string         `json:"description"`
	Environment                string         `json:"environment"`
	ExecutionMode              string         `json:"execution_mode"`
	FileTriggersEnabled        bool           `json:"file_triggers_enabled"`
	GlobalRemoteState          bool           `json:"global_remote_state"`
	MigrationEnvironment       string         `json:"migration_environment"`
	Name                       string         `json:"name"`
	QueueAllRuns               bool           `json:"queue_all_runs"`
	SpeculativeEnabled         bool           `json:"speculative_enabled"`
	SourceName                 string         `json:"source_name"`
	SourceURL                  string         `json:"source_url"`
	StructuredRunOutputEnabled bool           `json:"structured_run_output_enabled"`
	TerraformVersion           string         `json:"terraform_version"`
	TriggerPrefixes            []string       `json:"trigger_prefixes"`
	WorkingDirectory           string         `json:"working_directory"`
	OrganizationID             string         `json:"organization_id"`
	Organization               *Organizations `json:"organization"`
}

// FindWorkspaceByName implements Querier.FindWorkspaceByName.
func (q *DBQuerier) FindWorkspaceByName(ctx context.Context, params FindWorkspaceByNameParams) (FindWorkspaceByNameRow, error) {
	ctx = context.WithValue(ctx, "pggen_query_name", "FindWorkspaceByName")
	row := q.conn.QueryRow(ctx, findWorkspaceByNameSQL, params.IncludeOrganization, params.Name, params.OrganizationName)
	var item FindWorkspaceByNameRow
	organizationRow := q.types.newOrganizations()
	if err := row.Scan(&item.WorkspaceID, &item.CreatedAt, &item.UpdatedAt, &item.AllowDestroyPlan, &item.AutoApply, &item.CanQueueDestroyPlan, &item.Description, &item.Environment, &item.ExecutionMode, &item.FileTriggersEnabled, &item.GlobalRemoteState, &item.MigrationEnvironment, &item.Name, &item.QueueAllRuns, &item.SpeculativeEnabled, &item.SourceName, &item.SourceURL, &item.StructuredRunOutputEnabled, &item.TerraformVersion, &item.TriggerPrefixes, &item.WorkingDirectory, &item.OrganizationID, organizationRow); err != nil {
		return item, fmt.Errorf("query FindWorkspaceByName: %w", err)
	}
	if err := organizationRow.AssignTo(&item.Organization); err != nil {
		return item, fmt.Errorf("assign FindWorkspaceByName row: %w", err)
	}
	return item, nil
}

// FindWorkspaceByNameBatch implements Querier.FindWorkspaceByNameBatch.
func (q *DBQuerier) FindWorkspaceByNameBatch(batch genericBatch, params FindWorkspaceByNameParams) {
	batch.Queue(findWorkspaceByNameSQL, params.IncludeOrganization, params.Name, params.OrganizationName)
}

// FindWorkspaceByNameScan implements Querier.FindWorkspaceByNameScan.
func (q *DBQuerier) FindWorkspaceByNameScan(results pgx.BatchResults) (FindWorkspaceByNameRow, error) {
	row := results.QueryRow()
	var item FindWorkspaceByNameRow
	organizationRow := q.types.newOrganizations()
	if err := row.Scan(&item.WorkspaceID, &item.CreatedAt, &item.UpdatedAt, &item.AllowDestroyPlan, &item.AutoApply, &item.CanQueueDestroyPlan, &item.Description, &item.Environment, &item.ExecutionMode, &item.FileTriggersEnabled, &item.GlobalRemoteState, &item.MigrationEnvironment, &item.Name, &item.QueueAllRuns, &item.SpeculativeEnabled, &item.SourceName, &item.SourceURL, &item.StructuredRunOutputEnabled, &item.TerraformVersion, &item.TriggerPrefixes, &item.WorkingDirectory, &item.OrganizationID, organizationRow); err != nil {
		return item, fmt.Errorf("scan FindWorkspaceByNameBatch row: %w", err)
	}
	if err := organizationRow.AssignTo(&item.Organization); err != nil {
		return item, fmt.Errorf("assign FindWorkspaceByName row: %w", err)
	}
	return item, nil
}

const findWorkspaceByNameForUpdateSQL = `SELECT workspaces.*
FROM workspaces
JOIN organizations USING (organization_id)
WHERE workspaces.name = $1
AND organizations.name = $2
FOR UPDATE;`

type FindWorkspaceByNameForUpdateRow struct {
	WorkspaceID                string    `json:"workspace_id"`
	CreatedAt                  time.Time `json:"created_at"`
	UpdatedAt                  time.Time `json:"updated_at"`
	AllowDestroyPlan           bool      `json:"allow_destroy_plan"`
	AutoApply                  bool      `json:"auto_apply"`
	CanQueueDestroyPlan        bool      `json:"can_queue_destroy_plan"`
	Description                string    `json:"description"`
	Environment                string    `json:"environment"`
	ExecutionMode              string    `json:"execution_mode"`
	FileTriggersEnabled        bool      `json:"file_triggers_enabled"`
	GlobalRemoteState          bool      `json:"global_remote_state"`
	MigrationEnvironment       string    `json:"migration_environment"`
	Name                       string    `json:"name"`
	QueueAllRuns               bool      `json:"queue_all_runs"`
	SpeculativeEnabled         bool      `json:"speculative_enabled"`
	SourceName                 string    `json:"source_name"`
	SourceURL                  string    `json:"source_url"`
	StructuredRunOutputEnabled bool      `json:"structured_run_output_enabled"`
	TerraformVersion           string    `json:"terraform_version"`
	TriggerPrefixes            []string  `json:"trigger_prefixes"`
	WorkingDirectory           string    `json:"working_directory"`
	OrganizationID             string    `json:"organization_id"`
}

// FindWorkspaceByNameForUpdate implements Querier.FindWorkspaceByNameForUpdate.
func (q *DBQuerier) FindWorkspaceByNameForUpdate(ctx context.Context, name string, organizationName string) (FindWorkspaceByNameForUpdateRow, error) {
	ctx = context.WithValue(ctx, "pggen_query_name", "FindWorkspaceByNameForUpdate")
	row := q.conn.QueryRow(ctx, findWorkspaceByNameForUpdateSQL, name, organizationName)
	var item FindWorkspaceByNameForUpdateRow
	if err := row.Scan(&item.WorkspaceID, &item.CreatedAt, &item.UpdatedAt, &item.AllowDestroyPlan, &item.AutoApply, &item.CanQueueDestroyPlan, &item.Description, &item.Environment, &item.ExecutionMode, &item.FileTriggersEnabled, &item.GlobalRemoteState, &item.MigrationEnvironment, &item.Name, &item.QueueAllRuns, &item.SpeculativeEnabled, &item.SourceName, &item.SourceURL, &item.StructuredRunOutputEnabled, &item.TerraformVersion, &item.TriggerPrefixes, &item.WorkingDirectory, &item.OrganizationID); err != nil {
		return item, fmt.Errorf("query FindWorkspaceByNameForUpdate: %w", err)
	}
	return item, nil
}

// FindWorkspaceByNameForUpdateBatch implements Querier.FindWorkspaceByNameForUpdateBatch.
func (q *DBQuerier) FindWorkspaceByNameForUpdateBatch(batch genericBatch, name string, organizationName string) {
	batch.Queue(findWorkspaceByNameForUpdateSQL, name, organizationName)
}

// FindWorkspaceByNameForUpdateScan implements Querier.FindWorkspaceByNameForUpdateScan.
func (q *DBQuerier) FindWorkspaceByNameForUpdateScan(results pgx.BatchResults) (FindWorkspaceByNameForUpdateRow, error) {
	row := results.QueryRow()
	var item FindWorkspaceByNameForUpdateRow
	if err := row.Scan(&item.WorkspaceID, &item.CreatedAt, &item.UpdatedAt, &item.AllowDestroyPlan, &item.AutoApply, &item.CanQueueDestroyPlan, &item.Description, &item.Environment, &item.ExecutionMode, &item.FileTriggersEnabled, &item.GlobalRemoteState, &item.MigrationEnvironment, &item.Name, &item.QueueAllRuns, &item.SpeculativeEnabled, &item.SourceName, &item.SourceURL, &item.StructuredRunOutputEnabled, &item.TerraformVersion, &item.TriggerPrefixes, &item.WorkingDirectory, &item.OrganizationID); err != nil {
		return item, fmt.Errorf("scan FindWorkspaceByNameForUpdateBatch row: %w", err)
	}
	return item, nil
}

const findWorkspaceByIDSQL = `SELECT
    w.*,
    (users.*)::"users" AS user_lock,
    (runs.run_id)::"runs" AS run_lock,
    CASE WHEN $1 THEN (organizations.*)::"organizations" END AS organization
FROM workspaces w
JOIN organizations USING (organization_id)
LEFT JOIN (user_locks ul JOIN users USING (user_id)) ON w.workspace_id = ul.workspace_id
LEFT JOIN (run_locks rl JOIN runs USING (run_id)) ON w.workspace_id = rl.workspace_id
WHERE w.workspace_id = $2;`

type FindWorkspaceByIDRow struct {
	WorkspaceID                string         `json:"workspace_id"`
	CreatedAt                  time.Time      `json:"created_at"`
	UpdatedAt                  time.Time      `json:"updated_at"`
	AllowDestroyPlan           bool           `json:"allow_destroy_plan"`
	AutoApply                  bool           `json:"auto_apply"`
	CanQueueDestroyPlan        bool           `json:"can_queue_destroy_plan"`
	Description                string         `json:"description"`
	Environment                string         `json:"environment"`
	ExecutionMode              string         `json:"execution_mode"`
	FileTriggersEnabled        bool           `json:"file_triggers_enabled"`
	GlobalRemoteState          bool           `json:"global_remote_state"`
	MigrationEnvironment       string         `json:"migration_environment"`
	Name                       string         `json:"name"`
	QueueAllRuns               bool           `json:"queue_all_runs"`
	SpeculativeEnabled         bool           `json:"speculative_enabled"`
	SourceName                 string         `json:"source_name"`
	SourceURL                  string         `json:"source_url"`
	StructuredRunOutputEnabled bool           `json:"structured_run_output_enabled"`
	TerraformVersion           string         `json:"terraform_version"`
	TriggerPrefixes            []string       `json:"trigger_prefixes"`
	WorkingDirectory           string         `json:"working_directory"`
	OrganizationID             string         `json:"organization_id"`
	UserLock                   *Users         `json:"user_lock"`
	RunLock                    *Runs          `json:"run_lock"`
	Organization               *Organizations `json:"organization"`
}

// FindWorkspaceByID implements Querier.FindWorkspaceByID.
func (q *DBQuerier) FindWorkspaceByID(ctx context.Context, includeOrganization bool, id string) (FindWorkspaceByIDRow, error) {
	ctx = context.WithValue(ctx, "pggen_query_name", "FindWorkspaceByID")
	row := q.conn.QueryRow(ctx, findWorkspaceByIDSQL, includeOrganization, id)
	var item FindWorkspaceByIDRow
	userLockRow := q.types.newUsers()
	runLockRow := q.types.newRuns()
	organizationRow := q.types.newOrganizations()
	if err := row.Scan(&item.WorkspaceID, &item.CreatedAt, &item.UpdatedAt, &item.AllowDestroyPlan, &item.AutoApply, &item.CanQueueDestroyPlan, &item.Description, &item.Environment, &item.ExecutionMode, &item.FileTriggersEnabled, &item.GlobalRemoteState, &item.MigrationEnvironment, &item.Name, &item.QueueAllRuns, &item.SpeculativeEnabled, &item.SourceName, &item.SourceURL, &item.StructuredRunOutputEnabled, &item.TerraformVersion, &item.TriggerPrefixes, &item.WorkingDirectory, &item.OrganizationID, userLockRow, runLockRow, organizationRow); err != nil {
		return item, fmt.Errorf("query FindWorkspaceByID: %w", err)
	}
	if err := userLockRow.AssignTo(&item.UserLock); err != nil {
		return item, fmt.Errorf("assign FindWorkspaceByID row: %w", err)
	}
	if err := runLockRow.AssignTo(&item.RunLock); err != nil {
		return item, fmt.Errorf("assign FindWorkspaceByID row: %w", err)
	}
	if err := organizationRow.AssignTo(&item.Organization); err != nil {
		return item, fmt.Errorf("assign FindWorkspaceByID row: %w", err)
	}
	return item, nil
}

// FindWorkspaceByIDBatch implements Querier.FindWorkspaceByIDBatch.
func (q *DBQuerier) FindWorkspaceByIDBatch(batch genericBatch, includeOrganization bool, id string) {
	batch.Queue(findWorkspaceByIDSQL, includeOrganization, id)
}

// FindWorkspaceByIDScan implements Querier.FindWorkspaceByIDScan.
func (q *DBQuerier) FindWorkspaceByIDScan(results pgx.BatchResults) (FindWorkspaceByIDRow, error) {
	row := results.QueryRow()
	var item FindWorkspaceByIDRow
	userLockRow := q.types.newUsers()
	runLockRow := q.types.newRuns()
	organizationRow := q.types.newOrganizations()
	if err := row.Scan(&item.WorkspaceID, &item.CreatedAt, &item.UpdatedAt, &item.AllowDestroyPlan, &item.AutoApply, &item.CanQueueDestroyPlan, &item.Description, &item.Environment, &item.ExecutionMode, &item.FileTriggersEnabled, &item.GlobalRemoteState, &item.MigrationEnvironment, &item.Name, &item.QueueAllRuns, &item.SpeculativeEnabled, &item.SourceName, &item.SourceURL, &item.StructuredRunOutputEnabled, &item.TerraformVersion, &item.TriggerPrefixes, &item.WorkingDirectory, &item.OrganizationID, userLockRow, runLockRow, organizationRow); err != nil {
		return item, fmt.Errorf("scan FindWorkspaceByIDBatch row: %w", err)
	}
	if err := userLockRow.AssignTo(&item.UserLock); err != nil {
		return item, fmt.Errorf("assign FindWorkspaceByID row: %w", err)
	}
	if err := runLockRow.AssignTo(&item.RunLock); err != nil {
		return item, fmt.Errorf("assign FindWorkspaceByID row: %w", err)
	}
	if err := organizationRow.AssignTo(&item.Organization); err != nil {
		return item, fmt.Errorf("assign FindWorkspaceByID row: %w", err)
	}
	return item, nil
}

const findWorkspaceByIDForUpdateSQL = `SELECT workspaces.*
FROM workspaces
JOIN organizations USING (organization_id)
WHERE workspaces.workspace_id = $1
FOR UPDATE;`

type FindWorkspaceByIDForUpdateRow struct {
	WorkspaceID                string    `json:"workspace_id"`
	CreatedAt                  time.Time `json:"created_at"`
	UpdatedAt                  time.Time `json:"updated_at"`
	AllowDestroyPlan           bool      `json:"allow_destroy_plan"`
	AutoApply                  bool      `json:"auto_apply"`
	CanQueueDestroyPlan        bool      `json:"can_queue_destroy_plan"`
	Description                string    `json:"description"`
	Environment                string    `json:"environment"`
	ExecutionMode              string    `json:"execution_mode"`
	FileTriggersEnabled        bool      `json:"file_triggers_enabled"`
	GlobalRemoteState          bool      `json:"global_remote_state"`
	MigrationEnvironment       string    `json:"migration_environment"`
	Name                       string    `json:"name"`
	QueueAllRuns               bool      `json:"queue_all_runs"`
	SpeculativeEnabled         bool      `json:"speculative_enabled"`
	SourceName                 string    `json:"source_name"`
	SourceURL                  string    `json:"source_url"`
	StructuredRunOutputEnabled bool      `json:"structured_run_output_enabled"`
	TerraformVersion           string    `json:"terraform_version"`
	TriggerPrefixes            []string  `json:"trigger_prefixes"`
	WorkingDirectory           string    `json:"working_directory"`
	OrganizationID             string    `json:"organization_id"`
}

// FindWorkspaceByIDForUpdate implements Querier.FindWorkspaceByIDForUpdate.
func (q *DBQuerier) FindWorkspaceByIDForUpdate(ctx context.Context, id string) (FindWorkspaceByIDForUpdateRow, error) {
	ctx = context.WithValue(ctx, "pggen_query_name", "FindWorkspaceByIDForUpdate")
	row := q.conn.QueryRow(ctx, findWorkspaceByIDForUpdateSQL, id)
	var item FindWorkspaceByIDForUpdateRow
	if err := row.Scan(&item.WorkspaceID, &item.CreatedAt, &item.UpdatedAt, &item.AllowDestroyPlan, &item.AutoApply, &item.CanQueueDestroyPlan, &item.Description, &item.Environment, &item.ExecutionMode, &item.FileTriggersEnabled, &item.GlobalRemoteState, &item.MigrationEnvironment, &item.Name, &item.QueueAllRuns, &item.SpeculativeEnabled, &item.SourceName, &item.SourceURL, &item.StructuredRunOutputEnabled, &item.TerraformVersion, &item.TriggerPrefixes, &item.WorkingDirectory, &item.OrganizationID); err != nil {
		return item, fmt.Errorf("query FindWorkspaceByIDForUpdate: %w", err)
	}
	return item, nil
}

// FindWorkspaceByIDForUpdateBatch implements Querier.FindWorkspaceByIDForUpdateBatch.
func (q *DBQuerier) FindWorkspaceByIDForUpdateBatch(batch genericBatch, id string) {
	batch.Queue(findWorkspaceByIDForUpdateSQL, id)
}

// FindWorkspaceByIDForUpdateScan implements Querier.FindWorkspaceByIDForUpdateScan.
func (q *DBQuerier) FindWorkspaceByIDForUpdateScan(results pgx.BatchResults) (FindWorkspaceByIDForUpdateRow, error) {
	row := results.QueryRow()
	var item FindWorkspaceByIDForUpdateRow
	if err := row.Scan(&item.WorkspaceID, &item.CreatedAt, &item.UpdatedAt, &item.AllowDestroyPlan, &item.AutoApply, &item.CanQueueDestroyPlan, &item.Description, &item.Environment, &item.ExecutionMode, &item.FileTriggersEnabled, &item.GlobalRemoteState, &item.MigrationEnvironment, &item.Name, &item.QueueAllRuns, &item.SpeculativeEnabled, &item.SourceName, &item.SourceURL, &item.StructuredRunOutputEnabled, &item.TerraformVersion, &item.TriggerPrefixes, &item.WorkingDirectory, &item.OrganizationID); err != nil {
		return item, fmt.Errorf("scan FindWorkspaceByIDForUpdateBatch row: %w", err)
	}
	return item, nil
}

const updateWorkspaceByIDSQL = `UPDATE workspaces
SET
    allow_destroy_plan = $1,
    description = $2,
    execution_mode = $3,
    name = $4,
    queue_all_runs = $5,
    speculative_enabled = $6,
    structured_run_output_enabled = $7,
    terraform_version = $8,
    trigger_prefixes = $9,
    working_directory = $10,
    updated_at = $11
WHERE workspace_id = $12
RETURNING workspace_id;`

type UpdateWorkspaceByIDParams struct {
	AllowDestroyPlan           bool
	Description                string
	ExecutionMode              string
	Name                       string
	QueueAllRuns               bool
	SpeculativeEnabled         bool
	StructuredRunOutputEnabled bool
	TerraformVersion           string
	TriggerPrefixes            []string
	WorkingDirectory           string
	UpdatedAt                  time.Time
	ID                         string
}

// UpdateWorkspaceByID implements Querier.UpdateWorkspaceByID.
func (q *DBQuerier) UpdateWorkspaceByID(ctx context.Context, params UpdateWorkspaceByIDParams) (string, error) {
	ctx = context.WithValue(ctx, "pggen_query_name", "UpdateWorkspaceByID")
	row := q.conn.QueryRow(ctx, updateWorkspaceByIDSQL, params.AllowDestroyPlan, params.Description, params.ExecutionMode, params.Name, params.QueueAllRuns, params.SpeculativeEnabled, params.StructuredRunOutputEnabled, params.TerraformVersion, params.TriggerPrefixes, params.WorkingDirectory, params.UpdatedAt, params.ID)
	var item string
	if err := row.Scan(&item); err != nil {
		return item, fmt.Errorf("query UpdateWorkspaceByID: %w", err)
	}
	return item, nil
}

// UpdateWorkspaceByIDBatch implements Querier.UpdateWorkspaceByIDBatch.
func (q *DBQuerier) UpdateWorkspaceByIDBatch(batch genericBatch, params UpdateWorkspaceByIDParams) {
	batch.Queue(updateWorkspaceByIDSQL, params.AllowDestroyPlan, params.Description, params.ExecutionMode, params.Name, params.QueueAllRuns, params.SpeculativeEnabled, params.StructuredRunOutputEnabled, params.TerraformVersion, params.TriggerPrefixes, params.WorkingDirectory, params.UpdatedAt, params.ID)
}

// UpdateWorkspaceByIDScan implements Querier.UpdateWorkspaceByIDScan.
func (q *DBQuerier) UpdateWorkspaceByIDScan(results pgx.BatchResults) (string, error) {
	row := results.QueryRow()
	var item string
	if err := row.Scan(&item); err != nil {
		return item, fmt.Errorf("scan UpdateWorkspaceByIDBatch row: %w", err)
	}
	return item, nil
}

const deleteWorkspaceByIDSQL = `DELETE
FROM workspaces
WHERE workspace_id = $1;`

// DeleteWorkspaceByID implements Querier.DeleteWorkspaceByID.
func (q *DBQuerier) DeleteWorkspaceByID(ctx context.Context, workspaceID string) (pgconn.CommandTag, error) {
	ctx = context.WithValue(ctx, "pggen_query_name", "DeleteWorkspaceByID")
	cmdTag, err := q.conn.Exec(ctx, deleteWorkspaceByIDSQL, workspaceID)
	if err != nil {
		return cmdTag, fmt.Errorf("exec query DeleteWorkspaceByID: %w", err)
	}
	return cmdTag, err
}

// DeleteWorkspaceByIDBatch implements Querier.DeleteWorkspaceByIDBatch.
func (q *DBQuerier) DeleteWorkspaceByIDBatch(batch genericBatch, workspaceID string) {
	batch.Queue(deleteWorkspaceByIDSQL, workspaceID)
}

// DeleteWorkspaceByIDScan implements Querier.DeleteWorkspaceByIDScan.
func (q *DBQuerier) DeleteWorkspaceByIDScan(results pgx.BatchResults) (pgconn.CommandTag, error) {
	cmdTag, err := results.Exec()
	if err != nil {
		return cmdTag, fmt.Errorf("exec DeleteWorkspaceByIDBatch: %w", err)
	}
	return cmdTag, err
}

const deleteWorkspaceByNameSQL = `DELETE
FROM workspaces
USING organizations
WHERE workspaces.organization_id = organizations.organization_id
AND workspaces.name = $1
AND organizations.name = $2;`

// DeleteWorkspaceByName implements Querier.DeleteWorkspaceByName.
func (q *DBQuerier) DeleteWorkspaceByName(ctx context.Context, name string, organizationName string) (pgconn.CommandTag, error) {
	ctx = context.WithValue(ctx, "pggen_query_name", "DeleteWorkspaceByName")
	cmdTag, err := q.conn.Exec(ctx, deleteWorkspaceByNameSQL, name, organizationName)
	if err != nil {
		return cmdTag, fmt.Errorf("exec query DeleteWorkspaceByName: %w", err)
	}
	return cmdTag, err
}

// DeleteWorkspaceByNameBatch implements Querier.DeleteWorkspaceByNameBatch.
func (q *DBQuerier) DeleteWorkspaceByNameBatch(batch genericBatch, name string, organizationName string) {
	batch.Queue(deleteWorkspaceByNameSQL, name, organizationName)
}

// DeleteWorkspaceByNameScan implements Querier.DeleteWorkspaceByNameScan.
func (q *DBQuerier) DeleteWorkspaceByNameScan(results pgx.BatchResults) (pgconn.CommandTag, error) {
	cmdTag, err := results.Exec()
	if err != nil {
		return cmdTag, fmt.Errorf("exec DeleteWorkspaceByNameBatch: %w", err)
	}
	return cmdTag, err
}
