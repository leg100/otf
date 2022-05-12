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

type workspaceComposite interface {
	GetWorkspaceID() *string
	GetAllowDestroyPlan() *bool
	GetAutoApply() *bool
	GetCanQueueDestroyPlan() *bool
	GetDescription() *string
	GetEnvironment() *string
	GetExecutionMode() *string
	GetFileTriggersEnabled() *bool
	GetGlobalRemoteState() *bool
	GetLocked() *bool
	GetMigrationEnvironment() *string
	GetName() *string
	GetQueueAllRuns() *bool
	GetSpeculativeEnabled() *bool
	GetSourceName() *string
	GetSourceUrl() *string
	GetStructuredRunOutputEnabled() *bool
	GetTerraformVersion() *string
	GetTriggerPrefixes() []string
	GetWorkingDirectory() *string
	GetOrganizationID() *string

	Timestamps
}

type workspaceRow interface {
	workspaceComposite

	GetOrganization() Organizations
}

type workspaceRowList interface {
	workspaceRow

	GetFullCount() *int
}

func NewWorkspaceDB(conn *pgx.Conn) *WorkspaceDB {
	return &WorkspaceDB{
		Conn: conn,
	}
}

func (db WorkspaceDB) Create(ws *otf.Workspace) (*otf.Workspace, error) {
	q := NewQuerier(db.Conn)
	ctx := context.Background()

	_, err := q.InsertWorkspace(ctx, InsertWorkspaceParams{
		ID:             &ws.ID,
		Name:           &ws.Name,
		OrganizationID: &ws.Organization.ID,
	})
	if err != nil {
		return nil, err
	}

	// Return newly created workspcae to caller
	return getWorkspace(ctx, q, otf.WorkspaceSpec{ID: &ws.ID})
}

func (db WorkspaceDB) Update(spec otf.WorkspaceSpec, fn func(*otf.Workspace, otf.WorkspaceUpdater) error) (*otf.Workspace, error) {
	ctx := context.Background()

	tx, err := db.Conn.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	q := NewQuerier(tx)

	var ws *otf.Workspace
	if spec.ID != nil {
		result, err := q.FindWorkspaceByIDForUpdate(ctx, spec.ID)
		if err != nil {
			return nil, err
		}
		ws = convertWorkspace(result)
	} else if spec.Name != nil && spec.OrganizationName != nil {
		result, err := q.FindWorkspaceByNameForUpdate(ctx, spec.Name, spec.OrganizationName)
		if err != nil {
			return nil, err
		}
		ws = convertWorkspace(result)
	} else {
		return nil, fmt.Errorf("invalid spec")
	}

	updater := newWorkspaceUpdater(tx, ws.ID)

	if err := fn(ws, updater); err != nil {
		return nil, err
	}

	return convertWorkspaceComposite(updater.result), tx.Commit(ctx)
}

func (db WorkspaceDB) List(opts otf.WorkspaceListOptions) (*otf.WorkspaceList, error) {
	q := NewQuerier(db.Conn)
	ctx := context.Background()

	result, err := q.FindWorkspaces(ctx, FindWorkspacesParams{
		OrganizationName: opts.OrganizationName,
		Prefix:           opts.Prefix,
		Limit:            opts.GetLimit(),
		Offset:           opts.GetOffset(),
	})
	if err != nil {
		return nil, err
	}

	var items []*otf.Workspace
	for _, r := range result {
		items = append(items, convertWorkspace(r))
	}

	return &otf.WorkspaceList{
		Items:      items,
		Pagination: otf.NewPagination(opts.ListOptions, getCount(result)),
	}, nil
}

func (db WorkspaceDB) Get(spec otf.WorkspaceSpec) (*otf.Workspace, error) {
	ctx := context.Background()
	q := NewQuerier(db.Conn)

	return getWorkspace(ctx, q, spec)
}

// Delete deletes a specific workspace, along with its child records (runs etc).
func (db WorkspaceDB) Delete(spec otf.WorkspaceSpec) error {
	ctx := context.Background()
	q := NewQuerier(db.Conn)

	var result pgconn.CommandTag
	var err error

	if spec.ID != nil {
		_, err = q.DeleteWorkspaceByID(ctx, spec.ID)
	} else if spec.Name != nil && spec.OrganizationName != nil {
		result, err = q.DeleteWorkspaceByName(ctx, spec.Name, spec.OrganizationName)
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

func getWorkspace(ctx context.Context, q *DBQuerier, spec otf.WorkspaceSpec) (*otf.Workspace, error) {
	if spec.ID != nil {
		result, err := q.FindWorkspaceByID(ctx, spec.ID)
		if err != nil {
			return nil, err
		}
		return convertWorkspace(result), nil
	} else if spec.Name != nil && spec.OrganizationName != nil {
		result, err := q.FindWorkspaceByName(ctx, spec.Name, spec.OrganizationName)
		if err != nil {
			return nil, err
		}
		return convertWorkspace(result), nil
	} else {
		return nil, fmt.Errorf("no workspace spec provided")
	}
}

func convertWorkspaceComposite(row workspaceComposite) *otf.Workspace {
	ws := otf.Workspace{
		ID:                         *row.GetWorkspaceID(),
		Timestamps:                 convertTimestamps(row),
		AllowDestroyPlan:           *row.GetAllowDestroyPlan(),
		AutoApply:                  *row.GetAutoApply(),
		CanQueueDestroyPlan:        *row.GetCanQueueDestroyPlan(),
		Description:                *row.GetDescription(),
		Environment:                *row.GetEnvironment(),
		ExecutionMode:              *row.GetExecutionMode(),
		FileTriggersEnabled:        *row.GetFileTriggersEnabled(),
		GlobalRemoteState:          *row.GetGlobalRemoteState(),
		Locked:                     *row.GetLocked(),
		MigrationEnvironment:       *row.GetMigrationEnvironment(),
		Name:                       *row.GetName(),
		QueueAllRuns:               *row.GetQueueAllRuns(),
		SpeculativeEnabled:         *row.GetSpeculativeEnabled(),
		StructuredRunOutputEnabled: *row.GetStructuredRunOutputEnabled(),
		SourceName:                 *row.GetSourceName(),
		SourceURL:                  *row.GetSourceUrl(),
		TerraformVersion:           *row.GetTerraformVersion(),
		TriggerPrefixes:            row.GetTriggerPrefixes(),
		WorkingDirectory:           *row.GetWorkingDirectory(),
	}

	return &ws
}

func convertWorkspace(row workspaceRow) *otf.Workspace {
	ws := convertWorkspaceComposite(row)
	ws.Organization = convertOrganization(row.GetOrganization())

	return ws
}
