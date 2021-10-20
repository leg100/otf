package sql

import (
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/leg100/otf"
	"github.com/mitchellh/copystructure"
)

var (
	_ otf.WorkspaceStore = (*WorkspaceDB)(nil)

	workspaceColumns = []string{
		"workspace_id",
		"created_at",
		"updated_at",
		"allow_destroy_plan",
		"auto_apply",
		"can_queue_destroy_plan",
		"description",
		"environment",
		"execution_mode",
		"file_triggers_enabled",
		"global_remote_state",
		"locked",
		"migration_environment",
		"name",
		"queue_all_runs",
		"speculative_enabled",
		"source_name",
		"source_url",
		"structured_run_output_enabled",
		"terraform_version",
		"trigger_prefixes",
		"working_directory",
	}

	insertWorkspaceSQL = fmt.Sprintf("INSERT INTO workspaces (%s, organization_id) VALUES (%s, :organizations.organization_id)",
		strings.Join(workspaceColumns, ", "),
		strings.Join(otf.PrefixSlice(workspaceColumns, ":"), ", "))
)

type WorkspaceDB struct {
	*sqlx.DB
}

func NewWorkspaceDB(db *sqlx.DB) *WorkspaceDB {
	return &WorkspaceDB{
		DB: db,
	}
}

// Create persists a Workspace to the DB. The returned Workspace is adorned with
// additional metadata, i.e. CreatedAt, UpdatedAt, etc.
func (db WorkspaceDB) Create(ws *otf.Workspace) (*otf.Workspace, error) {
	// Insert workspace
	sql, args, err := db.BindNamed(insertWorkspaceSQL, ws)
	if err != nil {
		return nil, err
	}
	_, err = db.Exec(sql, args...)
	if err != nil {
		return nil, err
	}

	return ws, nil
}

// Update persists an updated Workspace to the DB. The existing run is fetched
// from the DB, the supplied func is invoked on the run, and the updated run is
// persisted back to the DB. The returned Workspace includes any changes,
// including a new UpdatedAt value.
func (db WorkspaceDB) Update(spec otf.WorkspaceSpecifier, fn func(*otf.Workspace) error) (*otf.Workspace, error) {
	tx := db.MustBegin()
	defer tx.Rollback()

	ws, err := getWorkspace(tx, spec)
	if err != nil {
		return nil, err
	}

	// Make a copy for comparison with the updated obj
	before, err := copystructure.Copy(ws)
	if err != nil {
		return nil, err
	}

	// Update obj using client-supplied fn
	if err := fn(ws); err != nil {
		return nil, err
	}

	updated, err := update(db.Mapper, tx, "workspaces", "workspace_id", before.(*otf.Workspace), ws)
	if err != nil {
		return nil, err
	}

	if updated {
		return ws, tx.Commit()
	}

	return ws, nil
}

func (db WorkspaceDB) List(opts otf.WorkspaceListOptions) (*otf.WorkspaceList, error) {
	selectBuilder := psql.Select().
		From("workspaces").
		Join("organizations USING (organization_id)")

	// Optionally filter by workspace name prefix
	if opts.Prefix != nil {
		selectBuilder = selectBuilder.Where("workspaces.name LIKE ?", fmt.Sprintf("%s%%", *opts.Prefix))
	}

	// Optionally filter by organization name
	if opts.OrganizationName != nil {
		selectBuilder = selectBuilder.Where("organizations.name = ?", *opts.OrganizationName)
	}

	var count int
	if err := selectBuilder.Columns("count(1)").RunWith(db).QueryRow().Scan(&count); err != nil {
		return nil, fmt.Errorf("counting total rows: %w", err)
	}

	selectBuilder = selectBuilder.
		Columns(asColumnList("workspaces", false, workspaceColumns...)).
		Columns(asColumnList("organizations", true, organizationColumns...)).
		Limit(opts.GetLimit()).
		Offset(opts.GetOffset())

	sql, args, err := selectBuilder.ToSql()
	if err != nil {
		return nil, err
	}

	var items []*otf.Workspace
	if err := db.Select(&items, sql, args...); err != nil {
		return nil, fmt.Errorf("unable to scan workspaces from db: %w", err)
	}

	return &otf.WorkspaceList{
		Items:      items,
		Pagination: otf.NewPagination(opts.ListOptions, count),
	}, nil
}

func (db WorkspaceDB) Get(spec otf.WorkspaceSpecifier) (*otf.Workspace, error) {
	return getWorkspace(db.DB, spec)
}

// Delete deletes a specific workspace, along with its child records (runs etc).
func (db WorkspaceDB) Delete(spec otf.WorkspaceSpecifier) error {
	tx := db.MustBegin()
	defer tx.Rollback()

	ws, err := getWorkspace(tx, spec)
	if err != nil {
		return err
	}

	_, err = tx.Exec("DELETE FROM workspaces WHERE workspace_id = $1", ws.ID)
	if err != nil {
		return fmt.Errorf("unable to delete workspace: %w", err)
	}

	return tx.Commit()
}

func getWorkspace(db Getter, spec otf.WorkspaceSpecifier) (*otf.Workspace, error) {
	selectBuilder := psql.Select(asColumnList("workspaces", false, workspaceColumns...)).
		Columns(asColumnList("organizations", true, organizationColumns...)).
		From("workspaces").
		Join("organizations USING (organization_id)")

	switch {
	case spec.ID != nil:
		// Get workspace by ID
		selectBuilder = selectBuilder.Where("workspace_id = ?", *spec.ID)
	case spec.Name != nil && spec.OrganizationName != nil:
		// Get workspace by name and organization name
		selectBuilder = selectBuilder.Where("workspaces.name = ?", *spec.Name)
		selectBuilder = selectBuilder.Where("organizations.name = ?", *spec.OrganizationName)
	default:
		return nil, otf.ErrInvalidWorkspaceSpecifier
	}

	sql, args, err := selectBuilder.ToSql()
	if err != nil {
		return nil, err
	}

	var ws otf.Workspace
	if err := db.Get(&ws, sql, args...); err != nil {
		return nil, databaseError(err)
	}

	return &ws, nil
}
