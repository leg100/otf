package sql

import (
	"context"
	"fmt"

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
		if err != nil {
			return err
		}
		return nil
	})
	return ws, err
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

// SetLatestRun sets the ID of the latest run for the specified workspace.
func (db *DB) SetLatestRun(ctx context.Context, workspaceID, runID string) error {
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
		OrganizationNames:   []string{organizationName},
		Prefix:              String(opts.Prefix),
		Limit:               opts.GetLimit(),
		Offset:              opts.GetOffset(),
		IncludeOrganization: includeOrganization(opts.Include),
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
		ws, err := otf.UnmarshalWorkspaceDBResult(otf.WorkspaceDBResult(r))
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

func (db *DB) GetWorkspace(ctx context.Context, spec otf.WorkspaceSpec) (*otf.Workspace, error) {
	if spec.ID != nil {
		// TODO: always include the organization regardless of whether caller
		// specified it. The complexity isn't worth the performance saving.
		result, err := db.FindWorkspaceByID(ctx, includeOrganization(spec.Include), String(*spec.ID))
		if err != nil {
			return nil, databaseError(err)
		}
		return otf.UnmarshalWorkspaceDBResult(otf.WorkspaceDBResult(result))
	} else if spec.Name != nil && spec.OrganizationName != nil {
		result, err := db.FindWorkspaceByName(ctx, pggen.FindWorkspaceByNameParams{
			Name:                String(*spec.Name),
			OrganizationName:    String(*spec.OrganizationName),
			IncludeOrganization: includeOrganization(spec.Include),
		})
		if err != nil {
			return nil, databaseError(err)
		}
		return otf.UnmarshalWorkspaceDBResult(otf.WorkspaceDBResult(result))
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
	return otf.UnmarshalWorkspaceDBResult(otf.WorkspaceDBResult(result))
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
