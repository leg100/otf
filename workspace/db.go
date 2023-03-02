package workspace

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/leg100/otf"
	"github.com/leg100/otf/sql"
	"github.com/leg100/otf/sql/pggen"
)

// pgdb is a state/state-version database on postgres
type pgdb struct {
	otf.DB // provides access to generated SQL queries
}

// WorkspaceResult represents the result of a database query for a workspace.
type WorkspaceResult struct {
	WorkspaceID                pgtype.Text           `json:"workspace_id"`
	CreatedAt                  pgtype.Timestamptz    `json:"created_at"`
	UpdatedAt                  pgtype.Timestamptz    `json:"updated_at"`
	AllowDestroyPlan           bool                  `json:"allow_destroy_plan"`
	AutoApply                  bool                  `json:"auto_apply"`
	CanQueueDestroyPlan        bool                  `json:"can_queue_destroy_plan"`
	Description                pgtype.Text           `json:"description"`
	Environment                pgtype.Text           `json:"environment"`
	ExecutionMode              pgtype.Text           `json:"execution_mode"`
	FileTriggersEnabled        bool                  `json:"file_triggers_enabled"`
	GlobalRemoteState          bool                  `json:"global_remote_state"`
	MigrationEnvironment       pgtype.Text           `json:"migration_environment"`
	Name                       pgtype.Text           `json:"name"`
	QueueAllRuns               bool                  `json:"queue_all_runs"`
	SpeculativeEnabled         bool                  `json:"speculative_enabled"`
	SourceName                 pgtype.Text           `json:"source_name"`
	SourceURL                  pgtype.Text           `json:"source_url"`
	StructuredRunOutputEnabled bool                  `json:"structured_run_output_enabled"`
	TerraformVersion           pgtype.Text           `json:"terraform_version"`
	TriggerPrefixes            []string              `json:"trigger_prefixes"`
	WorkingDirectory           pgtype.Text           `json:"working_directory"`
	LockRunID                  pgtype.Text           `json:"lock_run_id"`
	LockUserID                 pgtype.Text           `json:"lock_user_id"`
	LatestRunID                pgtype.Text           `json:"latest_run_id"`
	OrganizationName           pgtype.Text           `json:"organization_name"`
	UserLock                   *pggen.Users          `json:"user_lock"`
	RunLock                    *pggen.Runs           `json:"run_lock"`
	WorkspaceRepo              *pggen.WorkspaceRepos `json:"workspace_repo"`
	Webhook                    *pggen.Webhooks       `json:"webhook"`
}

func UnmarshalWorkspaceResult(result WorkspaceResult) (*Workspace, error) {
	ws := Workspace{
		id:                         result.WorkspaceID.String,
		createdAt:                  result.CreatedAt.Time.UTC(),
		updatedAt:                  result.UpdatedAt.Time.UTC(),
		allowDestroyPlan:           result.AllowDestroyPlan,
		autoApply:                  result.AutoApply,
		canQueueDestroyPlan:        result.CanQueueDestroyPlan,
		description:                result.Description.String,
		environment:                result.Environment.String,
		executionMode:              otf.ExecutionMode(result.ExecutionMode.String),
		fileTriggersEnabled:        result.FileTriggersEnabled,
		globalRemoteState:          result.GlobalRemoteState,
		migrationEnvironment:       result.MigrationEnvironment.String,
		name:                       result.Name.String,
		queueAllRuns:               result.QueueAllRuns,
		speculativeEnabled:         result.SpeculativeEnabled,
		structuredRunOutputEnabled: result.StructuredRunOutputEnabled,
		sourceName:                 result.SourceName.String,
		sourceURL:                  result.SourceURL.String,
		terraformVersion:           result.TerraformVersion.String,
		triggerPrefixes:            result.TriggerPrefixes,
		workingDirectory:           result.WorkingDirectory.String,
		organization:               result.OrganizationName.String,
	}

	if result.WorkspaceRepo != nil {
		ws.repo = &WorkspaceRepo{
			Branch:     result.WorkspaceRepo.Branch.String,
			ProviderID: result.WorkspaceRepo.VCSProviderID.String,
			WebhookID:  result.Webhook.WebhookID.Bytes,
			Identifier: result.Webhook.Identifier.String,
		}
	}

	if result.LatestRunID.Status == pgtype.Present {
		ws.latestRunID = &result.LatestRunID.String
	}

	if err := unmarshalWorkspaceLock(&ws, &result); err != nil {
		return nil, err
	}

	return &ws, nil
}

func newdb(db otf.DB) *pgdb {
	return &pgdb{db}
}

func (db *pgdb) CreateWorkspace(ctx context.Context, ws *Workspace) error {
	err := db.Transaction(ctx, func(tx otf.Database) error {
		_, err := tx.InsertWorkspace(ctx, pggen.InsertWorkspaceParams{
			ID:                         sql.String(ws.ID()),
			CreatedAt:                  sql.Timestamptz(ws.CreatedAt()),
			UpdatedAt:                  sql.Timestamptz(ws.UpdatedAt()),
			Name:                       sql.String(ws.Name()),
			AllowDestroyPlan:           ws.AllowDestroyPlan(),
			AutoApply:                  ws.AutoApply(),
			CanQueueDestroyPlan:        ws.CanQueueDestroyPlan(),
			Environment:                sql.String(ws.Environment()),
			Description:                sql.String(ws.Description()),
			ExecutionMode:              sql.String(string(ws.ExecutionMode())),
			FileTriggersEnabled:        ws.FileTriggersEnabled(),
			GlobalRemoteState:          ws.GlobalRemoteState(),
			MigrationEnvironment:       sql.String(ws.MigrationEnvironment()),
			SourceName:                 sql.String(ws.SourceName()),
			SourceURL:                  sql.String(ws.SourceURL()),
			SpeculativeEnabled:         ws.SpeculativeEnabled(),
			StructuredRunOutputEnabled: ws.StructuredRunOutputEnabled(),
			TerraformVersion:           sql.String(ws.TerraformVersion()),
			TriggerPrefixes:            ws.TriggerPrefixes(),
			QueueAllRuns:               ws.QueueAllRuns(),
			WorkingDirectory:           sql.String(ws.WorkingDirectory()),
			OrganizationName:           sql.String(ws.Organization()),
		})
		if err != nil {
			return sql.Error(err)
		}
		if ws.Repo() != nil {
			_, err = tx.InsertWorkspaceRepo(ctx, pggen.InsertWorkspaceRepoParams{
				Branch:        sql.String(ws.Repo().Branch),
				WebhookID:     sql.UUID(ws.Repo().WebhookID),
				VCSProviderID: sql.String(ws.Repo().ProviderID),
				WorkspaceID:   sql.String(ws.ID()),
			})
			if err != nil {
				return sql.Error(err)
			}
		}
		return nil
	})
	if err != nil {
		return sql.Error(err)
	}
	return nil
}

func (db *pgdb) UpdateWorkspace(ctx context.Context, workspaceID string, fn func(*Workspace) error) (*Workspace, error) {
	var ws *Workspace
	err := db.Transaction(ctx, func(tx otf.Database) error {
		var err error
		// retrieve workspace
		result, err := tx.FindWorkspaceByIDForUpdate(ctx, sql.String(workspaceID))
		if err != nil {
			return sql.Error(err)
		}
		ws, err = UnmarshalWorkspaceResult(WorkspaceResult(result))
		if err != nil {
			return err
		}
		// update workspace
		if err := fn(ws); err != nil {
			return err
		}
		// persist update
		_, err = tx.UpdateWorkspaceByID(ctx, pggen.UpdateWorkspaceByIDParams{
			ID:                         sql.String(ws.ID()),
			UpdatedAt:                  sql.Timestamptz(ws.UpdatedAt()),
			AllowDestroyPlan:           ws.AllowDestroyPlan(),
			AutoApply:                  ws.AutoApply(),
			Description:                sql.String(ws.Description()),
			ExecutionMode:              sql.String(string(ws.ExecutionMode())),
			Name:                       sql.String(ws.Name()),
			QueueAllRuns:               ws.QueueAllRuns(),
			SpeculativeEnabled:         ws.SpeculativeEnabled(),
			StructuredRunOutputEnabled: ws.StructuredRunOutputEnabled(),
			TerraformVersion:           sql.String(ws.TerraformVersion()),
			TriggerPrefixes:            ws.TriggerPrefixes(),
			WorkingDirectory:           sql.String(ws.WorkingDirectory()),
		})
		return err
	})
	return ws, err
}

func (db *pgdb) CreateWorkspaceRepo(ctx context.Context, workspaceID string, repo WorkspaceRepo) (*Workspace, error) {
	_, err := db.InsertWorkspaceRepo(ctx, pggen.InsertWorkspaceRepoParams{
		Branch:        sql.String(repo.Branch),
		WebhookID:     sql.UUID(repo.WebhookID),
		VCSProviderID: sql.String(repo.ProviderID),
		WorkspaceID:   sql.String(workspaceID),
	})
	if err != nil {
		return nil, sql.Error(err)
	}
	ws, err := db.GetWorkspace(ctx, workspaceID)
	return ws, sql.Error(err)
}

func (db *pgdb) UpdateWorkspaceRepo(ctx context.Context, workspaceID string, repo WorkspaceRepo) (*Workspace, error) {
	_, err := db.UpdateWorkspaceRepoByID(ctx, sql.String(repo.Branch), sql.String(workspaceID))
	if err != nil {
		return nil, sql.Error(err)
	}
	ws, err := db.GetWorkspace(ctx, workspaceID)
	return ws, sql.Error(err)
}

func (db *pgdb) DeleteWorkspaceRepo(ctx context.Context, workspaceID string) (*Workspace, error) {
	_, err := db.DeleteWorkspaceRepoByID(ctx, sql.String(workspaceID))
	if err != nil {
		return nil, sql.Error(err)
	}
	ws, err := db.GetWorkspace(ctx, workspaceID)
	return ws, sql.Error(err)
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
		ws, err := UnmarshalWorkspaceResult(WorkspaceResult(r))
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
		ws, err := UnmarshalWorkspaceResult(WorkspaceResult(r))
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
		ws, err := UnmarshalWorkspaceResult(WorkspaceResult(r))
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
	return UnmarshalWorkspaceResult(WorkspaceResult(result))
}

func (db *pgdb) GetWorkspaceByName(ctx context.Context, organization, workspace string) (*Workspace, error) {
	result, err := db.FindWorkspaceByName(ctx, sql.String(workspace), sql.String(organization))
	if err != nil {
		return nil, sql.Error(err)
	}
	return UnmarshalWorkspaceResult(WorkspaceResult(result))
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

// tx constructs a new pgdb within a transaction.
func (db *pgdb) tx(ctx context.Context, callback func(*pgdb) error) error {
	return db.Tx(ctx, func(tx otf.DB) error {
		return callback(newdb(tx))
	})
}
