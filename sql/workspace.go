package sql

import (
	"context"
	"fmt"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/leg100/otf"
)

var (
	_ otf.WorkspaceStore = (*WorkspaceDB)(nil)
)

type WorkspaceDB struct {
	*pgx.Conn
}

func NewWorkspaceDB(conn *pgx.Conn) *WorkspaceDB {
	return &WorkspaceDB{
		Conn: conn,
	}
}

func (db WorkspaceDB) Create(ws *otf.Workspace) (*otf.Workspace, error) {
	q := NewQuerier(db.Conn)
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

	tx, err := db.Conn.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	q := NewQuerier(tx)

	var row workspaceRow

	if spec.ID != nil {
		result, err := q.FindWorkspaceByIDForUpdate(ctx, *spec.ID)
		if err != nil {
			return nil, err
		}
		row = workspaceRow(result)
	} else if spec.Name != nil && spec.OrganizationName != nil {
		result, err := q.FindWorkspaceByNameForUpdate(ctx, *spec.Name, *spec.OrganizationName)
		if err != nil {
			return nil, err
		}
		row = workspaceRow(result)
	} else {
		return nil, fmt.Errorf("invalid spec")
	}

	ws := row.convert()

	if err := fn(ws); err != nil {
		return nil, err
	}

	if ws.Description != *row.Description {
		result, err := q.UpdateWorkspaceDescriptionByID(ctx, ws.Description, ws.ID)
		if err != nil {
			return nil, err
		}
		ws.UpdatedAt = result.UpdatedAt
	}

	if ws.Name != *row.Name {
		ws.UpdatedAt, err = q.UpdateWorkspaceNameByID(ctx, ws.Name, ws.ID)
		if err != nil {
			return nil, err
		}
	}

	if ws.Locked != *row.Locked {
		result, err := q.UpdateWorkspaceLockByID(ctx, ws.Locked, ws.ID)
		if err != nil {
			return nil, err
		}
		ws.UpdatedAt = result.UpdatedAt
	}

	return ws, tx.Commit(ctx)
}

func (db WorkspaceDB) List(opts otf.WorkspaceListOptions) (*otf.WorkspaceList, error) {
	q := NewQuerier(db.Conn)
	ctx := context.Background()

	result, err := q.FindWorkspaces(ctx, FindWorkspacesParams{
		OrganizationName: opts.OrganizationName,
		Prefix:           opts.Prefix,
		Limit:            100,
		Offset:           0,
	})
	if err != nil {
		return nil, err
	}

	var items []*otf.Workspace
	for _, row := range result {
		items = append(items, &otf.Workspace{
			ID: *row.WorkspaceID,
			Timestamps: otf.Timestamps{
				CreatedAt: row.CreatedAt,
				UpdatedAt: row.UpdatedAt,
			},
			AllowDestroyPlan:           *row.AllowDestroyPlan,
			AutoApply:                  *row.AutoApply,
			CanQueueDestroyPlan:        *row.CanQueueDestroyPlan,
			Description:                *row.Description,
			Environment:                *row.Environment,
			ExecutionMode:              *row.ExecutionMode,
			FileTriggersEnabled:        *row.FileTriggersEnabled,
			GlobalRemoteState:          *row.GlobalRemoteState,
			Locked:                     *row.Locked,
			MigrationEnvironment:       *row.MigrationEnvironment,
			Name:                       *row.Name,
			QueueAllRuns:               *row.QueueAllRuns,
			SpeculativeEnabled:         *row.SpeculativeEnabled,
			StructuredRunOutputEnabled: *row.StructuredRunOutputEnabled,
			SourceName:                 *row.SourceName,
			SourceURL:                  *row.SourceUrl,
			TerraformVersion:           *row.TerraformVersion,
			TriggerPrefixes:            row.TriggerPrefixes,
			WorkingDirectory:           *row.WorkingDirectory,
			Organization:               convertOrganizationComposite(row.Organization),
		})
	}

	return &otf.WorkspaceList{
		Items:      items,
		Pagination: otf.NewPagination(opts.ListOptions, getCount(result)),
	}, nil
}

func (db WorkspaceDB) Get(spec otf.WorkspaceSpec) (*otf.Workspace, error) {
	ctx := context.Background()
	q := NewQuerier(db.Conn)

	if spec.ID != nil {
		result, err := q.FindWorkspaceByID(ctx, *spec.ID)
		if err != nil {
			return nil, err
		}
		return workspaceRow(result).convert(), nil
	} else if spec.Name != nil && spec.OrganizationName != nil {
		result, err := q.FindWorkspaceByName(ctx, *spec.Name, *spec.OrganizationName)
		if err != nil {
			return nil, err
		}
		return workspaceRow(result).convert(), nil
	} else {
		return nil, fmt.Errorf("no workspace spec provided")
	}
}

// Delete deletes a specific workspace, along with its child records (runs etc).
func (db WorkspaceDB) Delete(spec otf.WorkspaceSpec) error {
	ctx := context.Background()
	q := NewQuerier(db.Conn)

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
