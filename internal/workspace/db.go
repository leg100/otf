package workspace

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/sql"
	"github.com/leg100/otf/internal/sql/pggen"
)

type (
	// pgdb is a workspace database on postgres
	pgdb struct {
		*sql.DB // provides access to generated SQL queries
	}

	// pgresult represents the result of a database query for a workspace.
	pgresult struct {
		WorkspaceID                pgtype.Text            `json:"workspace_id"`
		CreatedAt                  pgtype.Timestamptz     `json:"created_at"`
		UpdatedAt                  pgtype.Timestamptz     `json:"updated_at"`
		AllowDestroyPlan           bool                   `json:"allow_destroy_plan"`
		AutoApply                  bool                   `json:"auto_apply"`
		CanQueueDestroyPlan        bool                   `json:"can_queue_destroy_plan"`
		Description                pgtype.Text            `json:"description"`
		Environment                pgtype.Text            `json:"environment"`
		ExecutionMode              pgtype.Text            `json:"execution_mode"`
		GlobalRemoteState          bool                   `json:"global_remote_state"`
		MigrationEnvironment       pgtype.Text            `json:"migration_environment"`
		Name                       pgtype.Text            `json:"name"`
		QueueAllRuns               bool                   `json:"queue_all_runs"`
		SpeculativeEnabled         bool                   `json:"speculative_enabled"`
		SourceName                 pgtype.Text            `json:"source_name"`
		SourceURL                  pgtype.Text            `json:"source_url"`
		StructuredRunOutputEnabled bool                   `json:"structured_run_output_enabled"`
		TerraformVersion           pgtype.Text            `json:"terraform_version"`
		TriggerPrefixes            []string               `json:"trigger_prefixes"`
		WorkingDirectory           pgtype.Text            `json:"working_directory"`
		LockRunID                  pgtype.Text            `json:"lock_run_id"`
		LatestRunID                pgtype.Text            `json:"latest_run_id"`
		OrganizationName           pgtype.Text            `json:"organization_name"`
		Branch                     pgtype.Text            `json:"branch"`
		LockUsername               pgtype.Text            `json:"lock_username"`
		CurrentStateVersionID      pgtype.Text            `json:"current_state_version_id"`
		TriggerPatterns            []string               `json:"trigger_patterns"`
		VCSTagsRegex               pgtype.Text            `json:"vcs_tags_regex"`
		LatestRunStatus            pgtype.Text            `json:"latest_run_status"`
		Tags                       []string               `json:"tags"`
		UserLock                   *pggen.Users           `json:"user_lock"`
		RunLock                    *pggen.Runs            `json:"run_lock"`
		WorkspaceConnection        *pggen.RepoConnections `json:"workspace_connection"`
		Webhook                    *pggen.Webhooks        `json:"webhook"`
	}
)

func (r pgresult) toWorkspace() (*Workspace, error) {
	ws := Workspace{
		ID:                         r.WorkspaceID.String,
		CreatedAt:                  r.CreatedAt.Time.UTC(),
		UpdatedAt:                  r.UpdatedAt.Time.UTC(),
		AllowDestroyPlan:           r.AllowDestroyPlan,
		AutoApply:                  r.AutoApply,
		CanQueueDestroyPlan:        r.CanQueueDestroyPlan,
		Description:                r.Description.String,
		Environment:                r.Environment.String,
		ExecutionMode:              ExecutionMode(r.ExecutionMode.String),
		GlobalRemoteState:          r.GlobalRemoteState,
		MigrationEnvironment:       r.MigrationEnvironment.String,
		Name:                       r.Name.String,
		QueueAllRuns:               r.QueueAllRuns,
		SpeculativeEnabled:         r.SpeculativeEnabled,
		StructuredRunOutputEnabled: r.StructuredRunOutputEnabled,
		SourceName:                 r.SourceName.String,
		SourceURL:                  r.SourceURL.String,
		TerraformVersion:           r.TerraformVersion.String,
		TriggerPrefixes:            r.TriggerPrefixes,
		TriggerPatterns:            r.TriggerPatterns,
		WorkingDirectory:           r.WorkingDirectory.String,
		Organization:               r.OrganizationName.String,
		Tags:                       r.Tags,
	}

	if r.WorkspaceConnection != nil {
		ws.Connection = &Connection{
			VCSProviderID: r.Webhook.VCSProviderID.String,
			Repo:          r.Webhook.Identifier.String,
			Branch:        r.Branch.String,
		}
		if r.VCSTagsRegex.Status == pgtype.Present {
			ws.Connection.TagsRegex = r.VCSTagsRegex.String
		}
	}

	if r.LatestRunID.Status == pgtype.Present && r.LatestRunStatus.Status == pgtype.Present {
		ws.LatestRun = &LatestRun{
			ID:     r.LatestRunID.String,
			Status: internal.RunStatus(r.LatestRunStatus.String),
		}
	}

	if r.UserLock != nil {
		ws.Lock = &Lock{
			id:       r.UserLock.Username.String,
			LockKind: UserLock,
		}
	} else if r.RunLock != nil {
		ws.Lock = &Lock{
			id:       r.RunLock.RunID.String,
			LockKind: RunLock,
		}
	}

	return &ws, nil
}

func (db *pgdb) create(ctx context.Context, ws *Workspace) error {
	q := db.Conn(ctx)
	params := pggen.InsertWorkspaceParams{
		ID:                         sql.String(ws.ID),
		CreatedAt:                  sql.Timestamptz(ws.CreatedAt),
		UpdatedAt:                  sql.Timestamptz(ws.UpdatedAt),
		Name:                       sql.String(ws.Name),
		AllowDestroyPlan:           ws.AllowDestroyPlan,
		AutoApply:                  ws.AutoApply,
		CanQueueDestroyPlan:        ws.CanQueueDestroyPlan,
		Environment:                sql.String(ws.Environment),
		Description:                sql.String(ws.Description),
		ExecutionMode:              sql.String(string(ws.ExecutionMode)),
		GlobalRemoteState:          ws.GlobalRemoteState,
		MigrationEnvironment:       sql.String(ws.MigrationEnvironment),
		SourceName:                 sql.String(ws.SourceName),
		SourceURL:                  sql.String(ws.SourceURL),
		SpeculativeEnabled:         ws.SpeculativeEnabled,
		StructuredRunOutputEnabled: ws.StructuredRunOutputEnabled,
		TerraformVersion:           sql.String(ws.TerraformVersion),
		TriggerPrefixes:            ws.TriggerPrefixes,
		TriggerPatterns:            ws.TriggerPatterns,
		QueueAllRuns:               ws.QueueAllRuns,
		WorkingDirectory:           sql.String(ws.WorkingDirectory),
		OrganizationName:           sql.String(ws.Organization),
		Branch:                     sql.String(""),
		VCSTagsRegex:               sql.StringPtr(nil),
	}
	if ws.Connection != nil {
		params.Branch = sql.String(ws.Connection.Branch)
		params.VCSTagsRegex = sql.String(ws.Connection.TagsRegex)
	}
	_, err := q.InsertWorkspace(ctx, params)
	return sql.Error(err)
}

func (db *pgdb) update(ctx context.Context, workspaceID string, fn func(*Workspace) error) (*Workspace, error) {
	var ws *Workspace
	err := db.Tx(ctx, func(ctx context.Context, q pggen.Querier) error {
		var err error
		// retrieve workspace
		result, err := q.FindWorkspaceByIDForUpdate(ctx, sql.String(workspaceID))
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
		params := pggen.UpdateWorkspaceByIDParams{
			ID:                         sql.String(ws.ID),
			UpdatedAt:                  sql.Timestamptz(ws.UpdatedAt),
			AllowDestroyPlan:           ws.AllowDestroyPlan,
			AutoApply:                  ws.AutoApply,
			Description:                sql.String(ws.Description),
			ExecutionMode:              sql.String(string(ws.ExecutionMode)),
			Name:                       sql.String(ws.Name),
			QueueAllRuns:               ws.QueueAllRuns,
			SpeculativeEnabled:         ws.SpeculativeEnabled,
			StructuredRunOutputEnabled: ws.StructuredRunOutputEnabled,
			TerraformVersion:           sql.String(ws.TerraformVersion),
			TriggerPrefixes:            ws.TriggerPrefixes,
			TriggerPatterns:            ws.TriggerPatterns,
			WorkingDirectory:           sql.String(ws.WorkingDirectory),
			Branch:                     sql.String(""),
			VCSTagsRegex:               sql.StringPtr(nil),
		}
		if ws.Connection != nil {
			params.Branch = sql.String(ws.Connection.Branch)
			params.VCSTagsRegex = sql.String(ws.Connection.TagsRegex)
		}
		_, err = q.UpdateWorkspaceByID(ctx, params)
		return err
	})
	return ws, err
}

// setCurrentRun sets the ID of the current run for the specified workspace.
func (db *pgdb) setCurrentRun(ctx context.Context, workspaceID, runID string) (*Workspace, error) {
	q := db.Conn(ctx)
	_, err := q.UpdateWorkspaceLatestRun(ctx, sql.String(runID), sql.String(workspaceID))
	if err != nil {
		return nil, sql.Error(err)
	}
	return db.get(ctx, workspaceID)
}

func (db *pgdb) list(ctx context.Context, opts ListOptions) (*resource.Page[*Workspace], error) {
	q := db.Conn(ctx)
	batch := &pgx.Batch{}

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

	q.FindWorkspacesBatch(batch, pggen.FindWorkspacesParams{
		OrganizationNames: []string{organization},
		Search:            sql.String(opts.Search),
		Tags:              tags,
		Limit:             opts.GetLimit(),
		Offset:            opts.GetOffset(),
	})
	q.CountWorkspacesBatch(batch, pggen.CountWorkspacesParams{
		Search:            sql.String(opts.Search),
		OrganizationNames: []string{organization},
		Tags:              tags,
	})
	results := db.SendBatch(ctx, batch)
	defer results.Close()

	rows, err := q.FindWorkspacesScan(results)
	if err != nil {
		return nil, err
	}
	count, err := q.CountWorkspacesScan(results)
	if err != nil {
		return nil, err
	}

	var items []*Workspace
	for _, r := range rows {
		ws, err := pgresult(r).toWorkspace()
		if err != nil {
			return nil, err
		}
		items = append(items, ws)
	}

	return resource.NewPage(items, opts.PageOptions, internal.Int64(count.Int)), nil
}

func (db *pgdb) listByWebhookID(ctx context.Context, id uuid.UUID) ([]*Workspace, error) {
	q := db.Conn(ctx)
	rows, err := q.FindWorkspacesByWebhookID(ctx, sql.UUID(id))
	if err != nil {
		return nil, err
	}

	var items []*Workspace
	for _, r := range rows {
		ws, err := pgresult(r).toWorkspace()
		if err != nil {
			return nil, err
		}
		items = append(items, ws)
	}

	return items, nil
}

func (db *pgdb) listByUsername(ctx context.Context, username string, organization string, opts resource.PageOptions) (*resource.Page[*Workspace], error) {
	q := db.Conn(ctx)
	batch := &pgx.Batch{}

	q.FindWorkspacesByUsernameBatch(batch, pggen.FindWorkspacesByUsernameParams{
		OrganizationName: sql.String(organization),
		Username:         sql.String(username),
		Limit:            opts.GetLimit(),
		Offset:           opts.GetOffset(),
	})
	q.CountWorkspacesByUsernameBatch(batch, sql.String(organization), sql.String(username))
	results := db.SendBatch(ctx, batch)
	defer results.Close()

	rows, err := q.FindWorkspacesByUsernameScan(results)
	if err != nil {
		return nil, err
	}
	count, err := q.CountWorkspacesByUsernameScan(results)
	if err != nil {
		return nil, err
	}

	var items []*Workspace
	for _, r := range rows {
		ws, err := pgresult(r).toWorkspace()
		if err != nil {
			return nil, err
		}
		items = append(items, ws)
	}

	return resource.NewPage(items, opts, internal.Int64(count.Int)), nil
}

func (db *pgdb) get(ctx context.Context, workspaceID string) (*Workspace, error) {
	q := db.Conn(ctx)
	result, err := q.FindWorkspaceByID(ctx, sql.String(workspaceID))
	if err != nil {
		return nil, sql.Error(err)
	}
	return pgresult(result).toWorkspace()
}

func (db *pgdb) getByName(ctx context.Context, organization, workspace string) (*Workspace, error) {
	q := db.Conn(ctx)
	result, err := q.FindWorkspaceByName(ctx, sql.String(workspace), sql.String(organization))
	if err != nil {
		return nil, sql.Error(err)
	}
	return pgresult(result).toWorkspace()
}

func (db *pgdb) delete(ctx context.Context, workspaceID string) error {
	q := db.Conn(ctx)
	_, err := q.DeleteWorkspaceByID(ctx, sql.String(workspaceID))
	if err != nil {
		return sql.Error(err)
	}
	return nil
}
