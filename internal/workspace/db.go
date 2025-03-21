package workspace

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/authz"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/runstatus"
	"github.com/leg100/otf/internal/sql"
)

var q = &Queries{}

type (
	// pgdb is a workspace database on postgres
	pgdb struct {
		*sql.DB // provides access to generated SQL queries
	}

	// pgresult represents the result of a database query for a workspace.
	pgresult struct {
		WorkspaceID                resource.TfeID
		CreatedAt                  pgtype.Timestamptz
		UpdatedAt                  pgtype.Timestamptz
		AllowDestroyPlan           pgtype.Bool
		AutoApply                  pgtype.Bool
		CanQueueDestroyPlan        pgtype.Bool
		Description                pgtype.Text
		Environment                pgtype.Text
		ExecutionMode              pgtype.Text
		GlobalRemoteState          pgtype.Bool
		MigrationEnvironment       pgtype.Text
		Name                       pgtype.Text
		QueueAllRuns               pgtype.Bool
		SpeculativeEnabled         pgtype.Bool
		SourceName                 pgtype.Text
		SourceURL                  pgtype.Text
		StructuredRunOutputEnabled pgtype.Bool
		TerraformVersion           pgtype.Text
		TriggerPrefixes            []pgtype.Text
		WorkingDirectory           pgtype.Text
		LockRunID                  *resource.TfeID
		LatestRunID                *resource.TfeID
		OrganizationName           resource.OrganizationName
		Branch                     pgtype.Text
		CurrentStateVersionID      *resource.TfeID
		TriggerPatterns            []pgtype.Text
		VCSTagsRegex               pgtype.Text
		AllowCLIApply              pgtype.Bool
		AgentPoolID                *resource.TfeID
		LockUserID                 *resource.TfeID
		Tags                       []pgtype.Text
		LatestRunStatus            pgtype.Text
		VCSProviderID              resource.TfeID
		RepoPath                   pgtype.Text
	}
)

func (r pgresult) toWorkspace() (*Workspace, error) {
	ws := Workspace{
		ID:                         r.WorkspaceID,
		CreatedAt:                  r.CreatedAt.Time.UTC(),
		UpdatedAt:                  r.UpdatedAt.Time.UTC(),
		AllowDestroyPlan:           r.AllowDestroyPlan.Bool,
		AutoApply:                  r.AutoApply.Bool,
		CanQueueDestroyPlan:        r.CanQueueDestroyPlan.Bool,
		Description:                r.Description.String,
		Environment:                r.Environment.String,
		ExecutionMode:              ExecutionMode(r.ExecutionMode.String),
		GlobalRemoteState:          r.GlobalRemoteState.Bool,
		MigrationEnvironment:       r.MigrationEnvironment.String,
		Name:                       r.Name.String,
		QueueAllRuns:               r.QueueAllRuns.Bool,
		SpeculativeEnabled:         r.SpeculativeEnabled.Bool,
		StructuredRunOutputEnabled: r.StructuredRunOutputEnabled.Bool,
		SourceName:                 r.SourceName.String,
		SourceURL:                  r.SourceURL.String,
		TerraformVersion:           r.TerraformVersion.String,
		TriggerPrefixes:            sql.FromStringArray(r.TriggerPrefixes),
		TriggerPatterns:            sql.FromStringArray(r.TriggerPatterns),
		WorkingDirectory:           r.WorkingDirectory.String,
		Organization:               r.OrganizationName,
		Tags:                       sql.FromStringArray(r.Tags),
		AgentPoolID:                r.AgentPoolID,
	}
	if r.RepoPath.Valid {
		ws.Connection = &Connection{
			AllowCLIApply: r.AllowCLIApply.Bool,
			VCSProviderID: r.VCSProviderID,
			Repo:          r.RepoPath.String,
			Branch:        r.Branch.String,
		}
		if r.VCSTagsRegex.Valid {
			ws.Connection.TagsRegex = r.VCSTagsRegex.String
		}
	}

	if r.LatestRunID != nil {
		ws.LatestRun = &LatestRun{
			ID:     *r.LatestRunID,
			Status: runstatus.Status(r.LatestRunStatus.String),
		}
	}

	if r.LockUserID != nil {
		ws.Lock = r.LockUserID
	} else if r.LockRunID != nil {
		ws.Lock = r.LockRunID
	}

	return &ws, nil
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
			result, err := q.FindWorkspaceByIDForUpdate(ctx, conn, workspaceID)
			if err != nil {
				return nil, err
			}
			return pgresult(result).toWorkspace()
		},
		fn,
		func(ctx context.Context, conn sql.Connection, ws *Workspace) error {
			params := UpdateWorkspaceByIDParams{
				AgentPoolID:                ws.AgentPoolID,
				AllowDestroyPlan:           sql.Bool(ws.AllowDestroyPlan),
				AllowCLIApply:              sql.Bool(false),
				AutoApply:                  sql.Bool(ws.AutoApply),
				Branch:                     sql.String(""),
				Description:                sql.String(ws.Description),
				ExecutionMode:              sql.String(string(ws.ExecutionMode)),
				GlobalRemoteState:          sql.Bool(ws.GlobalRemoteState),
				Name:                       sql.String(ws.Name),
				QueueAllRuns:               sql.Bool(ws.QueueAllRuns),
				SpeculativeEnabled:         sql.Bool(ws.SpeculativeEnabled),
				StructuredRunOutputEnabled: sql.Bool(ws.StructuredRunOutputEnabled),
				TerraformVersion:           sql.String(ws.TerraformVersion),
				TriggerPrefixes:            sql.StringArray(ws.TriggerPrefixes),
				TriggerPatterns:            sql.StringArray(ws.TriggerPatterns),
				VCSTagsRegex:               sql.StringPtr(nil),
				WorkingDirectory:           sql.String(ws.WorkingDirectory),
				UpdatedAt:                  sql.Timestamptz(ws.UpdatedAt),
				ID:                         ws.ID,
			}
			if ws.Connection != nil {
				params.AllowCLIApply = sql.Bool(ws.Connection.AllowCLIApply)
				params.Branch = sql.String(ws.Connection.Branch)
				params.VCSTagsRegex = sql.String(ws.Connection.TagsRegex)
			}
			_, err := q.UpdateWorkspaceByID(ctx, conn, params)
			return err
		},
	)
}

// setLatestRun sets the ID of the current run for the specified workspace.
func (db *pgdb) setLatestRun(ctx context.Context, workspaceID, runID resource.TfeID) (*Workspace, error) {
	err := q.UpdateWorkspaceLatestRun(ctx, db.Conn(ctx), UpdateWorkspaceLatestRunParams{
		RunID:       &runID,
		WorkspaceID: workspaceID,
	})
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
	// Status is optional - if not provided use a % which in SQL means match any
	// status.
	var status []string
	if len(opts.Status) > 0 {
		status = internal.ToStringSlice(opts.Status)
	}

	rows, err := q.FindWorkspaces(ctx, db.Conn(ctx), FindWorkspacesParams{
		OrganizationNames: sql.StringArray([]string{organization}),
		Search:            sql.String(opts.Search),
		Tags:              sql.StringArray(tags),
		Status:            sql.StringArray(status),
		Limit:             sql.GetLimit(opts.PageOptions),
		Offset:            sql.GetOffset(opts.PageOptions),
	})
	if err != nil {
		return nil, sql.Error(err)
	}
	count, err := q.CountWorkspaces(ctx, db.Conn(ctx), CountWorkspacesParams{
		Search:            sql.String(opts.Search),
		OrganizationNames: sql.StringArray([]string{organization}),
		Tags:              sql.StringArray(tags),
		Status:            sql.StringArray(status),
	})
	if err != nil {
		return nil, sql.Error(err)
	}

	items := make([]*Workspace, len(rows))
	for i, r := range rows {
		ws, err := pgresult(r).toWorkspace()
		if err != nil {
			return nil, err
		}
		items[i] = ws
	}
	return resource.NewPage(items, opts.PageOptions, internal.Int64(count)), nil
}

func (db *pgdb) listByConnection(ctx context.Context, vcsProviderID resource.TfeID, repoPath string) ([]*Workspace, error) {
	rows, err := q.FindWorkspacesByConnection(ctx, db.Conn(ctx), FindWorkspacesByConnectionParams{
		VCSProviderID: vcsProviderID,
		RepoPath:      sql.String(repoPath),
	})
	if err != nil {
		return nil, err
	}

	items := make([]*Workspace, len(rows))
	for i, r := range rows {
		ws, err := pgresult(r).toWorkspace()
		if err != nil {
			return nil, err
		}
		items[i] = ws
	}
	return items, nil
}

func (db *pgdb) listByUsername(ctx context.Context, username string, organization resource.OrganizationName, opts resource.PageOptions) (*resource.Page[*Workspace], error) {
	rows, err := q.FindWorkspacesByUsername(ctx, db.Conn(ctx), FindWorkspacesByUsernameParams{
		OrganizationName: organization,
		Username:         sql.String(username),
		Limit:            sql.GetLimit(opts),
		Offset:           sql.GetOffset(opts),
	})
	if err != nil {
		return nil, err
	}
	count, err := q.CountWorkspacesByUsername(ctx, db.Conn(ctx), CountWorkspacesByUsernameParams{
		OrganizationName: organization,
		Username:         sql.String(username),
	})
	if err != nil {
		return nil, err
	}

	items := make([]*Workspace, len(rows))
	for i, r := range rows {
		ws, err := pgresult(r).toWorkspace()
		if err != nil {
			return nil, err
		}
		items[i] = ws
	}

	return resource.NewPage(items, opts, internal.Int64(count)), nil
}

func (db *pgdb) get(ctx context.Context, workspaceID resource.TfeID) (*Workspace, error) {
	result, err := q.FindWorkspaceByID(ctx, db.Conn(ctx), workspaceID)
	if err != nil {
		return nil, sql.Error(err)
	}
	return pgresult(result).toWorkspace()
}

func (db *pgdb) getByName(ctx context.Context, organization resource.OrganizationName, workspace string) (*Workspace, error) {
	result, err := q.FindWorkspaceByName(ctx, db.Conn(ctx), FindWorkspaceByNameParams{
		Name:             sql.String(workspace),
		OrganizationName: organization,
	})
	if err != nil {
		return nil, sql.Error(err)
	}
	return pgresult(result).toWorkspace()
}

func (db *pgdb) delete(ctx context.Context, workspaceID resource.TfeID) error {
	err := q.DeleteWorkspaceByID(ctx, db.Conn(ctx), workspaceID)
	if err != nil {
		return sql.Error(err)
	}
	return nil
}

func (db *pgdb) SetWorkspacePermission(ctx context.Context, workspaceID, teamID resource.TfeID, role authz.Role) error {
	err := q.UpsertWorkspacePermission(ctx, db.Conn(ctx), UpsertWorkspacePermissionParams{
		WorkspaceID: workspaceID,
		TeamID:      teamID,
		Role:        sql.String(role.String()),
	})
	if err != nil {
		return sql.Error(err)
	}
	return nil
}

func (db *pgdb) UnsetWorkspacePermission(ctx context.Context, workspaceID, teamID resource.TfeID) error {
	err := q.DeleteWorkspacePermissionByID(ctx, db.Conn(ctx), DeleteWorkspacePermissionByIDParams{
		WorkspaceID: workspaceID,
		TeamID:      teamID,
	})
	if err != nil {
		return sql.Error(err)
	}
	return nil
}

func (db *pgdb) GetWorkspacePolicy(ctx context.Context, workspaceID resource.TfeID) (authz.WorkspacePolicy, error) {
	perms, err := q.FindWorkspacePermissionsAndGlobalRemoteState(ctx, db.Conn(ctx), workspaceID)
	if err != nil {
		return authz.WorkspacePolicy{}, sql.Error(err)
	}
	p := authz.WorkspacePolicy{
		GlobalRemoteState: perms.GlobalRemoteState.Bool,
		Permissions:       make([]authz.WorkspacePermission, len(perms.WorkspacePermissions)),
	}
	for i, perm := range perms.WorkspacePermissions {
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
