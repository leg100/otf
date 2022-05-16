package sql

import (
	"context"
	"fmt"
	"reflect"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/leg100/otf"
)

var (
	_ otf.WorkspaceStore = (*WorkspaceDB)(nil)
)

type WorkspaceDB struct {
	*pgxpool.Pool
}

func NewWorkspaceDB(conn *pgxpool.Pool) *WorkspaceDB {
	return &WorkspaceDB{
		Pool: conn,
	}
}

func (db WorkspaceDB) Create(ws *otf.Workspace) (*otf.Workspace, error) {
	q := NewQuerier(db.Pool)
	ctx := context.Background()

	result, err := q.InsertWorkspace(ctx, InsertWorkspaceParams{
		ID:                         ws.ID,
		Name:                       ws.Name,
		AllowDestroyPlan:           ws.AllowDestroyPlan,
		CanQueueDestroyPlan:        ws.CanQueueDestroyPlan,
		Environment:                ws.Environment,
		Description:                ws.Description,
		ExecutionMode:              ws.ExecutionMode,
		FileTriggersEnabled:        ws.FileTriggersEnabled,
		GlobalRemoteState:          ws.GlobalRemoteState,
		Locked:                     ws.Locked,
		MigrationEnvironment:       ws.MigrationEnvironment,
		SourceName:                 ws.SourceName,
		SourceUrl:                  ws.SourceURL,
		SpeculativeEnabled:         ws.SpeculativeEnabled,
		StructuredRunOutputEnabled: ws.StructuredRunOutputEnabled,
		TerraformVersion:           ws.TerraformVersion,
		TriggerPrefixes:            ws.TriggerPrefixes,
		QueueAllRuns:               ws.QueueAllRuns,
		AutoApply:                  ws.AutoApply,
		WorkingDirectory:           ws.WorkingDirectory,
		OrganizationID:             ws.Organization.ID,
	})
	if err != nil {
		return nil, databaseError(err, insertWorkspaceSQL)
	}
	ws.CreatedAt = result.CreatedAt
	ws.UpdatedAt = result.UpdatedAt

	return ws, nil
}

func (db WorkspaceDB) Update(spec otf.WorkspaceSpec, fn func(*otf.Workspace) error) (*otf.Workspace, error) {
	ctx := context.Background()

	tx, err := db.Pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	q := NewQuerier(tx)

	var result interface{}
	if spec.ID != nil {
		result, err = q.FindWorkspaceByIDForUpdate(ctx, *spec.ID)
	} else if spec.Name != nil && spec.OrganizationName != nil {
		result, err = q.FindWorkspaceByNameForUpdate(ctx, *spec.Name, *spec.OrganizationName)
	} else {
		return nil, fmt.Errorf("invalid spec")
	}
	if err != nil {
		return nil, err
	}
	ws, err := otf.UnmarshalWorkspaceFromDB(result)
	if err != nil {
		return nil, err
	}

	// make copies of workspace attributes for comparison after update
	xdescription := ws.Description
	xname := ws.Name
	xexecutionMode := ws.ExecutionMode
	xlocked := ws.Locked
	xworkingDirectory := ws.WorkingDirectory
	xqueueAllRuns := ws.QueueAllRuns
	xspeculativeEnabled := ws.SpeculativeEnabled
	xstructuredRunOutputEnabled := ws.StructuredRunOutputEnabled
	xterraformVersion := ws.TerraformVersion
	xtriggerPrefixes := ws.TriggerPrefixes

	if err := fn(ws); err != nil {
		return nil, err
	}

	if ws.Description != xdescription {
		ws.UpdatedAt, err = q.UpdateWorkspaceDescriptionByID(ctx, ws.Description, ws.ID)
		if err != nil {
			return nil, err
		}
	}

	if ws.ExecutionMode != xexecutionMode {
		ws.UpdatedAt, err = q.UpdateWorkspaceExecutionModeByID(ctx, ws.ExecutionMode, ws.ID)
		if err != nil {
			return nil, err
		}
	}

	if ws.Name != xname {
		ws.UpdatedAt, err = q.UpdateWorkspaceNameByID(ctx, ws.Name, ws.ID)
		if err != nil {
			return nil, err
		}
	}

	if ws.Locked != xlocked {
		ws.UpdatedAt, err = q.UpdateWorkspaceLockByID(ctx, ws.Locked, ws.ID)
		if err != nil {
			return nil, err
		}
	}

	if ws.QueueAllRuns != xqueueAllRuns {
		ws.UpdatedAt, err = q.UpdateWorkspaceLockByID(ctx, ws.QueueAllRuns, ws.ID)
		if err != nil {
			return nil, err
		}
	}

	if !reflect.DeepEqual(ws.TriggerPrefixes, xtriggerPrefixes) {
		ws.UpdatedAt, err = q.UpdateWorkspaceTriggerPrefixesByID(ctx, ws.TriggerPrefixes, ws.ID)
		if err != nil {
			return nil, err
		}
	}

	if ws.SpeculativeEnabled != xspeculativeEnabled {
		ws.UpdatedAt, err = q.UpdateWorkspaceSpeculativeEnabledByID(ctx, ws.SpeculativeEnabled, ws.ID)
		if err != nil {
			return nil, err
		}
	}

	if ws.StructuredRunOutputEnabled != xstructuredRunOutputEnabled {
		ws.UpdatedAt, err = q.UpdateWorkspaceStructuredRunOutputEnabledByID(ctx, ws.StructuredRunOutputEnabled, ws.ID)
		if err != nil {
			return nil, err
		}
	}

	if ws.TerraformVersion != xterraformVersion {
		ws.UpdatedAt, err = q.UpdateWorkspaceTerraformVersionByID(ctx, ws.TerraformVersion, ws.ID)
		if err != nil {
			return nil, err
		}
	}

	if ws.WorkingDirectory != xworkingDirectory {
		ws.UpdatedAt, err = q.UpdateWorkspaceWorkingDirectoryByID(ctx, ws.WorkingDirectory, ws.ID)
		if err != nil {
			return nil, err
		}
	}

	return ws, tx.Commit(ctx)
}

func (db WorkspaceDB) List(opts otf.WorkspaceListOptions) (*otf.WorkspaceList, error) {
	q := NewQuerier(db.Pool)
	batch := &pgx.Batch{}
	ctx := context.Background()

	q.FindWorkspacesBatch(batch, FindWorkspacesParams{
		OrganizationName: opts.OrganizationName,
		Prefix:           opts.Prefix,
		Limit:            opts.GetLimit(),
		Offset:           opts.GetOffset(),
	})
	q.CountWorkspacesBatch(batch, opts.Prefix, opts.OrganizationName)

	results := db.Pool.SendBatch(ctx, batch)
	defer results.Close()

	rows, err := q.FindWorkspacesScan(results)
	if err != nil {
		return nil, err
	}
	count, err := q.CountWorkspacesScan(results)
	if err != nil {
		return nil, err
	}

	workspaces, err := otf.UnmarshalWorkspaceListFromDB(rows)
	if err != nil {
		return nil, err
	}

	return &otf.WorkspaceList{
		Items:      workspaces,
		Pagination: otf.NewPagination(opts.ListOptions, *count),
	}, nil
}

func (db WorkspaceDB) Get(spec otf.WorkspaceSpec) (*otf.Workspace, error) {
	ctx := context.Background()
	q := NewQuerier(db.Pool)

	if spec.ID != nil {
		result, err := q.FindWorkspaceByID(ctx, *spec.ID)
		if err != nil {
			return nil, err
		}
		return otf.UnmarshalWorkspaceFromDB(result)
	} else if spec.Name != nil && spec.OrganizationName != nil {
		result, err := q.FindWorkspaceByName(ctx, *spec.Name, *spec.OrganizationName)
		if err != nil {
			return nil, err
		}
		return otf.UnmarshalWorkspaceFromDB(result)
	} else {
		return nil, fmt.Errorf("no workspace spec provided")
	}
}

// Delete deletes a specific workspace, along with its child records (runs etc).
func (db WorkspaceDB) Delete(spec otf.WorkspaceSpec) error {
	ctx := context.Background()
	q := NewQuerier(db.Pool)

	var result pgconn.CommandTag
	var err error

	if spec.ID != nil {
		result, err = q.DeleteWorkspaceByID(ctx, *spec.ID)
	} else if spec.Name != nil && spec.OrganizationName != nil {
		result, err = q.DeleteWorkspaceByName(ctx, *spec.Name, *spec.OrganizationName)
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
