package sqlite

import (
	"fmt"

	"github.com/leg100/otf"
	"sqlx.io/sqlx"
	"sqlx.io/sqlx/clause"
)

var (
	_ otf.StateVersionStore = (*StateVersionService)(nil)

	stateVersionTableName = "state_versions"

	insertStateVersionSql = `INSERT INTO state_versions (
    created_at,
    updated_at,
    external_id,
    serial,
    blob_id,
    workspace_id)
VALUES (
	:created_at,
    :updated_at,
    :external_id,
    :serial,
    :blob_id,
    :workspace_id)
`

	getStateVersionColumns = `
state_versions.created_at   AS state_versions.created_at,
state_versions.updated_at   AS state_versions.updated_at,
state_versions.external_id  AS state_versions.external_id,
state_versions.serial       AS state_versions.serial,
state_versions.blob_id      AS state_versions.blob_id,
state_versions.workspace_id AS state_versions.workspace_id,
`
	stateVersionColumns = []string{"created_at", "updated_at", "external_id", "serial", "blob_id", "workspace_id"}

	listStateVersionsSql = fmt.Sprintf(`SELECT %s, %s
FROM state_versions
JOIN workspaces ON workspaces.id = state_versions.workspace_id
JOIN organizations ON organizations.id = workspaces.organization_id
WHERE workspaces.external_id = :workspace_external_id
AND workspaces.name = :workspace_name
AND organizations.name = :organization_name
LIMIT :limit
OFFSET :offset
`, asColumnList(stateVersionTableName, getStateVersionColumns), asColumnList(workspacesTableName, workspaceColumns))
)

type StateVersionService struct {
	*sqlx.DB
	columns []string
}

func NewStateVersionDB(db *sqlx.DB) *StateVersionService {
	return &StateVersionService{
		DB:      db,
		columns: []string{"created_at", "updated_at", "external_id", "serial", "blob_id", "workspace_id"},
	}
}

// Create persists a StateVersion to the DB.
func (s StateVersionService) Create(sv *otf.StateVersion) (*otf.StateVersion, error) {
	tx := db.MustBegin()
	defer tx.Rollback()

	// Insert
	result, err := tx.NamedExec(insertStateVersionSql, sv)
	if err != nil {
		return nil, err
	}
	sv.Model.ID, err = result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return sv, nil
}

func (s StateVersionService) List(opts otf.StateVersionListOptions) (*otf.StateVersionList, error) {
	limit, offset := opts.GetSQLWindow()

	params := map[string]interface{}{
		"workspace_name":    opts.Workspace,
		"organization_name": opts.Organization,
		"limit":             limit,
		"offset":            offset,
	}

	var result []otf.StateVersion
	if err := s.Select(&result, listStateVersionsSql, params); err != nil {
		return nil, err
	}

	// Convert from []otf.StateVersion to []*otf.StateVersion
	var items []*otf.StateVersion
	for _, r := range result {
		items = append(items, &r)
	}

	return &otf.StateVersionList{
		Items:      items,
		Pagination: otf.NewPagination(opts.ListOptions, len(items)),
	}, nil
}

func (s StateVersionService) Get(opts otf.StateVersionGetOptions) (*otf.StateVersion, error) {
	sv, err := getStateVersion(s.DB, opts)
	if err != nil {
		return nil, err
	}
	return sv.ToDomain(), nil
}

func getStateVersion(db *sqlx.DB, opts otf.StateVersionGetOptions) (*StateVersion, error) {
	var model StateVersion

	query := db.Preload(clause.Associations)

	switch {
	case opts.ID != nil:
		// Get state version by ID
		query = query.Where("external_id = ?", *opts.ID)
	case opts.WorkspaceID != nil:
		// Get most recent state version belonging to workspace
		query = query.Joins("JOIN workspaces ON workspaces.id = state_versions.workspace_id").
			Order("state_versions.serial desc, state_versions.created_at desc").
			Where("workspaces.external_id = ?", *opts.WorkspaceID)
	default:
		return nil, otf.ErrInvalidStateVersionGetOptions
	}

	if result := query.First(&model); result.Error != nil {
		return nil, result.Error
	}

	return &model, nil
}
