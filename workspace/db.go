package workspace

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/leg100/otf"
	"github.com/leg100/otf/sql"
	"github.com/leg100/otf/sql/pggen"
)

// pgdb is a state/state-version database on postgres
type DB struct {
	otf.Database // provides access to generated SQL queries
}

func NewDB(db otf.Database) *DB {
	return newPGDB(db)
}

func newPGDB(db otf.Database) *DB {
	return &DB{db}
}

func (db *DB) CreateWorkspace(ctx context.Context, ws *Workspace) error {
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

func (db *DB) UpdateWorkspace(ctx context.Context, workspaceID string, fn func(*Workspace) error) (*Workspace, error) {
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

func (db *DB) CreateWorkspaceRepo(ctx context.Context, workspaceID string, repo WorkspaceRepo) (*Workspace, error) {
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

func CreateWorkspaceRepo(ctx context.Context, db otf.Database, workspaceID string, repo WorkspaceRepo) error {
	_, err := db.InsertWorkspaceRepo(ctx, pggen.InsertWorkspaceRepoParams{
		Branch:        sql.String(repo.Branch),
		WebhookID:     sql.UUID(repo.WebhookID),
		VCSProviderID: sql.String(repo.ProviderID),
		WorkspaceID:   sql.String(workspaceID),
	})
	return sql.Error(err)
}

func (db *DB) UpdateWorkspaceRepo(ctx context.Context, workspaceID string, repo WorkspaceRepo) (*Workspace, error) {
	_, err := db.UpdateWorkspaceRepoByID(ctx, sql.String(repo.Branch), sql.String(workspaceID))
	if err != nil {
		return nil, sql.Error(err)
	}
	ws, err := db.GetWorkspace(ctx, workspaceID)
	return ws, sql.Error(err)
}

func (db *DB) DeleteWorkspaceRepo(ctx context.Context, workspaceID string) (*otf.Workspace, error) {
	_, err := db.Querier.DeleteWorkspaceRepoByID(ctx, sql.String(workspaceID))
	if err != nil {
		return nil, sql.Error(err)
	}
	ws, err := db.GetWorkspace(ctx, workspaceID)
	return ws, sql.Error(err)
}

func DeleteWorkspaceRepo(ctx context.Context, db otf.Database, workspaceID string) error {
	_, err := db.DeleteWorkspaceRepoByID(ctx, sql.String(workspaceID))
	return sql.Error(err)
}

// LockWorkspace locks the specified workspace.
func (db *DB) LockWorkspace(ctx context.Context, workspaceID string, opts otf.WorkspaceLockOptions) (*otf.Workspace, error) {
	subj, err := otf.LockFromContext(ctx)
	if err != nil {
		return nil, err
	}

	var ws *otf.Workspace
	err = db.tx(ctx, func(tx *DB) error {
		// retrieve workspace
		result, err := tx.FindWorkspaceByIDForUpdate(ctx, sql.String(workspaceID))
		if err != nil {
			return err
		}
		ws, err = otf.UnmarshalWorkspaceResult(otf.WorkspaceResult(result))
		if err != nil {
			return err
		}
		// lock the workspace
		if err := ws.Lock(subj); err != nil {
			return err
		}
		// persist to db
		params, err := otf.MarshalWorkspaceLockParams(ws)
		if err != nil {
			return err
		}
		_, err = tx.UpdateWorkspaceLockByID(ctx, params)
		if err != nil {
			return sql.Error(err)
		}
		return nil
	})
	// return ws with new lock
	return ws, err
}

// SetCurrentRun sets the ID of the current run for the specified workspace.
func (db *DB) SetCurrentRun(ctx context.Context, workspaceID, runID string) (*otf.Workspace, error) {
	_, err := db.UpdateWorkspaceLatestRun(ctx, sql.String(runID), sql.String(workspaceID))
	if err != nil {
		return nil, sql.Error(err)
	}
	return db.GetWorkspace(ctx, workspaceID)
}

// UnlockWorkspace unlocks the specified workspace; the caller has the
// opportunity to check the current locker passed into the provided callback. If
// an error is returned the unlock will not go ahead.
func (db *DB) UnlockWorkspace(ctx context.Context, workspaceID string, opts otf.WorkspaceUnlockOptions) (*otf.Workspace, error) {
	subj, err := otf.LockFromContext(ctx)
	if err != nil {
		return nil, err
	}

	var ws *otf.Workspace
	err = db.tx(ctx, func(tx *DB) error {
		// retrieve workspace
		result, err := tx.FindWorkspaceByIDForUpdate(ctx, sql.String(workspaceID))
		if err != nil {
			return err
		}
		ws, err = otf.UnmarshalWorkspaceResult(otf.WorkspaceResult(result))
		if err != nil {
			return err
		}
		// unlock workspace
		if err := ws.Unlock(subj, opts.Force); err != nil {
			return err
		}
		// persist to db
		params, err := otf.MarshalWorkspaceLockParams(ws)
		if err != nil {
			return err
		}
		_, err = tx.UpdateWorkspaceLockByID(ctx, params)
		if err != nil {
			return sql.Error(err)
		}
		return nil
	})
	// return ws with new lock
	return ws, err
}

func (db *DB) ListWorkspaces(ctx context.Context, opts otf.WorkspaceListOptions) (*otf.WorkspaceList, error) {
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
		Pagination: NewPagination(opts.ListOptions, *count),
	}, nil
}

func (db *DB) ListWorkspacesByWebhookID(ctx context.Context, id uuid.UUID) ([]*Workspace, error) {
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

func (db *DB) ListWorkspacesByUserID(ctx context.Context, userID string, organization string, opts otf.ListOptions) (*otf.WorkspaceList, error) {
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

func (db *DB) GetWorkspaceIDByRunID(ctx context.Context, runID string) (string, error) {
	workspaceID, err := db.FindWorkspaceIDByRunID(ctx, sql.String(runID))
	if err != nil {
		return "", sql.Error(err)
	}
	return workspaceID.String, nil
}

func (db *DB) GetWorkspaceIDByStateVersionID(ctx context.Context, svID string) (string, error) {
	workspaceID, err := db.FindWorkspaceIDByStateVersionID(ctx, sql.String(svID))
	if err != nil {
		return "", sql.Error(err)
	}
	return workspaceID.String, nil
}

func (db *DB) GetWorkspaceIDByCVID(ctx context.Context, cvID string) (string, error) {
	workspaceID, err := db.FindWorkspaceIDByCVID(ctx, sql.String(cvID))
	if err != nil {
		return "", sql.Error(err)
	}
	return workspaceID.String, nil
}

func (db *DB) GetWorkspace(ctx context.Context, workspaceID string) (*Workspace, error) {
	result, err := db.FindWorkspaceByID(ctx, sql.String(workspaceID))
	if err != nil {
		return nil, sql.Error(err)
	}
	return UnmarshalWorkspaceResult(WorkspaceResult(result))
}

func (db *DB) GetWorkspaceByName(ctx context.Context, organization, workspace string) (*Workspace, error) {
	result, err := db.FindWorkspaceByName(ctx, sql.String(workspace), sql.String(organization))
	if err != nil {
		return nil, sql.Error(err)
	}
	return UnmarshalWorkspaceResult(WorkspaceResult(result))
}

func (db *DB) DeleteWorkspace(ctx context.Context, workspaceID string) error {
	_, err := db.Querier.DeleteWorkspaceByID(ctx, sql.String(workspaceID))
	if err != nil {
		return sql.Error(err)
	}
	return nil
}

func (db *DB) GetOrganizationNameByWorkspaceID(ctx context.Context, workspaceID string) (string, error) {
	name, err := db.FindOrganizationNameByWorkspaceID(ctx, sql.String(workspaceID))
	if err != nil {
		return "", sql.Error(err)
	}
	return name.sql.String, nil
}
