package sql

import (
	"context"
	"fmt"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/leg100/otf"
	"github.com/leg100/otf/sql/pggen"
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
	q := pggen.NewQuerier(db.Pool)
	ctx := context.Background()
	_, err := q.InsertWorkspace(ctx, pggen.InsertWorkspaceParams{
		ID:                         ws.ID(),
		CreatedAt:                  ws.CreatedAt(),
		UpdatedAt:                  ws.UpdatedAt(),
		Name:                       ws.Name(),
		AllowDestroyPlan:           ws.AllowDestroyPlan(),
		CanQueueDestroyPlan:        ws.CanQueueDestroyPlan(),
		Environment:                ws.Environment(),
		Description:                ws.Description(),
		ExecutionMode:              string(ws.ExecutionMode()),
		FileTriggersEnabled:        ws.FileTriggersEnabled(),
		GlobalRemoteState:          ws.GlobalRemoteState(),
		Locked:                     ws.Locked(),
		MigrationEnvironment:       ws.MigrationEnvironment(),
		SourceName:                 ws.SourceName(),
		SourceURL:                  ws.SourceURL(),
		SpeculativeEnabled:         ws.SpeculativeEnabled(),
		StructuredRunOutputEnabled: ws.StructuredRunOutputEnabled(),
		TerraformVersion:           ws.TerraformVersion(),
		TriggerPrefixes:            ws.TriggerPrefixes(),
		QueueAllRuns:               ws.QueueAllRuns(),
		AutoApply:                  ws.AutoApply(),
		WorkingDirectory:           ws.WorkingDirectory(),
		OrganizationID:             ws.Organization.ID(),
	})
	if err != nil {
		return nil, databaseError(err)
	}
	return ws, nil
}

func (db WorkspaceDB) Update(spec otf.WorkspaceSpec, fn func(*otf.Workspace) error) (*otf.Workspace, error) {
	ctx := context.Background()

	tx, err := db.Pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	q := pggen.NewQuerier(tx)

	var ws *otf.Workspace
	if spec.ID != nil {
		result, err := q.FindWorkspaceByIDForUpdate(ctx, *spec.ID)
		if err != nil {
			return nil, err
		}
		ws, err = otf.UnmarshalWorkspaceDBType(pggen.Workspaces(result))
		if err != nil {
			return nil, err
		}
	} else if spec.Name != nil && spec.OrganizationName != nil {
		result, err := q.FindWorkspaceByNameForUpdate(ctx, *spec.Name, *spec.OrganizationName)
		if err != nil {
			return nil, err
		}
		ws, err = otf.UnmarshalWorkspaceDBType(pggen.Workspaces(result))
		if err != nil {
			return nil, err
		}
	} else {
		return nil, fmt.Errorf("invalid spec")
	}
	if err := fn(ws); err != nil {
		return nil, err
	}
	_, err = q.UpdateWorkspaceByID(ctx, pggen.UpdateWorkspaceByIDParams{
		ID:                         ws.ID(),
		UpdatedAt:                  ws.UpdatedAt(),
		AllowDestroyPlan:           ws.AllowDestroyPlan(),
		Description:                ws.Description(),
		ExecutionMode:              string(ws.ExecutionMode()),
		Locked:                     ws.Locked(),
		Name:                       ws.Name(),
		QueueAllRuns:               ws.QueueAllRuns(),
		SpeculativeEnabled:         ws.SpeculativeEnabled(),
		StructuredRunOutputEnabled: ws.StructuredRunOutputEnabled(),
		TerraformVersion:           ws.TerraformVersion(),
		TriggerPrefixes:            ws.TriggerPrefixes(),
		WorkingDirectory:           ws.WorkingDirectory(),
	})
	if err != nil {
		return nil, err
	}

	return ws, tx.Commit(ctx)
}

func (db WorkspaceDB) List(opts otf.WorkspaceListOptions) (*otf.WorkspaceList, error) {
	q := pggen.NewQuerier(db.Pool)
	batch := &pgx.Batch{}
	ctx := context.Background()

	q.FindWorkspacesBatch(batch, pggen.FindWorkspacesParams{
		OrganizationName:    opts.OrganizationName,
		Prefix:              opts.Prefix,
		Limit:               opts.GetLimit(),
		Offset:              opts.GetOffset(),
		IncludeOrganization: includeOrganization(opts.Include),
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

func (db WorkspaceDB) Get(spec otf.WorkspaceSpec) (*otf.Workspace, error) {
	ctx := context.Background()
	q := pggen.NewQuerier(db.Pool)

	if spec.ID != nil {
		result, err := q.FindWorkspaceByID(ctx, includeOrganization(spec.Include), *spec.ID)
		if err != nil {
			return nil, err
		}
		return otf.UnmarshalWorkspaceDBResult(otf.WorkspaceDBResult(result))
	} else if spec.Name != nil && spec.OrganizationName != nil {
		result, err := q.FindWorkspaceByName(ctx, pggen.FindWorkspaceByNameParams{
			Name:                *spec.Name,
			OrganizationName:    *spec.OrganizationName,
			IncludeOrganization: includeOrganization(spec.Include),
		})
		if err != nil {
			return nil, err
		}
		return otf.UnmarshalWorkspaceDBResult(otf.WorkspaceDBResult(result))
	} else {
		return nil, fmt.Errorf("no workspace spec provided")
	}
}

// Delete deletes a specific workspace, along with its child records (runs etc).
func (db WorkspaceDB) Delete(spec otf.WorkspaceSpec) error {
	ctx := context.Background()
	q := pggen.NewQuerier(db.Pool)

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
