package workspace

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/authz"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/runstatus"
	"github.com/leg100/otf/internal/sql"
)

// pgdb is a workspace database on postgres
type pgdb struct {
	*sql.DB // provides access to generated SQL queries
}

func (db *pgdb) create(ctx context.Context, ws *Workspace) error {
	var (
		allowCLIApply bool
		branch        string
		VCSTagsRegex  *string
	)
	if ws.Connection != nil {
		allowCLIApply = ws.Connection.AllowCLIApply
		branch = ws.Connection.Branch
		VCSTagsRegex = &ws.Connection.TagsRegex
	}
	_, err := db.Conn(ctx).Exec(ctx, `
INSERT INTO workspaces (
    workspace_id,
    created_at,
    updated_at,
    agent_pool_id,
    allow_cli_apply,
    allow_destroy_plan,
    auto_apply,
    branch,
    can_queue_destroy_plan,
    description,
    environment,
    execution_mode,
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
    trigger_patterns,
    vcs_tags_regex,
    working_directory,
    organization_name
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
    $22,
    $23,
    $24,
    $25,
    $26
)
`,
		ws.ID,
		ws.CreatedAt,
		ws.UpdatedAt,
		ws.AgentPoolID,
		allowCLIApply,
		ws.AllowDestroyPlan,
		ws.AutoApply,
		branch,
		ws.CanQueueDestroyPlan,
		ws.Description,
		ws.Environment,
		ws.ExecutionMode,
		ws.GlobalRemoteState,
		ws.MigrationEnvironment,
		ws.Name,
		ws.QueueAllRuns,
		ws.SpeculativeEnabled,
		ws.SourceName,
		ws.SourceURL,
		ws.StructuredRunOutputEnabled,
		ws.TerraformVersion,
		ws.TriggerPrefixes,
		ws.TriggerPatterns,
		VCSTagsRegex,
		ws.WorkingDirectory,
		ws.Organization,
	)
	return sql.Error(err)
}

func (db *pgdb) update(ctx context.Context, workspaceID resource.TfeID, fn func(context.Context, *Workspace) error) (*Workspace, error) {
	return sql.Updater(
		ctx,
		db.DB,
		func(ctx context.Context, conn sql.Connection) (*Workspace, error) {
			return db.forUpdate(ctx, conn, workspaceID)
		},
		fn,
		func(ctx context.Context, conn sql.Connection, ws *Workspace) error {
			var (
				allowCLIApply bool
				branch        string
				VCSTagsRegex  *string
			)
			if ws.Connection != nil {
				allowCLIApply = ws.Connection.AllowCLIApply
				branch = ws.Connection.Branch
				VCSTagsRegex = &ws.Connection.TagsRegex
			}
			_, err := db.Conn(ctx).Exec(ctx, `
				UPDATE workspaces
				SET
					agent_pool_id                 = $1,
					allow_destroy_plan            = $2,
					allow_cli_apply               = $3,
					auto_apply                    = $4,
					branch                        = $5,
					description                   = $6,
					execution_mode                = $7,
					global_remote_state           = $8,
					name                          = $9,
					queue_all_runs                = $10,
					speculative_enabled           = $11,
					structured_run_output_enabled = $12,
					terraform_version             = $13,
					trigger_prefixes              = $14,
					trigger_patterns              = $15,
					vcs_tags_regex                = $16,
					working_directory             = $17,
					updated_at                    = $18
				WHERE workspace_id = $19
			`,
				ws.AgentPoolID,
				ws.AllowDestroyPlan,
				allowCLIApply,
				ws.AutoApply,
				branch,
				ws.Description,
				ws.ExecutionMode,
				ws.GlobalRemoteState,
				ws.Name,
				ws.QueueAllRuns,
				ws.SpeculativeEnabled,
				ws.StructuredRunOutputEnabled,
				ws.TerraformVersion,
				ws.TriggerPrefixes,
				ws.TriggerPatterns,
				VCSTagsRegex,
				ws.WorkingDirectory,
				ws.UpdatedAt,
				ws.ID,
			)
			return err
		},
	)
}

func (db *pgdb) forUpdate(ctx context.Context, conn sql.Connection, workspaceID resource.TfeID) (*Workspace, error) {
	row, _ := conn.Query(ctx, `
SELECT
    w.workspace_id, w.created_at, w.updated_at, w.allow_destroy_plan, w.auto_apply, w.can_queue_destroy_plan, w.description, w.environment, w.execution_mode, w.global_remote_state, w.migration_environment, w.name, w.queue_all_runs, w.speculative_enabled, w.source_name, w.source_url, w.structured_run_output_enabled, w.terraform_version, w.trigger_prefixes, w.working_directory, w.lock_run_id, w.latest_run_id, w.organization_name, w.branch, w.current_state_version_id, w.trigger_patterns, w.vcs_tags_regex, w.allow_cli_apply, w.agent_pool_id, w.lock_user_id,
    (
        SELECT array_agg(name)::text[]
        FROM tags
        JOIN workspace_tags wt USING (tag_id)
        WHERE wt.workspace_id = w.workspace_id
    ) AS tags,
    r.status AS latest_run_status,
    rc.vcs_provider_id,
    rc.repo_path
FROM workspaces w
LEFT JOIN runs r ON w.latest_run_id = r.run_id
LEFT JOIN repo_connections rc ON w.workspace_id = rc.workspace_id
WHERE w.workspace_id = $1
FOR UPDATE OF w
`,
		workspaceID)
	return pgx.CollectOneRow(row, db.scan)
}

// setLatestRun sets the ID of the current run for the specified workspace.
func (db *pgdb) setLatestRun(ctx context.Context, workspaceID, runID resource.TfeID) (*Workspace, error) {
	_, err := db.Conn(ctx).Exec(ctx, `
UPDATE workspaces
SET latest_run_id = $1
WHERE workspace_id = $2
`,
		&runID,
		workspaceID,
	)
	if err != nil {
		return nil, sql.Error(err)
	}

	return db.get(ctx, workspaceID)
}

func (db *pgdb) list(ctx context.Context, opts ListOptions) (*resource.Page[*Workspace], error) {
	// Organization name filter is optional - if not provided use a % which in
	// SQL means match any organization.
	organization := "%"
	if opts.Organization != nil {
		organization = opts.Organization.String()
	}
	tags := []string{}
	if len(opts.Tags) > 0 {
		tags = opts.Tags
	}
	// Status is optional.
	var status []string
	if len(opts.Status) > 0 {
		status = internal.ToStringSlice(opts.Status)
	}

	rows, _ := db.Conn(ctx).Query(ctx, `
SELECT
    w.workspace_id, w.created_at, w.updated_at, w.allow_destroy_plan, w.auto_apply, w.can_queue_destroy_plan, w.description, w.environment, w.execution_mode, w.global_remote_state, w.migration_environment, w.name, w.queue_all_runs, w.speculative_enabled, w.source_name, w.source_url, w.structured_run_output_enabled, w.terraform_version, w.trigger_prefixes, w.working_directory, w.lock_run_id, w.latest_run_id, w.organization_name, w.branch, w.current_state_version_id, w.trigger_patterns, w.vcs_tags_regex, w.allow_cli_apply, w.agent_pool_id, w.lock_user_id,
    (
        SELECT array_agg(name)::text[]
        FROM tags
        JOIN workspace_tags wt USING (tag_id)
        WHERE wt.workspace_id = w.workspace_id
        GROUP BY wt.workspace_id
    ) AS tags,
    r.status AS latest_run_status,
    rc.vcs_provider_id,
    rc.repo_path
FROM workspaces w
LEFT JOIN runs r ON w.latest_run_id = r.run_id
LEFT JOIN repo_connections rc ON w.workspace_id = rc.workspace_id
LEFT JOIN (workspace_tags wt JOIN tags t USING (tag_id)) ON wt.workspace_id = w.workspace_id
WHERE w.name                LIKE '%' || $1 || '%'
AND   w.organization_name   LIKE ANY($2::text[])
AND   (($3::text[] IS NULL) OR (r.status = ANY($3::text[])))
GROUP BY w.workspace_id, r.status, rc.vcs_provider_id, rc.repo_path
HAVING array_agg(t.name) @> $4::text[]
ORDER BY w.name ASC
LIMIT $5::int
OFFSET $6::int
`,
		opts.Search,
		[]string{organization},
		status,
		tags,
		sql.GetLimit(opts.PageOptions),
		sql.GetOffset(opts.PageOptions),
	)
	items, err := pgx.CollectRows(rows, db.scan)
	if err != nil {
		return nil, sql.Error(err)
	}

	row := db.Conn(ctx).QueryRow(ctx, `
WITH
    workspaces AS (
        SELECT w.workspace_id
        FROM workspaces w
        LEFT JOIN (workspace_tags wt JOIN tags t USING (tag_id)) ON w.workspace_id = wt.workspace_id
		LEFT JOIN runs r ON w.latest_run_id = r.run_id
        WHERE w.name              LIKE '%' || $1 || '%'
        AND   w.organization_name LIKE ANY($2::text[])
		AND (($3::text[] IS NULL) OR (r.status = ANY($3::text[])))
        GROUP BY w.workspace_id
        HAVING array_agg(t.name) @> $4::text[]
    )
SELECT count(*)
FROM workspaces
`,
		opts.Search,
		[]string{organization},
		status,
		tags,
	)
	var count int64
	if err := row.Scan(&count); err != nil {
		return nil, fmt.Errorf("counting workspaces: %w", err)
	}
	return resource.NewPage(items, opts.PageOptions, internal.Int64(count)), nil
}

func (db *pgdb) listByConnection(ctx context.Context, vcsProviderID resource.TfeID, repoPath string) ([]*Workspace, error) {
	rows, _ := db.Conn(ctx).Query(ctx, `
SELECT
    w.workspace_id, w.created_at, w.updated_at, w.allow_destroy_plan, w.auto_apply, w.can_queue_destroy_plan, w.description, w.environment, w.execution_mode, w.global_remote_state, w.migration_environment, w.name, w.queue_all_runs, w.speculative_enabled, w.source_name, w.source_url, w.structured_run_output_enabled, w.terraform_version, w.trigger_prefixes, w.working_directory, w.lock_run_id, w.latest_run_id, w.organization_name, w.branch, w.current_state_version_id, w.trigger_patterns, w.vcs_tags_regex, w.allow_cli_apply, w.agent_pool_id, w.lock_user_id,
    (
        SELECT array_agg(name)::text[]
        FROM tags
        JOIN workspace_tags wt USING (tag_id)
        WHERE wt.workspace_id = w.workspace_id
    ) AS tags,
    r.status AS latest_run_status,
    rc.vcs_provider_id,
    rc.repo_path
FROM workspaces w
LEFT JOIN runs r ON w.latest_run_id = r.run_id
JOIN repo_connections rc ON w.workspace_id = rc.workspace_id
WHERE rc.vcs_provider_id = $1
AND   rc.repo_path = $2
`,

		vcsProviderID,
		repoPath,
	)
	items, err := pgx.CollectRows(rows, db.scan)
	if err != nil {
		return nil, err
	}
	return items, nil
}

func (db *pgdb) listByUsername(ctx context.Context, username string, organization resource.OrganizationName, opts resource.PageOptions) (*resource.Page[*Workspace], error) {
	rows, _ := db.Conn(ctx).Query(ctx, `
SELECT
    w.workspace_id, w.created_at, w.updated_at, w.allow_destroy_plan, w.auto_apply, w.can_queue_destroy_plan, w.description, w.environment, w.execution_mode, w.global_remote_state, w.migration_environment, w.name, w.queue_all_runs, w.speculative_enabled, w.source_name, w.source_url, w.structured_run_output_enabled, w.terraform_version, w.trigger_prefixes, w.working_directory, w.lock_run_id, w.latest_run_id, w.organization_name, w.branch, w.current_state_version_id, w.trigger_patterns, w.vcs_tags_regex, w.allow_cli_apply, w.agent_pool_id, w.lock_user_id,
    (
        SELECT array_agg(name)::text[]
        FROM tags
        JOIN workspace_tags wt USING (tag_id)
        WHERE wt.workspace_id = w.workspace_id
    ) AS tags,
    r.status AS latest_run_status,
    rc.vcs_provider_id,
    rc.repo_path
FROM workspaces w
JOIN workspace_permissions p USING (workspace_id)
LEFT JOIN runs r ON w.latest_run_id = r.run_id
LEFT JOIN repo_connections rc ON w.workspace_id = rc.workspace_id
JOIN teams t USING (team_id)
JOIN team_memberships tm USING (team_id)
JOIN users u ON tm.username = u.username
WHERE w.organization_name  = $1
AND   u.username           = $2
ORDER BY w.updated_at DESC
LIMIT $3::int
OFFSET $4::int
`,
		organization,
		username,
		sql.GetLimit(opts),
		sql.GetOffset(opts),
	)
	items, err := pgx.CollectRows(rows, db.scan)
	if err != nil {
		return nil, err
	}

	row := db.Conn(ctx).QueryRow(ctx, `
SELECT count(*)
FROM workspaces w
JOIN workspace_permissions p USING (workspace_id)
JOIN teams t USING (team_id)
JOIN team_memberships tm USING (team_id)
JOIN users u USING (username)
WHERE w.organization_name = $1
AND   u.username          = $2
`,
		organization,
		username,
	)
	var count int64
	if err := row.Scan(&count); err != nil {
		return nil, err
	}
	return resource.NewPage(items, opts, internal.Int64(count)), nil
}

func (db *pgdb) get(ctx context.Context, workspaceID resource.TfeID) (*Workspace, error) {
	row, _ := db.Conn(ctx).Query(ctx, `
SELECT
    w.workspace_id, w.created_at, w.updated_at, w.allow_destroy_plan, w.auto_apply, w.can_queue_destroy_plan, w.description, w.environment, w.execution_mode, w.global_remote_state, w.migration_environment, w.name, w.queue_all_runs, w.speculative_enabled, w.source_name, w.source_url, w.structured_run_output_enabled, w.terraform_version, w.trigger_prefixes, w.working_directory, w.lock_run_id, w.latest_run_id, w.organization_name, w.branch, w.current_state_version_id, w.trigger_patterns, w.vcs_tags_regex, w.allow_cli_apply, w.agent_pool_id, w.lock_user_id,
    (

        SELECT array_agg(name)::text[]
        FROM tags
        JOIN workspace_tags wt USING (tag_id)
        WHERE wt.workspace_id = w.workspace_id
    ) AS tags,
    r.status AS latest_run_status,
    rc.vcs_provider_id,
    rc.repo_path
FROM workspaces w
LEFT JOIN runs r ON w.latest_run_id = r.run_id
LEFT JOIN repo_connections rc ON w.workspace_id = rc.workspace_id
WHERE w.workspace_id = $1
`,
		workspaceID)
	return pgx.CollectOneRow(row, db.scan)
}

func (db *pgdb) getByName(ctx context.Context, organization resource.OrganizationName, workspace string) (*Workspace, error) {
	row, _ := db.Conn(ctx).Query(ctx, `
SELECT
    w.workspace_id, w.created_at, w.updated_at, w.allow_destroy_plan, w.auto_apply, w.can_queue_destroy_plan, w.description, w.environment, w.execution_mode, w.global_remote_state, w.migration_environment, w.name, w.queue_all_runs, w.speculative_enabled, w.source_name, w.source_url, w.structured_run_output_enabled, w.terraform_version, w.trigger_prefixes, w.working_directory, w.lock_run_id, w.latest_run_id, w.organization_name, w.branch, w.current_state_version_id, w.trigger_patterns, w.vcs_tags_regex, w.allow_cli_apply, w.agent_pool_id, w.lock_user_id,
    (
        SELECT array_agg(name)::text[]
        FROM tags
        JOIN workspace_tags wt USING (tag_id)
        WHERE wt.workspace_id = w.workspace_id
    ) AS tags,
    r.status AS latest_run_status,
    rc.vcs_provider_id,
    rc.repo_path
FROM workspaces w
LEFT JOIN runs r ON w.latest_run_id = r.run_id
LEFT JOIN repo_connections rc ON w.workspace_id = rc.workspace_id
LEFT JOIN (workspace_tags wt JOIN tags t USING (tag_id)) ON w.workspace_id = rc.workspace_id
WHERE w.name              = $1
AND   w.organization_name = $2
`,
		workspace,
		organization,
	)
	return pgx.CollectOneRow(row, db.scan)
}

func (db *pgdb) delete(ctx context.Context, workspaceID resource.TfeID) error {
	_, err := db.Conn(ctx).Exec(ctx, `
DELETE
FROM workspaces
WHERE workspace_id = $1
`,
		workspaceID)
	if err != nil {
		return sql.Error(err)
	}
	return nil
}

func (db *pgdb) SetWorkspacePermission(ctx context.Context, workspaceID, teamID resource.TfeID, role authz.Role) error {
	_, err := db.Conn(ctx).Exec(ctx, `
INSERT INTO workspace_permissions (
    workspace_id,
    team_id,
    role
) VALUES (
    $1,
    $2,
    $3
) ON CONFLICT (workspace_id, team_id) DO UPDATE SET role = $3
`,
		workspaceID,
		teamID,
		role.String(),
	)
	if err != nil {
		return sql.Error(err)
	}
	return nil
}

func (db *pgdb) UnsetWorkspacePermission(ctx context.Context, workspaceID, teamID resource.TfeID) error {
	_, err := db.Conn(ctx).Exec(ctx, `
DELETE
FROM workspace_permissions
WHERE workspace_id = $1
AND team_id = $2
`,
		workspaceID,
		teamID,
	)
	if err != nil {
		return sql.Error(err)
	}
	return nil
}

func (db *pgdb) GetWorkspacePolicy(ctx context.Context, workspaceID resource.TfeID) (authz.WorkspacePolicy, error) {

	row := db.Conn(ctx).QueryRow(ctx, `
SELECT
    w.global_remote_state,
    (
        SELECT array_agg(wp.*)::workspace_permissions[]
        FROM workspace_permissions wp
        WHERE wp.workspace_id = w.workspace_id
    ) AS workspace_permissions
FROM workspaces w
WHERE w.workspace_id = $1
`,
		workspaceID)
	var (
		globalRemoteState bool
		perms             []WorkspacePermission
	)
	if err := row.Scan(&globalRemoteState, &perms); err != nil {
		return authz.WorkspacePolicy{}, sql.Error(err)
	}
	p := authz.WorkspacePolicy{
		GlobalRemoteState: globalRemoteState,
		Permissions:       make([]authz.WorkspacePermission, len(perms)),
	}
	for i, perm := range perms {
		role, err := authz.WorkspaceRoleFromString(perm.Role.String)
		if err != nil {
			return authz.WorkspacePolicy{}, err
		}
		p.Permissions[i] = authz.WorkspacePermission{
			TeamID: perm.TeamID,
			Role:   role,
		}
	}
	return p, nil
}

func (db *pgdb) scan(scanner pgx.CollectableRow) (*Workspace, error) {
	var (
		ws                    Workspace
		repoPath              *string
		conn                  Connection
		tagsRegex             *string
		lockRunID             *resource.TfeID
		lockUserID            *resource.TfeID
		latestRunID           *resource.TfeID
		latestRunStatus       *runstatus.Status
		currentStateVersionID *resource.TfeID
	)
	err := scanner.Scan(
		&ws.ID,
		&ws.CreatedAt,
		&ws.UpdatedAt,
		&ws.AllowDestroyPlan,
		&ws.AutoApply,
		&ws.CanQueueDestroyPlan,
		&ws.Description,
		&ws.Environment,
		&ws.ExecutionMode,
		&ws.GlobalRemoteState,
		&ws.MigrationEnvironment,
		&ws.Name,
		&ws.QueueAllRuns,
		&ws.SpeculativeEnabled,
		&ws.SourceName,
		&ws.SourceURL,
		&ws.StructuredRunOutputEnabled,
		&ws.TerraformVersion,
		&ws.TriggerPrefixes,
		&ws.WorkingDirectory,
		&lockRunID,
		&latestRunID,
		&ws.Organization,
		&conn.Branch,
		&currentStateVersionID,
		&ws.TriggerPatterns,
		&tagsRegex,
		&conn.AllowCLIApply,
		&ws.AgentPoolID,
		&lockUserID,
		&ws.Tags,
		&latestRunStatus,
		&conn.VCSProviderID,
		&repoPath,
	)
	if repoPath != nil {
		ws.Connection = &Connection{
			AllowCLIApply: conn.AllowCLIApply,
			VCSProviderID: conn.VCSProviderID,
			Repo:          *repoPath,
			Branch:        conn.Branch,
		}
		if tagsRegex != nil {
			ws.Connection.TagsRegex = *tagsRegex
		}
	}
	if latestRunID != nil {
		ws.LatestRun = &LatestRun{
			ID:     *latestRunID,
			Status: *latestRunStatus,
		}
	}

	if lockUserID != nil {
		ws.Lock = lockUserID
	} else if lockRunID != nil {
		ws.Lock = lockRunID
	}
	ws.CreatedAt = ws.CreatedAt.UTC()
	ws.UpdatedAt = ws.UpdatedAt.UTC()
	return &ws, sql.Error(err)
}
