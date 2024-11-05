package workspace

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/leg100/otf/internal"
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
		WorkspaceID                pgtype.Text
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
		LockRunID                  pgtype.Text
		LatestRunID                pgtype.Text
		OrganizationName           pgtype.Text
		Branch                     pgtype.Text
		LockUsername               pgtype.Text
		CurrentStateVersionID      pgtype.Text
		TriggerPatterns            []pgtype.Text
		VCSTagsRegex               pgtype.Text
		AllowCLIApply              pgtype.Bool
		AgentPoolID                pgtype.Text
		Tags                       []pgtype.Text
		LatestRunStatus            pgtype.Text
		VCSProviderID              pgtype.Text
		RepoPath                   pgtype.Text
	}
)

func (r pgresult) toWorkspace() (*Workspace, error) {
	ws := Workspace{
		ID:                         resource.ID{Kind: WorkspaceKind, ID: r.WorkspaceID.String},
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
	}
	if r.AgentPoolID.Valid {
		agentPoolIDValue := resource.ParseID(r.AgentPoolID.String)
		ws.AgentPoolID = &agentPoolIDValue
	}

	if r.VCSProviderID.Valid && r.RepoPath.Valid {
		ws.Connection = &Connection{
			AllowCLIApply: r.AllowCLIApply.Bool,
			VCSProviderID: resource.ParseID(r.VCSProviderID.String),
			Repo:          r.RepoPath.String,
			Branch:        r.Branch.String,
		}
		if r.VCSTagsRegex.Valid {
			ws.Connection.TagsRegex = r.VCSTagsRegex.String
		}
	}

	if r.LatestRunID.Valid && r.LatestRunStatus.Valid {
		ws.LatestRun = &LatestRun{
			ID:     resource.ParseID(r.LatestRunID.String),
			Status: runStatus(r.LatestRunStatus.String),
		}
	}

	if r.LockUsername.Valid {
		ws.Lock = &Lock{
			id:       r.LockUsername.String,
			LockKind: UserLock,
		}
	} else if r.LockRunID.Valid {
		ws.Lock = &Lock{
			id:       r.LockRunID.String,
			LockKind: RunLock,
		}
	}

	return &ws, nil
}

func (db *pgdb) create(ctx context.Context, ws *Workspace) error {
	q := db.Querier(ctx)
	params := sqlc.InsertWorkspaceParams{
		ID:                         sql.ID(ws.ID),
		CreatedAt:                  sql.Timestamptz(ws.CreatedAt),
		UpdatedAt:                  sql.Timestamptz(ws.UpdatedAt),
		AgentPoolID:                sql.IDPtr(ws.AgentPoolID),
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

func (db *pgdb) update(ctx context.Context, workspaceID resource.ID, fn func(*Workspace) error) (*Workspace, error) {
	var ws *Workspace
	err := db.Tx(ctx, func(ctx context.Context, q *sqlc.Queries) error {
		var err error
		// retrieve workspace
		result, err := q.FindWorkspaceByIDForUpdate(ctx, sql.ID(workspaceID))
		if err != nil {
			return sql.Error(err)
		}
		ws, err = pgresult(result).toWorkspace()
		if err != nil {
			return err
		}
		// update workspace
		if err := fn(ws); err != nil {
			return err
		}
		// persist update
		params := sqlc.UpdateWorkspaceByIDParams{
			AgentPoolID:                sql.IDPtr(ws.AgentPoolID),
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
			ID:                         sql.ID(ws.ID),
		}
		if ws.Connection != nil {
			params.AllowCLIApply = sql.Bool(ws.Connection.AllowCLIApply)
			params.Branch = sql.String(ws.Connection.Branch)
			params.VCSTagsRegex = sql.String(ws.Connection.TagsRegex)
		}
		_, err = q.UpdateWorkspaceByID(ctx, params)
		return err
	})
	return ws, err
}

// setCurrentRun sets the ID of the current run for the specified workspace.
func (db *pgdb) setCurrentRun(ctx context.Context, workspaceID, runID resource.ID) (*Workspace, error) {
	q := db.Querier(ctx)

	err := q.UpdateWorkspaceLatestRun(ctx, sqlc.UpdateWorkspaceLatestRunParams{
		RunID:       sql.ID(runID),
		WorkspaceID: sql.ID(workspaceID),
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
		VCSProviderID: sql.ID(vcsProviderID),
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
	result, err := q.FindWorkspaceByID(ctx, sql.ID(workspaceID))
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
	err := q.DeleteWorkspaceByID(ctx, sql.ID(workspaceID))
	if err != nil {
		return sql.Error(err)
	}
	return nil
}
