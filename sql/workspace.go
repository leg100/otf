package sql

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/leg100/otf"
	"github.com/leg100/otf/sql/pggen"
)

func (db *DB) CreateWorkspace(ctx context.Context, ws *otf.Workspace) error {
	_, err := db.InsertWorkspace(ctx, pggen.InsertWorkspaceParams{
		ID:                         String(ws.ID()),
		CreatedAt:                  Timestamptz(ws.CreatedAt()),
		UpdatedAt:                  Timestamptz(ws.UpdatedAt()),
		Name:                       String(ws.Name()),
		AllowDestroyPlan:           ws.AllowDestroyPlan(),
		CanQueueDestroyPlan:        ws.CanQueueDestroyPlan(),
		Environment:                String(ws.Environment()),
		Description:                String(ws.Description()),
		ExecutionMode:              String(string(ws.ExecutionMode())),
		FileTriggersEnabled:        ws.FileTriggersEnabled(),
		GlobalRemoteState:          ws.GlobalRemoteState(),
		MigrationEnvironment:       String(ws.MigrationEnvironment()),
		SourceName:                 String(ws.SourceName()),
		SourceURL:                  String(ws.SourceURL()),
		SpeculativeEnabled:         ws.SpeculativeEnabled(),
		StructuredRunOutputEnabled: ws.StructuredRunOutputEnabled(),
		TerraformVersion:           String(ws.TerraformVersion()),
		TriggerPrefixes:            ws.TriggerPrefixes(),
		QueueAllRuns:               ws.QueueAllRuns(),
		AutoApply:                  ws.AutoApply(),
		WorkingDirectory:           String(ws.WorkingDirectory()),
		OrganizationID:             String(ws.OrganizationID()),
	})
	if err != nil {
		return databaseError(err)
	}
	return nil
}

func (db *DB) UpdateWorkspace(ctx context.Context, spec otf.WorkspaceSpec, fn func(*otf.Workspace) error) (*otf.Workspace, error) {
	var ws *otf.Workspace
	err := db.tx(ctx, func(tx *DB) error {
		var err error
		// retrieve workspace
		ws, err = tx.getWorkspaceForUpdate(ctx, spec)
		if err != nil {
			return databaseError(err)
		}
		// update workspace
		if err := fn(ws); err != nil {
			return err
		}
		// persist update
		_, err = tx.UpdateWorkspaceByID(ctx, pggen.UpdateWorkspaceByIDParams{
			ID:                         String(ws.ID()),
			UpdatedAt:                  Timestamptz(ws.UpdatedAt()),
			AllowDestroyPlan:           ws.AllowDestroyPlan(),
			Description:                String(ws.Description()),
			ExecutionMode:              String(string(ws.ExecutionMode())),
			Name:                       String(ws.Name()),
			QueueAllRuns:               ws.QueueAllRuns(),
			SpeculativeEnabled:         ws.SpeculativeEnabled(),
			StructuredRunOutputEnabled: ws.StructuredRunOutputEnabled(),
			TerraformVersion:           String(ws.TerraformVersion()),
			TriggerPrefixes:            ws.TriggerPrefixes(),
			WorkingDirectory:           String(ws.WorkingDirectory()),
		})
		return err
	})
	return ws, err
}

func (db *DB) CreateWorkspaceRepo(ctx context.Context, spec otf.WorkspaceSpec, repo otf.WorkspaceRepo) (*otf.Workspace, error) {
	workspaceID, err := db.getWorkspaceID(ctx, spec)
	if err != nil {
		return nil, databaseError(err)
	}
	_, err = db.InsertWorkspaceRepo(ctx, pggen.InsertWorkspaceRepoParams{
		Branch:        String(repo.Branch),
		WebhookID:     UUID(repo.WebhookID),
		VCSProviderID: String(repo.ProviderID),
		WorkspaceID:   workspaceID,
	})
	if err != nil {
		return nil, databaseError(err)
	}
	ws, err := db.GetWorkspace(ctx, spec)
	return ws, databaseError(err)
}

func (db *DB) UpdateWorkspaceRepo(ctx context.Context, spec otf.WorkspaceSpec, repo otf.WorkspaceRepo) (*otf.Workspace, error) {
	workspaceID, err := db.getWorkspaceID(ctx, spec)
	if err != nil {
		return nil, databaseError(err)
	}
	_, err = db.UpdateWorkspaceRepoByID(ctx, String(repo.Branch), workspaceID)
	if err != nil {
		return nil, databaseError(err)
	}
	ws, err := db.GetWorkspace(ctx, spec)
	return ws, databaseError(err)
}

func (db *DB) DeleteWorkspaceRepo(ctx context.Context, spec otf.WorkspaceSpec) (*otf.Workspace, error) {
	id, err := db.getWorkspaceID(ctx, spec)
	if err != nil {
		return nil, databaseError(err)
	}
	_, err = db.Querier.DeleteWorkspaceRepo(ctx, id)
	if err != nil {
		return nil, databaseError(err)
	}
	ws, err := db.GetWorkspace(ctx, spec)
	return ws, databaseError(err)
}

// LockWorkspace locks the specified workspace.
func (db *DB) LockWorkspace(ctx context.Context, spec otf.WorkspaceSpec, opts otf.WorkspaceLockOptions) (*otf.Workspace, error) {
	subj, err := otf.LockFromContext(ctx)
	if err != nil {
		return nil, err
	}

	var ws *otf.Workspace
	err = db.tx(ctx, func(tx *DB) error {
		// retrieve workspace
		ws, err = tx.getWorkspaceForUpdate(ctx, spec)
		if err != nil {
			return databaseError(err)
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
			return databaseError(err)
		}
		return nil
	})
	// return ws with new lock
	return ws, err
}

// SetCurrentRun sets the ID of the current run for the specified workspace.
func (db *DB) SetCurrentRun(ctx context.Context, workspaceID, runID string) error {
	_, err := db.UpdateWorkspaceLatestRun(ctx, String(runID), String(workspaceID))
	if err != nil {
		return databaseError(err)
	}
	return nil
}

// UnlockWorkspace unlocks the specified workspace; the caller has the
// opportunity to check the current locker passed into the provided callback. If
// an error is returned the unlock will not go ahead.
func (db *DB) UnlockWorkspace(ctx context.Context, spec otf.WorkspaceSpec, opts otf.WorkspaceUnlockOptions) (*otf.Workspace, error) {
	subj, err := otf.LockFromContext(ctx)
	if err != nil {
		return nil, err
	}

	var ws *otf.Workspace
	err = db.tx(ctx, func(tx *DB) error {
		// retrieve workspace
		ws, err = db.getWorkspaceForUpdate(ctx, spec)
		if err != nil {
			return databaseError(err)
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
			return databaseError(err)
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
	if opts.OrganizationName != nil {
		organizationName = *opts.OrganizationName
	} else {
		organizationName = "%"
	}

	db.FindWorkspacesBatch(batch, pggen.FindWorkspacesParams{
		OrganizationNames: []string{organizationName},
		Prefix:            String(opts.Prefix),
		Limit:             opts.GetLimit(),
		Offset:            opts.GetOffset(),
	})
	db.CountWorkspacesBatch(batch, String(opts.Prefix), []string{organizationName})
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

	var items []*otf.Workspace
	for _, r := range rows {
		ws, err := otf.UnmarshalWorkspaceResult(otf.WorkspaceResult(r))
		if err != nil {
			return nil, err
		}
		items = append(items, ws)
	}

	return &otf.WorkspaceList{
		Items:      items,
		Pagination: otf.NewPagination(opts.ListOptions, *count),
	}, nil
}

func (db *DB) ListWorkspacesByWebhookID(ctx context.Context, id string) ([]*otf.Workspace, error) {
	rows, err := db.FindWorkspacesByWebhookID(ctx, UUID(uuid.MustParse(id)))
	if err != nil {
		return nil, err
	}

	var items []*otf.Workspace
	for _, r := range rows {
		ws, err := otf.UnmarshalWorkspaceResult(otf.WorkspaceResult(r))
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
		OrganizationName: String(organization),
		UserID:           String(userID),
		Limit:            opts.GetLimit(),
		Offset:           opts.GetOffset(),
	})
	db.CountWorkspacesByUserIDBatch(batch, String(organization), String(userID))
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

	var items []*otf.Workspace
	for _, r := range rows {
		ws, err := otf.UnmarshalWorkspaceResult(otf.WorkspaceResult(r))
		if err != nil {
			return nil, err
		}
		items = append(items, ws)
	}

	return &otf.WorkspaceList{
		Items:      items,
		Pagination: otf.NewPagination(opts, *count),
	}, nil
}

func (db *DB) GetWorkspaceIDByRunID(ctx context.Context, runID string) (string, error) {
	workspaceID, err := db.FindWorkspaceIDByRunID(ctx, String(runID))
	if err != nil {
		return "", databaseError(err)
	}
	return workspaceID.String, nil
}

func (db *DB) GetWorkspaceIDByStateVersionID(ctx context.Context, svID string) (string, error) {
	workspaceID, err := db.FindWorkspaceIDByStateVersionID(ctx, String(svID))
	if err != nil {
		return "", databaseError(err)
	}
	return workspaceID.String, nil
}

func (db *DB) GetWorkspaceIDByCVID(ctx context.Context, cvID string) (string, error) {
	workspaceID, err := db.FindWorkspaceIDByCVID(ctx, String(cvID))
	if err != nil {
		return "", databaseError(err)
	}
	return workspaceID.String, nil
}

func (db *DB) GetWorkspaceID(ctx context.Context, spec otf.WorkspaceSpec) (string, error) {
	if spec.ID != nil {
		return *spec.ID, nil
	}
	if spec.Name != nil && spec.OrganizationName != nil {
		id, err := db.FindWorkspaceIDByName(ctx, String(*spec.Name), String(*spec.OrganizationName))
		if err != nil {
			return "", err
		}
		return id.String, nil
	}
	return "", otf.ErrInvalidWorkspaceSpec
}

func (db *DB) GetWorkspace(ctx context.Context, spec otf.WorkspaceSpec) (*otf.Workspace, error) {
	if spec.ID != nil {
		result, err := db.FindWorkspaceByID(ctx, String(*spec.ID))
		if err != nil {
			return nil, databaseError(err)
		}
		return otf.UnmarshalWorkspaceResult(otf.WorkspaceResult(result))
	} else if spec.Name != nil && spec.OrganizationName != nil {
		result, err := db.FindWorkspaceByName(ctx, String(*spec.Name), String(*spec.OrganizationName))
		if err != nil {
			return nil, databaseError(err)
		}
		return otf.UnmarshalWorkspaceResult(otf.WorkspaceResult(result))
	} else {
		return nil, fmt.Errorf("no workspace spec provided")
	}
}

// DeleteWorkspace deletes a specific workspace, along with its child records
// (runs etc).
func (db *DB) DeleteWorkspace(ctx context.Context, spec otf.WorkspaceSpec) error {
	var err error
	if spec.ID != nil {
		_, err = db.DeleteWorkspaceByID(ctx, String(*spec.ID))
	} else if spec.Name != nil && spec.OrganizationName != nil {
		_, err = db.DeleteWorkspaceByName(ctx, String(*spec.Name), String(*spec.OrganizationName))
	} else {
		return fmt.Errorf("no workspace spec provided")
	}
	if err != nil {
		return databaseError(err)
	}
	return nil
}

func (db *DB) getWorkspaceForUpdate(ctx context.Context, spec otf.WorkspaceSpec) (*otf.Workspace, error) {
	workspaceID, err := db.getWorkspaceID(ctx, spec)
	if err != nil {
		return nil, err
	}
	result, err := db.FindWorkspaceByIDForUpdate(ctx, workspaceID)
	if err != nil {
		return nil, err
	}
	return otf.UnmarshalWorkspaceResult(otf.WorkspaceResult(result))
}

func (db *DB) getWorkspaceID(ctx context.Context, spec otf.WorkspaceSpec) (pgtype.Text, error) {
	if spec.ID != nil {
		return String(*spec.ID), nil
	}
	if spec.Name != nil && spec.OrganizationName != nil {
		return db.FindWorkspaceIDByName(ctx, String(*spec.Name), String(*spec.OrganizationName))
	}
	return pgtype.Text{}, otf.ErrInvalidWorkspaceSpec
}
