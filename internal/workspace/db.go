package workspace

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/authz"
	"github.com/leg100/otf/internal/connections"
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
	_, err := db.Exec(ctx, `
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
	return err
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
			_, err := db.Exec(ctx, `
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
	(rc.*)::"repo_connections" AS connection
FROM workspaces w
LEFT JOIN runs r ON w.latest_run_id = r.run_id
LEFT JOIN repo_connections rc ON w.workspace_id = rc.workspace_id
WHERE w.workspace_id = $1
FOR UPDATE OF w
`,
		workspaceID)
	return sql.CollectOneRow(row, db.scan)
}

// setLatestRun sets the ID of the current run for the specified workspace.
func (db *pgdb) setLatestRun(ctx context.Context, workspaceID, runID resource.TfeID) (*Workspace, error) {
	_, err := db.Exec(ctx, `
UPDATE workspaces
SET latest_run_id = $1
WHERE workspace_id = $2
`,
		&runID,
		workspaceID,
	)
	if err != nil {
		return nil, err
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
	// Status is optional.
	var status []string
	if len(opts.Status) > 0 {
		status = internal.ToStringSlice(opts.Status)
	}

	rows := db.Query(ctx, `
WITH tags_grouped_by_workspace AS (
	SELECT array_agg(t.name)::text[] AS tags, workspace_id
	FROM tags t
	JOIN workspace_tags wt USING (tag_id)
	JOIN workspaces w USING (workspace_id)
	GROUP BY wt.workspace_id
)
SELECT
    w.workspace_id, w.created_at, w.updated_at, w.allow_destroy_plan, w.auto_apply, w.can_queue_destroy_plan, w.description, w.environment, w.execution_mode, w.global_remote_state, w.migration_environment, w.name, w.queue_all_runs, w.speculative_enabled, w.source_name, w.source_url, w.structured_run_output_enabled, w.terraform_version, w.trigger_prefixes, w.working_directory, w.lock_run_id, w.latest_run_id, w.organization_name, w.branch, w.current_state_version_id, w.trigger_patterns, w.vcs_tags_regex, w.allow_cli_apply, w.agent_pool_id, w.lock_user_id,
	t.tags,
    r.status AS latest_run_status,
	(rc.*)::"repo_connections" AS connection
FROM workspaces w
LEFT JOIN runs r ON w.latest_run_id = r.run_id
LEFT JOIN repo_connections rc ON w.workspace_id = rc.workspace_id
LEFT JOIN tags_grouped_by_workspace t ON t.workspace_id = w.workspace_id
WHERE w.name                LIKE '%' || @search || '%'
AND   w.organization_name   LIKE ANY(@organization::text[])
AND   ((@status::text[] IS NULL) OR (r.status = ANY(@status::text[])))
AND   ((@tags::text[] IS NULL) OR (t.tags @> @tags::text[]))
ORDER BY w.name ASC
LIMIT @limit::int
OFFSET @offset::int
`, pgx.NamedArgs{
		"search":       opts.Search,
		"organization": []string{organization},
		"status":       status,
		"tags":         opts.Tags,
		"limit":        sql.GetLimit(opts.PageOptions),
		"offset":       sql.GetOffset(opts.PageOptions),
	})
	items, err := sql.CollectRows(rows, db.scan)
	if err != nil {
		return nil, fmt.Errorf("listing workspaces: %w", err)
	}

	count, err := db.Int(ctx, `
WITH tags_grouped_by_workspace AS (
	SELECT array_agg(t.name)::text[] AS tags, workspace_id
	FROM tags t
	JOIN workspace_tags wt USING (tag_id)
	JOIN workspaces w USING (workspace_id)
	GROUP BY wt.workspace_id
),
workspaces AS (
	SELECT w.workspace_id
	FROM workspaces w
	LEFT JOIN tags_grouped_by_workspace t ON t.workspace_id = w.workspace_id
	LEFT JOIN runs r ON w.latest_run_id = r.run_id
	WHERE w.name              LIKE '%' || @search || '%'
	AND   w.organization_name LIKE ANY(@organization::text[])
	AND ((@status::text[] IS NULL) OR (r.status = ANY(@status::text[])))
	AND ((@tags::text[] IS NULL) OR (t.tags @> @tags::text[]))
)
SELECT count(*)
FROM workspaces
`, pgx.NamedArgs{
		"search":       opts.Search,
		"organization": []string{organization},
		"status":       status,
		"tags":         opts.Tags,
	})
	if err != nil {
		return nil, fmt.Errorf("counting workspaces: %w", err)
	}
	return resource.NewPage(items, opts.PageOptions, &count), nil
}

func (db *pgdb) listByConnection(ctx context.Context, vcsProviderID resource.TfeID, repoPath string) ([]*Workspace, error) {
	rows := db.Query(ctx, `
SELECT
    w.workspace_id, w.created_at, w.updated_at, w.allow_destroy_plan, w.auto_apply, w.can_queue_destroy_plan, w.description, w.environment, w.execution_mode, w.global_remote_state, w.migration_environment, w.name, w.queue_all_runs, w.speculative_enabled, w.source_name, w.source_url, w.structured_run_output_enabled, w.terraform_version, w.trigger_prefixes, w.working_directory, w.lock_run_id, w.latest_run_id, w.organization_name, w.branch, w.current_state_version_id, w.trigger_patterns, w.vcs_tags_regex, w.allow_cli_apply, w.agent_pool_id, w.lock_user_id,
    (
        SELECT array_agg(name)::text[]
        FROM tags
        JOIN workspace_tags wt USING (tag_id)
        WHERE wt.workspace_id = w.workspace_id
    ) AS tags,
    r.status AS latest_run_status,
	(rc.*)::"repo_connections" AS connection
FROM workspaces w
LEFT JOIN runs r ON w.latest_run_id = r.run_id
JOIN repo_connections rc ON w.workspace_id = rc.workspace_id
WHERE rc.vcs_provider_id = $1
AND   rc.repo_path = $2
`,

		vcsProviderID,
		repoPath,
	)
	items, err := sql.CollectRows(rows, db.scan)
	if err != nil {
		return nil, err
	}
	return items, nil
}

func (db *pgdb) listByUsername(ctx context.Context, username string, organization resource.OrganizationName, opts resource.PageOptions) (*resource.Page[*Workspace], error) {
	rows := db.Query(ctx, `
SELECT
    w.workspace_id, w.created_at, w.updated_at, w.allow_destroy_plan, w.auto_apply, w.can_queue_destroy_plan, w.description, w.environment, w.execution_mode, w.global_remote_state, w.migration_environment, w.name, w.queue_all_runs, w.speculative_enabled, w.source_name, w.source_url, w.structured_run_output_enabled, w.terraform_version, w.trigger_prefixes, w.working_directory, w.lock_run_id, w.latest_run_id, w.organization_name, w.branch, w.current_state_version_id, w.trigger_patterns, w.vcs_tags_regex, w.allow_cli_apply, w.agent_pool_id, w.lock_user_id,
    (
        SELECT array_agg(name)::text[]
        FROM tags
        JOIN workspace_tags wt USING (tag_id)
        WHERE wt.workspace_id = w.workspace_id
    ) AS tags,
    r.status AS latest_run_status,
	(rc.*)::"repo_connections" AS connection
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
	items, err := sql.CollectRows(rows, db.scan)
	if err != nil {
		return nil, err
	}

	count, err := db.Int(ctx, `
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
	if err != nil {
		return nil, err
	}
	return resource.NewPage(items, opts, &count), nil
}

func (db *pgdb) get(ctx context.Context, workspaceID resource.TfeID) (*Workspace, error) {
	row := db.Query(ctx, `
SELECT
    w.workspace_id, w.created_at, w.updated_at, w.allow_destroy_plan, w.auto_apply, w.can_queue_destroy_plan, w.description, w.environment, w.execution_mode, w.global_remote_state, w.migration_environment, w.name, w.queue_all_runs, w.speculative_enabled, w.source_name, w.source_url, w.structured_run_output_enabled, w.terraform_version, w.trigger_prefixes, w.working_directory, w.lock_run_id, w.latest_run_id, w.organization_name, w.branch, w.current_state_version_id, w.trigger_patterns, w.vcs_tags_regex, w.allow_cli_apply, w.agent_pool_id, w.lock_user_id,
    (

        SELECT array_agg(name)::text[]
        FROM tags
        JOIN workspace_tags wt USING (tag_id)
        WHERE wt.workspace_id = w.workspace_id
    ) AS tags,
    r.status AS latest_run_status,
	(rc.*)::"repo_connections" AS connection
FROM workspaces w
LEFT JOIN runs r ON w.latest_run_id = r.run_id
LEFT JOIN repo_connections rc ON w.workspace_id = rc.workspace_id
WHERE w.workspace_id = $1
`,
		workspaceID)
	return sql.CollectOneRow(row, db.scan)
}

func (db *pgdb) getByName(ctx context.Context, organization resource.OrganizationName, workspace string) (*Workspace, error) {
	row := db.Query(ctx, `
SELECT
    w.workspace_id, w.created_at, w.updated_at, w.allow_destroy_plan, w.auto_apply, w.can_queue_destroy_plan, w.description, w.environment, w.execution_mode, w.global_remote_state, w.migration_environment, w.name, w.queue_all_runs, w.speculative_enabled, w.source_name, w.source_url, w.structured_run_output_enabled, w.terraform_version, w.trigger_prefixes, w.working_directory, w.lock_run_id, w.latest_run_id, w.organization_name, w.branch, w.current_state_version_id, w.trigger_patterns, w.vcs_tags_regex, w.allow_cli_apply, w.agent_pool_id, w.lock_user_id,
    (
        SELECT array_agg(name)::text[]
        FROM tags
        JOIN workspace_tags wt USING (tag_id)
        WHERE wt.workspace_id = w.workspace_id
    ) AS tags,
    r.status AS latest_run_status,
	(rc.*)::"repo_connections" AS connection
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
	return sql.CollectOneRow(row, db.scan)
}

func (db *pgdb) delete(ctx context.Context, workspaceID resource.TfeID) error {
	_, err := db.Exec(ctx, `
DELETE
FROM workspaces
WHERE workspace_id = $1
`,
		workspaceID)
	if err != nil {
		return err
	}
	return nil
}

func (db *pgdb) SetWorkspacePermission(ctx context.Context, workspaceID, teamID resource.TfeID, role authz.Role) error {
	_, err := db.Exec(ctx, `
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
		return err
	}
	return nil
}

func (db *pgdb) UnsetWorkspacePermission(ctx context.Context, workspaceID, teamID resource.TfeID) error {
	_, err := db.Exec(ctx, `
DELETE
FROM workspace_permissions
WHERE workspace_id = $1
AND team_id = $2
`,
		workspaceID,
		teamID,
	)
	if err != nil {
		return err
	}
	return nil
}

func (db *pgdb) GetWorkspacePolicy(ctx context.Context, workspaceID resource.TfeID) (authz.WorkspacePolicy, error) {

	row := db.QueryRow(ctx, `
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
	type workspacePermissionModel struct {
		WorkspaceID resource.TfeID `db:"workspace_id"`
		TeamID      resource.TfeID `db:"team_id"`
		Role        string
	}
	var (
		globalRemoteState bool
		perms             []workspacePermissionModel
	)
	if err := row.Scan(&globalRemoteState, &perms); err != nil {
		return authz.WorkspacePolicy{}, err
	}
	p := authz.WorkspacePolicy{
		GlobalRemoteState: globalRemoteState,
		Permissions:       make([]authz.WorkspacePermission, len(perms)),
	}
	for i, perm := range perms {
		role, err := authz.WorkspaceRoleFromString(perm.Role)
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
	type model struct {
		ID                         resource.TfeID            `db:"workspace_id"`
		CreatedAt                  time.Time                 `db:"created_at"`
		UpdatedAt                  time.Time                 `db:"updated_at"`
		AgentPoolID                *resource.TfeID           `db:"agent_pool_id"`
		AllowDestroyPlan           bool                      `db:"allow_destroy_plan"`
		AllowCLIApply              bool                      `db:"allow_cli_apply"`
		AutoApply                  bool                      `db:"auto_apply"`
		Branch                     string                    `db:"branch"`
		CanQueueDestroyPlan        bool                      `db:"can_queue_destroy_plan"`
		Description                string                    `db:"description"`
		Environment                string                    `db:"environment"`
		ExecutionMode              ExecutionMode             `db:"execution_mode"`
		GlobalRemoteState          bool                      `db:"global_remote_state"`
		MigrationEnvironment       string                    `db:"migration_environment"`
		Name                       string                    `db:"name"`
		QueueAllRuns               bool                      `db:"queue_all_runs"`
		SpeculativeEnabled         bool                      `db:"speculative_enabled"`
		StructuredRunOutputEnabled bool                      `db:"structured_run_output_enabled"`
		SourceName                 string                    `db:"source_name"`
		SourceURL                  string                    `db:"source_url"`
		TerraformVersion           string                    `db:"terraform_version"`
		WorkingDirectory           string                    `db:"working_directory"`
		Organization               resource.OrganizationName `db:"organization_name"`
		LatestRunStatus            *runstatus.Status         `db:"latest_run_status"`
		LatestRunID                *resource.TfeID           `db:"latest_run_id"`
		Tags                       []string                  `db:"tags"`
		TriggerPatterns            []string                  `db:"trigger_patterns"`
		TriggerPrefixes            []string                  `db:"trigger_prefixes"`
		VCSTagsRegex               *string                   `db:"vcs_tags_regex"`
		LockUserID                 *resource.TfeID           `db:"lock_user_id"`
		LockRunID                  *resource.TfeID           `db:"lock_run_id"`
		CurrentStateVersionID      *resource.TfeID           `db:"current_state_version_id"`
		Connection                 *connections.Connection
	}
	m, err := pgx.RowToStructByName[model](scanner)
	if err != nil {
		return nil, err
	}
	ws := &Workspace{
		ID:                         m.ID,
		CreatedAt:                  m.CreatedAt,
		UpdatedAt:                  m.UpdatedAt,
		AgentPoolID:                m.AgentPoolID,
		AllowDestroyPlan:           m.AllowDestroyPlan,
		AutoApply:                  m.AutoApply,
		CanQueueDestroyPlan:        m.CanQueueDestroyPlan,
		Description:                m.Description,
		Environment:                m.Environment,
		ExecutionMode:              m.ExecutionMode,
		GlobalRemoteState:          m.GlobalRemoteState,
		MigrationEnvironment:       m.MigrationEnvironment,
		Name:                       m.Name,
		QueueAllRuns:               m.QueueAllRuns,
		SpeculativeEnabled:         m.SpeculativeEnabled,
		StructuredRunOutputEnabled: m.StructuredRunOutputEnabled,
		SourceName:                 m.SourceName,
		SourceURL:                  m.SourceURL,
		TerraformVersion:           m.TerraformVersion,
		WorkingDirectory:           m.WorkingDirectory,
		Organization:               m.Organization,
		Tags:                       m.Tags,
		TriggerPatterns:            m.TriggerPatterns,
		TriggerPrefixes:            m.TriggerPrefixes,
	}
	if m.Connection != nil {
		ws.Connection = &Connection{
			AllowCLIApply: m.AllowCLIApply,
			VCSProviderID: m.Connection.VCSProviderID,
			Repo:          m.Connection.Repo,
			Branch:        m.Branch,
		}
		if m.VCSTagsRegex != nil {
			ws.Connection.TagsRegex = *m.VCSTagsRegex
		}
	}
	if m.LatestRunID != nil && m.LatestRunStatus != nil {
		ws.LatestRun = &LatestRun{
			ID:     *m.LatestRunID,
			Status: *m.LatestRunStatus,
		}
	}

	if m.LockUserID != nil {
		ws.Lock = m.LockUserID
	} else if m.LockRunID != nil {
		ws.Lock = m.LockRunID
	}
	return ws, err
}
