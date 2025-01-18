package workspace

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/authz"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/sql"
	"github.com/leg100/otf/internal/sql/sqlc"
)

type (
	// pgdb is a workspace database on postgres
	pgdb struct {
		*sql.DB // provides access to generated SQL queries
	}

	// pgresult represents the result of a database query for a workspace.
	pgresult struct {
		WorkspaceID                resource.ID
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
		LockRunID                  *resource.ID
		LatestRunID                *resource.ID
		OrganizationName           pgtype.Text
		Branch                     pgtype.Text
		CurrentStateVersionID      *resource.ID
		TriggerPatterns            []pgtype.Text
		VCSTagsRegex               pgtype.Text
		AllowCLIApply              pgtype.Bool
		AgentPoolID                *resource.ID
		LockUserID                 *resource.ID
		Tags                       []pgtype.Text
		LatestRunStatus            pgtype.Text
		VCSProviderID              resource.ID
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
		Organization:               r.OrganizationName.String,
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
			Status: runStatus(r.LatestRunStatus.String),
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
	q := db.Querier(ctx)
	params := sqlc.InsertWorkspaceParams{
		ID:                         ws.ID,
		CreatedAt:                  sql.Timestamptz(ws.CreatedAt),
		UpdatedAt:                  sql.Timestamptz(ws.UpdatedAt),
		AgentPoolID:                ws.AgentPoolID,
		AllowCLIApply:              sql.Bool(false),
		AllowDestroyPlan:           sql.Bool(ws.AllowDestroyPlan),
		AutoApply:                  sql.Bool(ws.AutoApply),
		Branch:                     sql.String(""),
		CanQueueDestroyPlan:        sql.Bool(ws.CanQueueDestroyPlan),
		Description:                sql.String(ws.Description),
		Environment:                sql.String(ws.Environment),
		ExecutionMode:              sql.String(string(ws.ExecutionMode)),
		GlobalRemoteState:          sql.Bool(ws.GlobalRemoteState),
		MigrationEnvironment:       sql.String(ws.MigrationEnvironment),
		Name:                       sql.String(ws.Name),
		QueueAllRuns:               sql.Bool(ws.QueueAllRuns),
		SpeculativeEnabled:         sql.Bool(ws.SpeculativeEnabled),
		SourceName:                 sql.String(ws.SourceName),
		SourceURL:                  sql.String(ws.SourceURL),
		StructuredRunOutputEnabled: sql.Bool(ws.StructuredRunOutputEnabled),
		TerraformVersion:           sql.String(ws.TerraformVersion),
		TriggerPrefixes:            sql.StringArray(ws.TriggerPrefixes),
		TriggerPatterns:            sql.StringArray(ws.TriggerPatterns),
		VCSTagsRegex:               sql.StringPtr(nil),
		WorkingDirectory:           sql.String(ws.WorkingDirectory),
		OrganizationName:           sql.String(ws.Organization),
	}
	if ws.Connection != nil {
		params.AllowCLIApply = sql.Bool(ws.Connection.AllowCLIApply)
		params.Branch = sql.String(ws.Connection.Branch)
		params.VCSTagsRegex = sql.String(ws.Connection.TagsRegex)
	}
	err := q.InsertWorkspace(ctx, params)
	return sql.Error(err)
}

func (db *pgdb) update(ctx context.Context, workspaceID resource.ID, fn func(context.Context, *Workspace) error) (*Workspace, error) {
	return sql.Updater(
		ctx,
		db.DB,
		func(ctx context.Context, q *sqlc.Queries) (*Workspace, error) {
			result, err := q.FindWorkspaceByIDForUpdate(ctx, workspaceID)
			if err != nil {
				return nil, err
			}
			return pgresult(result).toWorkspace()
		},
		fn,
		func(ctx context.Context, q *sqlc.Queries, ws *Workspace) error {
			params := sqlc.UpdateWorkspaceByIDParams{
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
			_, err := q.UpdateWorkspaceByID(ctx, params)
			return err
		},
	)
}

// setLatestRun sets the ID of the current run for the specified workspace.
func (db *pgdb) setLatestRun(ctx context.Context, workspaceID, runID resource.ID) (*Workspace, error) {
	q := db.Querier(ctx)

	err := q.UpdateWorkspaceLatestRun(ctx, sqlc.UpdateWorkspaceLatestRunParams{
		RunID:       &runID,
		WorkspaceID: workspaceID,
	})
	if err != nil {
		return nil, sql.Error(err)
	}

	return db.get(ctx, workspaceID)
}

func (db *pgdb) list(ctx context.Context, opts ListOptions) (*resource.Page[*Workspace], error) {
	q := db.Querier(ctx)

	// Organization name filter is optional - if not provided use a % which in
	// SQL means match any organization.
	organization := "%"
	if opts.Organization != nil {
		organization = *opts.Organization
	}
	tags := []string{}
	if len(opts.Tags) > 0 {
		tags = opts.Tags
	}

	rows, err := q.FindWorkspaces(ctx, sqlc.FindWorkspacesParams{
		OrganizationNames: sql.StringArray([]string{organization}),
		Search:            sql.String(opts.Search),
		Tags:              sql.StringArray(tags),
		Limit:             sql.GetLimit(opts.PageOptions),
		Offset:            sql.GetOffset(opts.PageOptions),
	})
	if err != nil {
		return nil, sql.Error(err)
	}
	count, err := q.CountWorkspaces(ctx, sqlc.CountWorkspacesParams{
		Search:            sql.String(opts.Search),
		OrganizationNames: sql.StringArray([]string{organization}),
		Tags:              sql.StringArray(tags),
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

func (db *pgdb) listByConnection(ctx context.Context, vcsProviderID resource.ID, repoPath string) ([]*Workspace, error) {
	q := db.Querier(ctx)

	rows, err := q.FindWorkspacesByConnection(ctx, sqlc.FindWorkspacesByConnectionParams{
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

func (db *pgdb) listByUsername(ctx context.Context, username string, organization string, opts resource.PageOptions) (*resource.Page[*Workspace], error) {
	q := db.Querier(ctx)

	rows, err := q.FindWorkspacesByUsername(ctx, sqlc.FindWorkspacesByUsernameParams{
		OrganizationName: sql.String(organization),
		Username:         sql.String(username),
		Limit:            sql.GetLimit(opts),
		Offset:           sql.GetOffset(opts),
	})
	if err != nil {
		return nil, err
	}
	count, err := q.CountWorkspacesByUsername(ctx, sqlc.CountWorkspacesByUsernameParams{
		OrganizationName: sql.String(organization),
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

func (db *pgdb) get(ctx context.Context, workspaceID resource.ID) (*Workspace, error) {
	q := db.Querier(ctx)
	result, err := q.FindWorkspaceByID(ctx, workspaceID)
	if err != nil {
		return nil, sql.Error(err)
	}
	return pgresult(result).toWorkspace()
}

func (db *pgdb) getByName(ctx context.Context, organization, workspace string) (*Workspace, error) {
	q := db.Querier(ctx)
	result, err := q.FindWorkspaceByName(ctx, sqlc.FindWorkspaceByNameParams{
		Name:             sql.String(workspace),
		OrganizationName: sql.String(organization),
	})
	if err != nil {
		return nil, sql.Error(err)
	}
	return pgresult(result).toWorkspace()
}

func (db *pgdb) delete(ctx context.Context, workspaceID resource.ID) error {
	q := db.Querier(ctx)
	err := q.DeleteWorkspaceByID(ctx, workspaceID)
	if err != nil {
		return sql.Error(err)
	}
	return nil
}

func (db *pgdb) SetWorkspacePermission(ctx context.Context, workspaceID, teamID resource.ID, role authz.Role) error {
	err := db.Querier(ctx).UpsertWorkspacePermission(ctx, sqlc.UpsertWorkspacePermissionParams{
		WorkspaceID: workspaceID,
		TeamID:      teamID,
		Role:        sql.String(role.String()),
	})
	if err != nil {
		return sql.Error(err)
	}
	return nil
}

func (db *pgdb) UnsetWorkspacePermission(ctx context.Context, workspaceID, teamID resource.ID) error {
	err := db.Querier(ctx).DeleteWorkspacePermissionByID(ctx, sqlc.DeleteWorkspacePermissionByIDParams{
		WorkspaceID: workspaceID,
		TeamID:      teamID,
	})
	if err != nil {
		return sql.Error(err)
	}
	return nil
}

func (db *pgdb) GetWorkspacePolicy(ctx context.Context, workspaceID resource.ID) (authz.WorkspacePolicy, error) {
	perms, err := db.Querier(ctx).FindWorkspacePermissionsAndGlobalRemoteState(ctx, workspaceID)
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
