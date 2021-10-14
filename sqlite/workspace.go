package sqlite

import (
	"fmt"
	"strings"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"github.com/leg100/otf"
	"github.com/mitchellh/copystructure"
)

var (
	_ otf.WorkspaceStore = (*WorkspaceDB)(nil)

	workspaceColumnsWithoutID = []string{"created_at", "updated_at", "external_id", "allow_destroy_plan", "auto_apply", "can_queue_destroy_plan", "description", "environment", "execution_mode", "file_triggers_enabled", "global_remote_state", "locked", "migration_environment", "name", "queue_all_runs", "speculative_enabled", "source_name", "source_url", "terraform_version", "trigger_prefixes", "working_directory"}
	workspaceColumns          = append(workspaceColumnsWithoutID, "id")

	insertWorkspaceSQL = fmt.Sprintf("INSERT INTO workspaces (%s, organization_id) VALUES (%s, :organizations.id)",
		strings.Join(workspaceColumnsWithoutID, ", "),
		strings.Join(otf.PrefixSlice(workspaceColumnsWithoutID, ":"), ", "))
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
	result, err := db.NamedExec(insertWorkspaceSQL, ws)
	if err != nil {
		return nil, err
	}
	ws.Model.ID, err = result.LastInsertId()
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

	updated, err := update(db.Mapper, tx, "workspaces", before.(*otf.Workspace), ws)
	if err != nil {
		return nil, err
	}

	if updated {
		return ws, tx.Commit()
	}

	return ws, nil
}

func (db WorkspaceDB) List(opts otf.WorkspaceListOptions) (*otf.WorkspaceList, error) {
	selectBuilder := sq.Select(asColumnList("workspaces", false, workspaceColumns...)).
		Columns(asColumnList("organizations", true, organizationColumns...)).
		From("workspaces").
		Join("organizations ON organizations.id = workspaces.organization_id").
		Limit(opts.GetLimit()).
		Offset(opts.GetOffset())

	// Optionally filter by workspace name prefix
	if opts.Prefix != nil {
		selectBuilder = selectBuilder.Where("workspaces.name LIKE ?", fmt.Sprintf("%s%%", *opts.Prefix))
	}

	// Optionally filter by organization name
	if opts.OrganizationName != nil {
		selectBuilder = selectBuilder.Where("organizations.name = ?", *opts.OrganizationName)
	}

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
		Pagination: otf.NewPagination(opts.ListOptions, len(items)),
	}, nil
}

func (db WorkspaceDB) Get(spec otf.WorkspaceSpecifier) (*otf.Workspace, error) {
	return getWorkspace(db.DB, spec)
}

// Delete deletes a specific workspace, along with its associated records (runs
// etc).
func (db WorkspaceDB) Delete(spec otf.WorkspaceSpecifier) error {
	tx := db.MustBegin()
	defer tx.Rollback()

	ws, err := getWorkspace(tx, spec)
	if err != nil {
		return err
	}

	// Delete workspace
	result, err := tx.Exec("DELETE FROM workspaces WHERE external_id = ?", ws.ID)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return fmt.Errorf("no workspace found")
	}

	// Delete associated runs
	_, err = tx.Exec("DELETE FROM runs WHERE workspace_id = ?", ws.Model.ID)
	if err != nil {
		return err
	}

	// Delete associated state versions
	_, err = tx.Exec("DELETE FROM state_versions WHERE workspace_id = ?", ws.Model.ID)
	if err != nil {
		return err
	}

	// Delete associated configuration versions
	_, err = tx.Exec("DELETE FROM configuration_versions WHERE workspace_id = ?", ws.Model.ID)
	if err != nil {
		return err
	}

	if err != nil {
		return err
	}

	return tx.Commit()
}

func getWorkspace(db Getter, spec otf.WorkspaceSpecifier) (*otf.Workspace, error) {
	selectBuilder := sq.Select(asColumnList("workspaces", false, workspaceColumns...)).
		Columns(asColumnList("organizations", true, organizationColumns...)).
		From("workspaces").
		Join("organizations ON organizations.id = workspaces.organization_id")

	switch {
	case spec.ID != nil:
		// Get workspace by (external) ID
		selectBuilder = selectBuilder.Where("workspaces.external_id = ?", *spec.ID)
	case spec.InternalID != nil:
		// Get workspace by internal ID
		selectBuilder = selectBuilder.Where("workspaces.id = ?", *spec.InternalID)
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
		return nil, err
	}

	return &ws, nil
}
