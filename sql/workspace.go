package sql

import (
	"context"
	"fmt"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/leg100/otf"
	"github.com/leg100/otf/sql/pggen"
)

func (db *DB) CreateWorkspace(ctx context.Context, ws *otf.Workspace) error {
	_, err := db.InsertWorkspace(ctx, pggen.InsertWorkspaceParams{
		ID:                         pgtype.Text{String: ws.ID(), Status: pgtype.Present},
		CreatedAt:                  ws.CreatedAt(),
		UpdatedAt:                  ws.UpdatedAt(),
		Name:                       pgtype.Text{String: ws.Name(), Status: pgtype.Present},
		AllowDestroyPlan:           ws.AllowDestroyPlan(),
		CanQueueDestroyPlan:        ws.CanQueueDestroyPlan(),
		Environment:                pgtype.Text{String: ws.Environment(), Status: pgtype.Present},
		Description:                pgtype.Text{String: ws.Description(), Status: pgtype.Present},
		ExecutionMode:              pgtype.Text{String: string(ws.ExecutionMode()), Status: pgtype.Present},
		FileTriggersEnabled:        ws.FileTriggersEnabled(),
		GlobalRemoteState:          ws.GlobalRemoteState(),
		MigrationEnvironment:       pgtype.Text{String: ws.MigrationEnvironment(), Status: pgtype.Present},
		SourceName:                 pgtype.Text{String: ws.SourceName(), Status: pgtype.Present},
		SourceURL:                  pgtype.Text{String: ws.SourceURL(), Status: pgtype.Present},
		SpeculativeEnabled:         ws.SpeculativeEnabled(),
		StructuredRunOutputEnabled: ws.StructuredRunOutputEnabled(),
		TerraformVersion:           pgtype.Text{String: ws.TerraformVersion(), Status: pgtype.Present},
		TriggerPrefixes:            ws.TriggerPrefixes(),
		QueueAllRuns:               ws.QueueAllRuns(),
		AutoApply:                  ws.AutoApply(),
		WorkingDirectory:           pgtype.Text{String: ws.WorkingDirectory(), Status: pgtype.Present},
		OrganizationID:             pgtype.Text{String: ws.OrganizationID(), Status: pgtype.Present},
	})
	if err != nil {
		return databaseError(err)
	}
	return nil
}

func (db *DB) UpdateWorkspace(ctx context.Context, spec otf.WorkspaceSpec, fn func(*otf.Workspace) error) (*otf.Workspace, error) {
	tx, err := db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)
	q := pggen.NewQuerier(tx)
	// retrieve workspace
	ws, err := db.getWorkspaceForUpdate(ctx, tx, spec)
	if err != nil {
		return nil, databaseError(err)
	}
	// update workspace
	if err := fn(ws); err != nil {
		return nil, err
	}
	// persist update
	_, err = q.UpdateWorkspaceByID(ctx, pggen.UpdateWorkspaceByIDParams{
		ID:                         pgtype.Text{String: ws.ID(), Status: pgtype.Present},
		UpdatedAt:                  ws.UpdatedAt(),
		AllowDestroyPlan:           ws.AllowDestroyPlan(),
		Description:                pgtype.Text{String: ws.Description(), Status: pgtype.Present},
		ExecutionMode:              pgtype.Text{String: string(ws.ExecutionMode()), Status: pgtype.Present},
		Name:                       pgtype.Text{String: ws.Name(), Status: pgtype.Present},
		QueueAllRuns:               ws.QueueAllRuns(),
		SpeculativeEnabled:         ws.SpeculativeEnabled(),
		StructuredRunOutputEnabled: ws.StructuredRunOutputEnabled(),
		TerraformVersion:           pgtype.Text{String: ws.TerraformVersion(), Status: pgtype.Present},
		TriggerPrefixes:            ws.TriggerPrefixes(),
		WorkingDirectory:           pgtype.Text{String: ws.WorkingDirectory(), Status: pgtype.Present},
	})
	if err != nil {
		return nil, err
	}
	return ws, tx.Commit(ctx)
}

// LockWorkspace locks the specified workspace.
func (db *DB) LockWorkspace(ctx context.Context, spec otf.WorkspaceSpec, opts otf.WorkspaceLockOptions) (*otf.Workspace, error) {
	tx, err := db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)
	q := pggen.NewQuerier(tx)
	// retrieve workspace
	ws, err := db.getWorkspaceForUpdate(ctx, tx, spec)
	if err != nil {
		return nil, databaseError(err)
	}
	// lock the workspace
	if err := ws.Lock(opts.Requestor); err != nil {
		return nil, err
	}
	// persist to db
	params, err := otf.MarshalWorkspaceLockParams(ws)
	if err != nil {
		return nil, err
	}
	_, err = q.UpdateWorkspaceLockByID(ctx, params)
	if err != nil {
		return nil, databaseError(err)
	}
	// return ws with new lock
	return ws, tx.Commit(ctx)
}

// UnlockWorkspace unlocks the specified workspace; the caller has the
// opportunity to check the current locker passed into the provided callback. If
// an error is returned the unlock will not go ahead.
func (db *DB) UnlockWorkspace(ctx context.Context, spec otf.WorkspaceSpec, opts otf.WorkspaceUnlockOptions) (*otf.Workspace, error) {
	tx, err := db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)
	q := pggen.NewQuerier(tx)
	// retrieve workspace
	ws, err := db.getWorkspaceForUpdate(ctx, tx, spec)
	if err != nil {
		return nil, databaseError(err)
	}
	// unlock workspace
	if err := ws.Unlock(opts.Requestor, opts.Force); err != nil {
		return nil, err
	}
	// persist to db
	params, err := otf.MarshalWorkspaceLockParams(ws)
	if err != nil {
		return nil, err
	}
	_, err = q.UpdateWorkspaceLockByID(ctx, params)
	if err != nil {
		return nil, databaseError(err)
	}
	// return ws with new lock
	return ws, tx.Commit(ctx)
}

func (db *DB) ListWorkspaces(ctx context.Context, opts otf.WorkspaceListOptions) (*otf.WorkspaceList, error) {
	q := pggen.NewQuerier(db)
	batch := &pgx.Batch{}

	q.FindWorkspacesBatch(batch, pggen.FindWorkspacesParams{
		OrganizationName:    pgtype.Text{String: opts.OrganizationName, Status: pgtype.Present},
		Prefix:              pgtype.Text{String: opts.Prefix, Status: pgtype.Present},
		Limit:               opts.GetLimit(),
		Offset:              opts.GetOffset(),
		IncludeOrganization: includeOrganization(opts.Include),
	})
	q.CountWorkspacesBatch(batch,
		pgtype.Text{String: opts.Prefix, Status: pgtype.Present},
		pgtype.Text{String: opts.OrganizationName, Status: pgtype.Present},
	)

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
	q := pggen.NewQuerier(db)

	if spec.ID != nil {
		result, err := q.FindWorkspaceByID(ctx,
			includeOrganization(spec.Include),
			pgtype.Text{String: *spec.ID, Status: pgtype.Present},
		)
		if err != nil {
			return nil, databaseError(err)
		}
		return otf.UnmarshalWorkspaceDBResult(otf.WorkspaceDBResult(result))
	} else if spec.Name != nil && spec.OrganizationName != nil {
		result, err := q.FindWorkspaceByName(ctx, pggen.FindWorkspaceByNameParams{
			Name:                pgtype.Text{String: *spec.Name, Status: pgtype.Present},
			OrganizationName:    pgtype.Text{String: *spec.OrganizationName, Status: pgtype.Present},
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
	var result pgconn.CommandTag
	var err error

	if spec.ID != nil {
		result, err = db.DeleteWorkspaceByID(ctx, pgtype.Text{String: *spec.ID, Status: pgtype.Present})
	} else if spec.Name != nil && spec.OrganizationName != nil {
		result, err = db.DeleteWorkspaceByName(ctx,
			pgtype.Text{String: *spec.Name, Status: pgtype.Present},
			pgtype.Text{String: *spec.OrganizationName, Status: pgtype.Present},
		)
	} else {
		return fmt.Errorf("no workspace spec provided")
	}
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return otf.ErrResourceNotFound
	}

	return nil
}

func (db *DB) getWorkspaceForUpdate(ctx context.Context, tx pgx.Tx, spec otf.WorkspaceSpec) (*otf.Workspace, error) {
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
		return pgtype.Text{String: *spec.ID, Status: pgtype.Present}, nil
	}
	if spec.Name != nil && spec.OrganizationName != nil {
		return db.FindWorkspaceIDByName(ctx,
			pgtype.Text{String: *spec.Name, Status: pgtype.Present},
			pgtype.Text{String: *spec.OrganizationName, Status: pgtype.Present},
		)
	}
	return pgtype.Text{}, otf.ErrInvalidWorkspaceSpec
}
