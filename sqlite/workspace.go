package sqlite

import (
	"fmt"
	"strings"
	"time"

	"github.com/jinzhu/copier"
	"github.com/jmoiron/sqlx"
	"github.com/leg100/otf"
)

var (
	_ otf.WorkspaceStore = (*WorkspaceDB)(nil)

	insertWorkspaceSql = `INSERT INTO workspaces (
    created_at,
    updated_at,
    external_id,
    allow_destroy_plan,
    auto_apply,
    can_queue_destroy_plan,
    description,
    environment,
    execution_mode,
    file_triggers_enabled,
    global_remote_state,
    locked,
    migration_environment,
    name,
    queue_all_runs,
    speculative_enabled,
    source_name,
    source_url,
    terraform_version,
    trigger_prefixes,
    working_directory,
    organization_id,
VALUES (
	:created_at,
    :updated_at,
    :external_id,
    :allow_destroy_plan,
    :auto_apply,
    :can_queue_destroy_plan,
    :description,
    :environment,
    :execution_mode,
    :file_triggers_enabled,
    :global_remote_state,
    :locked,
    :migration_environment,
    :name,
    :queue_all_runs,
    :speculative_enabled,
    :source_name,
    :source_url,
    :terraform_version,
    :trigger_prefixes,
    :working_directory,
    :organization_id,
`

	getWorkspaceColumns = `
workspaces.id               		AS workspaces.id
workspaces.created_at               AS workspaces.created_at
workspaces.updated_at               AS workspaces.updated_at
workspaces.external_id              AS workspaces.external_id
workspaces.allow_destroy_plan       AS workspaces.allow_destroy_plan
workspaces.auto_apply               AS workspaces.auto_apply
workspaces.can_queue_destroy_plan   AS workspaces.can_queue_destroy_plan
workspaces.description              AS workspaces.description
workspaces.environment              AS workspaces.environment
workspaces.execution_mode           AS workspaces.execution_mode
workspaces.file_triggers_enabled    AS workspaces.file_triggers_enabled
workspaces.global_remote_state      AS workspaces.global_remote_state
workspaces.locked                   AS workspaces.locked
workspaces.migration_environment    AS workspaces.migration_environment
workspaces.name                     AS workspaces.name
workspaces.queue_all_runs           AS workspaces.queue_all_runs
workspaces.speculative_enabled      AS workspaces.speculative_enabled
workspaces.source_name              AS workspaces.source_name
workspaces.source_url               AS workspaces.source_url
workspaces.terraform_version        AS workspaces.terraform_version
workspaces.trigger_prefixes         AS workspaces.trigger_prefixes
workspaces.working_directory        AS workspaces.working_directory
workspaces.organization_id          AS workspaces.organization_id
`

	getWorkspaceJoins = `JOIN organizations ON organizations.id = workspaces.organization_id`
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
	tx := db.MustBegin()
	defer tx.Rollback()

	// Insert workspace
	result, err := tx.NamedExec(insertWorkspaceSql, ws)
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

	before := otf.Workspace{}
	copier.Copy(&before, ws)

	// Update obj using client-supplied fn
	if err := fn(ws); err != nil {
		return nil, err
	}

	updates := FindUpdates(db.Mapper, before, ws)
	if len(updates) == 0 {
		return ws, nil
	}

	ws.UpdatedAt = time.Now()
	updates["updated_at"] = ws.UpdatedAt

	var sql strings.Builder
	fmt.Fprintln(&sql, "UPDATE workspaces")

	for k := range updates {
		fmt.Fprintln(&sql, "SET %s = :%[1]s", k)
	}

	fmt.Fprintf(&sql, "WHERE %s = :id", ws.Model.ID)

	_, err = db.NamedExec(sql.String(), updates)
	if err != nil {
		return nil, err
	}

	return ws, tx.Commit()
}

func (db WorkspaceDB) List(opts otf.WorkspaceListOptions) (*otf.WorkspaceList, error) {
	type listParams struct {
		Limit            int
		Offset           int
		Prefix           string
		OrganizationName string
	}

	params := listParams{}

	var sql strings.Builder
	fmt.Fprintln(&sql, "SELECT", getWorkspaceColumns, "FROM", "workspaces", getWorkspaceJoins)

	var conditions []string

	// Optionally filter by workspace name prefix
	if opts.Prefix != nil {
		conditions = append(conditions, "name LIKE ?")
		params.Prefix = fmt.Sprintf("%s%%", *opts.Prefix)
	}

	// Optionally filter by organization name
	if opts.OrganizationName != nil {
		conditions = append(conditions, "organizations.name = ?")
		params.OrganizationName = *opts.OrganizationName
	}

	fmt.Fprintln(&sql, "WHERE", strings.Join(conditions, " AND "))

	if opts.PageSize > 0 {
		params.Limit = opts.PageSize
	}

	if opts.PageNumber > 0 {
		params.Offset = (opts.PageNumber - 1) * opts.PageSize
	}

	var result []otf.Workspace
	if err := db.Select(&result, sql.String(), params); err != nil {
		return nil, err
	}

	// Convert from []otf.Workspace to []*otf.Workspace
	var items []*otf.Workspace
	for _, r := range result {
		items = append(items, &r)
	}

	return &otf.WorkspaceList{
		Items:      items,
		Pagination: otf.NewPagination(opts.ListOptions, len(items)),
	}, nil
}

func (db WorkspaceDB) Get(spec otf.WorkspaceSpecifier) (*otf.Workspace, error) {
	return getWorkspace(db.MustBegin(), spec)
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
	_, err = db.MustExec("DELETE FROM workspaces WHERE external_id = ?", spec.ID)
	if err != nil {
		return err
	}

	// Delete associated runs
	_, err = db.MustExec("DELETE FROM runs WHERE workspace_id = ?", ws.Model.ID)
	if err != nil {
		return err
	}

	// Delete associated state versions
	_, err = db.MustExec("DELETE FROM state_versions WHERE workspace_id = ?", ws.Model.ID)
	if err != nil {
		return err
	}

	// Delete associated configuration versions
	_, err = db.MustExec("DELETE FROM configuration_versions WHERE workspace_id = ?", ws.Model.ID)
	if err != nil {
		return err
	}

	if err != nil {
		return err
	}

	return tx.Commit()
}

func getWorkspace(db Getter, spec otf.WorkspaceSpecifier) (*otf.Workspace, error) {
	type getWorkspaceParams struct {
		ID               string
		InternalID       int64
		Name             string
		OrganizationName string
	}
	params := getWorkspaceParams{}

	var sql strings.Builder
	fmt.Fprintln(&sql, "SELECT", getWorkspaceColumns, "FROM", "workspaces", getWorkspaceJoins)

	var conditions []string

	switch {
	case spec.ID != nil:
		// Get workspace by (external) ID
		conditions = append(conditions, "workspaces.external_id = :id")
		params.ID = *spec.ID
	case spec.InternalID != nil:
		// Get workspace by internal ID
		conditions = append(conditions, "workspaces.id = :internal_id")
		params.InternalID = *spec.InternalID
	case spec.Name != nil && spec.OrganizationName != nil:
		// Get workspace by name and organization name
		conditions = append(conditions, "workspaces.name = :name")
		conditions = append(conditions, "organizations.name = :organization_name")
		params.Name = *spec.Name
		params.OrganizationName = *spec.OrganizationName
	default:
		return nil, otf.ErrInvalidWorkspaceSpecifier
	}

	fmt.Fprintln(&sql, "WHERE", strings.Join(conditions, " AND "))

	var ws otf.Workspace
	if err := db.Get(&ws, sql.String(), params); err != nil {
		return nil, err
	}

	return &ws, nil
}
