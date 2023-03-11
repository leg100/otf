package workspace

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/leg100/otf"
	"github.com/leg100/otf/sql"
	"github.com/leg100/otf/sql/pggen"
)

type (
	// pgdb is a state/state-version database on postgres
	pgdb struct {
		otf.DB // provides access to generated SQL queries
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
		FileTriggersEnabled        bool                   `json:"file_triggers_enabled"`
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
		LockUserID                 pgtype.Text            `json:"lock_user_id"`
		LatestRunID                pgtype.Text            `json:"latest_run_id"`
		OrganizationName           pgtype.Text            `json:"organization_name"`
		Branch                     pgtype.Text            `json:"branch"`
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
		Branch:                     r.Branch.String,
		CanQueueDestroyPlan:        r.CanQueueDestroyPlan,
		Description:                r.Description.String,
		Environment:                r.Environment.String,
		ExecutionMode:              ExecutionMode(r.ExecutionMode.String),
		FileTriggersEnabled:        r.FileTriggersEnabled,
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
		WorkingDirectory:           r.WorkingDirectory.String,
		Organization:               r.OrganizationName.String,
	}

	if r.WorkspaceConnection != nil {
		ws.Connection = &otf.Connection{
			VCSProviderID: r.WorkspaceConnection.VCSProviderID.String,
			Repo:          r.Webhook.Identifier.String,
		}
	}

	if r.LatestRunID.Status == pgtype.Present {
		ws.LatestRunID = &r.LatestRunID.String
	}

	if r.UserLock == nil && r.RunLock == nil {
		ws.LockedState = nil
	} else if r.UserLock != nil {
		ws.LockedState = UserLock{
			id: r.UserLock.UserID.String, username: r.UserLock.Username.String,
		}
	} else if r.RunLock != nil {
		ws.LockedState = RunLock{id: r.RunLock.RunID.String}
	} else {
		return nil, fmt.Errorf("workspace cannot be locked by both a run and a user")
	}

	return &ws, nil
}

func newdb(db otf.DB) *pgdb {
	return &pgdb{db}
}

func (db *pgdb) CreateWorkspace(ctx context.Context, ws *Workspace) error {
	_, err := db.InsertWorkspace(ctx, pggen.InsertWorkspaceParams{
		ID:                         sql.String(ws.ID),
		CreatedAt:                  sql.Timestamptz(ws.CreatedAt),
		UpdatedAt:                  sql.Timestamptz(ws.UpdatedAt),
		Name:                       sql.String(ws.Name),
		AllowDestroyPlan:           ws.AllowDestroyPlan,
		AutoApply:                  ws.AutoApply,
		Branch:                     sql.String(ws.Branch),
		CanQueueDestroyPlan:        ws.CanQueueDestroyPlan,
		Environment:                sql.String(ws.Environment),
		Description:                sql.String(ws.Description),
		ExecutionMode:              sql.String(string(ws.ExecutionMode)),
		FileTriggersEnabled:        ws.FileTriggersEnabled,
		GlobalRemoteState:          ws.GlobalRemoteState,
		MigrationEnvironment:       sql.String(ws.MigrationEnvironment),
		SourceName:                 sql.String(ws.SourceName),
		SourceURL:                  sql.String(ws.SourceURL),
		SpeculativeEnabled:         ws.SpeculativeEnabled,
		StructuredRunOutputEnabled: ws.StructuredRunOutputEnabled,
		TerraformVersion:           sql.String(ws.TerraformVersion),
		TriggerPrefixes:            ws.TriggerPrefixes,
		QueueAllRuns:               ws.QueueAllRuns,
		WorkingDirectory:           sql.String(ws.WorkingDirectory),
		OrganizationName:           sql.String(ws.Organization),
	})
	return sql.Error(err)
}

func (db *pgdb) UpdateWorkspace(ctx context.Context, workspaceID string, fn func(*Workspace) error) (*Workspace, error) {
	var ws *Workspace
	err := db.Tx(ctx, func(tx otf.DB) error {
		var err error
		// retrieve workspace
		result, err := tx.FindWorkspaceByIDForUpdate(ctx, sql.String(workspaceID))
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
		_, err = tx.UpdateWorkspaceByID(ctx, pggen.UpdateWorkspaceByIDParams{
			ID:                         sql.String(ws.ID),
			UpdatedAt:                  sql.Timestamptz(ws.UpdatedAt),
			AllowDestroyPlan:           ws.AllowDestroyPlan,
			AutoApply:                  ws.AutoApply,
			Branch:                     sql.String(ws.Branch),
			Description:                sql.String(ws.Description),
			ExecutionMode:              sql.String(string(ws.ExecutionMode)),
			Name:                       sql.String(ws.Name),
			QueueAllRuns:               ws.QueueAllRuns,
			SpeculativeEnabled:         ws.SpeculativeEnabled,
			StructuredRunOutputEnabled: ws.StructuredRunOutputEnabled,
			TerraformVersion:           sql.String(ws.TerraformVersion),
			TriggerPrefixes:            ws.TriggerPrefixes,
			WorkingDirectory:           sql.String(ws.WorkingDirectory),
		})
		return err
	})
	return ws, err
}

// SetCurrentRun sets the ID of the current run for the specified workspace.
func (db *pgdb) SetCurrentRun(ctx context.Context, workspaceID, runID string) (*Workspace, error) {
	_, err := db.UpdateWorkspaceLatestRun(ctx, sql.String(runID), sql.String(workspaceID))
	if err != nil {
		return nil, sql.Error(err)
	}
	return db.GetWorkspace(ctx, workspaceID)
}

func (db *pgdb) ListWorkspaces(ctx context.Context, opts WorkspaceListOptions) (*WorkspaceList, error) {
	batch := &pgx.Batch{}

	// Organization name filter is optional - if not provided use a % which in
	// SQL means match any organization.
	var organizationName string
	if opts.Organization != nil {
		organizationName = *opts.Organization
	} else {
		organizationName = "%"
	}

	db.FindWorkspacesBatch(batch, pggen.FindWorkspacesParams{
		OrganizationNames: []string{organizationName},
		Prefix:            sql.String(opts.Prefix),
		Limit:             opts.GetLimit(),
		Offset:            opts.GetOffset(),
	})
	db.CountWorkspacesBatch(batch, sql.String(opts.Prefix), []string{organizationName})
	results := db.SendBatch(ctx, batch)
	defer results.Close()

	rows, err := db.FindWorkspacesScan(results)
	if err != nil {
		return nil, err
	}
	count, err := db.CountWorkspacesScan(results)
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

	return &WorkspaceList{
		Items:      items,
		Pagination: otf.NewPagination(opts.ListOptions, *count),
	}, nil
}

func (db *pgdb) ListWorkspacesByWebhookID(ctx context.Context, id uuid.UUID) ([]*Workspace, error) {
	rows, err := db.FindWorkspacesByWebhookID(ctx, sql.UUID(id))
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

func (db *pgdb) ListWorkspacesByUserID(ctx context.Context, userID string, organization string, opts otf.ListOptions) (*WorkspaceList, error) {
	batch := &pgx.Batch{}

	db.FindWorkspacesByUserIDBatch(batch, pggen.FindWorkspacesByUserIDParams{
		OrganizationName: sql.String(organization),
		UserID:           sql.String(userID),
		Limit:            opts.GetLimit(),
		Offset:           opts.GetOffset(),
	})
	db.CountWorkspacesByUserIDBatch(batch, sql.String(organization), sql.String(userID))
	results := db.SendBatch(ctx, batch)
	defer results.Close()

	rows, err := db.FindWorkspacesByUserIDScan(results)
	if err != nil {
		return nil, err
	}
	count, err := db.CountWorkspacesByUserIDScan(results)
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

	return &WorkspaceList{
		Items:      items,
		Pagination: otf.NewPagination(opts, *count),
	}, nil
}

func (db *pgdb) GetWorkspaceIDByRunID(ctx context.Context, runID string) (string, error) {
	workspaceID, err := db.FindWorkspaceIDByRunID(ctx, sql.String(runID))
	if err != nil {
		return "", sql.Error(err)
	}
	return workspaceID.String, nil
}

func (db *pgdb) GetWorkspaceIDByStateVersionID(ctx context.Context, svID string) (string, error) {
	workspaceID, err := db.FindWorkspaceIDByStateVersionID(ctx, sql.String(svID))
	if err != nil {
		return "", sql.Error(err)
	}
	return workspaceID.String, nil
}

func (db *pgdb) GetWorkspaceIDByCVID(ctx context.Context, cvID string) (string, error) {
	workspaceID, err := db.FindWorkspaceIDByCVID(ctx, sql.String(cvID))
	if err != nil {
		return "", sql.Error(err)
	}
	return workspaceID.String, nil
}

func (db *pgdb) GetWorkspace(ctx context.Context, workspaceID string) (*Workspace, error) {
	result, err := db.FindWorkspaceByID(ctx, sql.String(workspaceID))
	if err != nil {
		return nil, sql.Error(err)
	}
	return pgresult(result).toWorkspace()
}

func (db *pgdb) GetWorkspaceByName(ctx context.Context, organization, workspace string) (*Workspace, error) {
	result, err := db.FindWorkspaceByName(ctx, sql.String(workspace), sql.String(organization))
	if err != nil {
		return nil, sql.Error(err)
	}
	return pgresult(result).toWorkspace()
}

func (db *pgdb) DeleteWorkspace(ctx context.Context, workspaceID string) error {
	_, err := db.DeleteWorkspaceByID(ctx, sql.String(workspaceID))
	if err != nil {
		return sql.Error(err)
	}
	return nil
}

func (db *pgdb) GetOrganizationNameByWorkspaceID(ctx context.Context, workspaceID string) (string, error) {
	name, err := db.FindOrganizationNameByWorkspaceID(ctx, sql.String(workspaceID))
	if err != nil {
		return "", sql.Error(err)
	}
	return name.String, nil
}

func (db *pgdb) getRun(ctx context.Context, runID string) (run, error) {
	result, err := db.FindRunByID(ctx, sql.String(runID))
	if err != nil {
		return run{}, sql.Error(err)
	}
	return run{
		ID:                     result.RunID.String,
		CreatedAt:              result.CreatedAt.Time.UTC(),
		IsDestroy:              result.IsDestroy,
		PositionInQueue:        result.PositionInQueue,
		Refresh:                result.Refresh,
		RefreshOnly:            result.RefreshOnly,
		Status:                 otf.RunStatus(result.Status.String),
		ReplaceAddrs:           result.ReplaceAddrs,
		TargetAddrs:            result.TargetAddrs,
		AutoApply:              result.AutoApply,
		Speculative:            result.Speculative,
		Latest:                 result.Latest,
		Organization:           result.OrganizationName.String,
		WorkspaceID:            result.WorkspaceID.String,
		ConfigurationVersionID: result.ConfigurationVersionID.String,
	}, nil
}

// tx constructs a new pgdb within a transaction.
func (db *pgdb) tx(ctx context.Context, callback func(*pgdb) error) error {
	return db.Tx(ctx, func(tx otf.DB) error {
		return callback(newdb(tx))
	})
}
