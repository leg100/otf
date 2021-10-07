package sqlite

import (
	"fmt"
	"strings"
	"time"

	"github.com/jinzhu/copier"
	"github.com/jmoiron/sqlx"
	"github.com/leg100/otf"
	"sqlx.io/sqlx"
)

var (
	_ otf.ConfigurationVersionStore = (*ConfigurationVersionDB)(nil)

	configurationVersionsTableName = "configuration_versions"

	insertConfigurationVersionSql = `INSERT INTO configuration_versions (
    created_at,
    updated_at,
    external_id,
    auto_queue_runs,
    source,
    speculative,
    status,
    status_timestamps,
    blob_id,
    workspace_id)
VALUES (
	:created_at,
    :updated_at,
    :external_id,
	:auto_queue_runs,
    :source,
    :speculative,
    :status,
    :status_timestamps,
    :blob_id,
    :workspace_id)
`

	configurationVersionColumns = []string{"created_at", "updated_at", "external_id", "auto_queue_runs", "source", "speculative", "status", "status_timestamps", "blob_id", "workspace_id"}

	configurationVersionColumnList = asColumnList(configurationVersionsTableName, configurationVersionColumns...)

	listConfigurationVersionsSql = fmt.Sprintf(`SELECT %s, %s
FROM configuration_versions
JOIN workspaces ON workspaces.id = configuration_versions.workspace_id
WHERE workspaces.external_id = ?`, configurationVersionColumnList, workspacesColumnList)

	getConfigurationVersionSql = fmt.Sprintf(`SELECT %s, %s
FROM configuration_versions")
JOIN workspaces ON workspaces.id = configuration_versions.workspace_id
`, configurationVersionColumnList, workspacesColumnList)
)

type ConfigurationVersionDB struct {
	*sqlx.DB
}

func NewConfigurationVersionDB(db *sqlx.DB) *ConfigurationVersionDB {
	return &ConfigurationVersionDB{
		DB: db,
	}
}

func (db ConfigurationVersionDB) Create(cv *otf.ConfigurationVersion) (*otf.ConfigurationVersion, error) {
	tx := db.MustBegin()
	defer tx.Rollback()

	// Insert
	result, err := tx.NamedExec(insertConfigurationVersionSql, cv)
	if err != nil {
		return nil, err
	}
	cv.Model.ID, err = result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return cv, nil
}

// Update persists an updated ConfigurationVersion to the DB. The existing run
// is fetched from the DB, the supplied func is invoked on the run, and the
// updated run is persisted back to the DB. The returned ConfigurationVersion
// includes any changes, including a new UpdatedAt value.
func (db ConfigurationVersionDB) Update(id string, fn func(*otf.ConfigurationVersion) error) (*otf.ConfigurationVersion, error) {
	tx := db.MustBegin()
	defer tx.Rollback()

	cv, err := getConfigurationVersion(tx, otf.ConfigurationVersionGetOptions{ID: &id})
	if err != nil {
		return nil, err
	}

	before := otf.ConfigurationVersion{}
	copier.Copy(&before, cv)

	// Update obj using client-supplied fn
	if err := fn(cv); err != nil {
		return nil, err
	}

	updates := FindUpdates(db.Mapper, before, cv)
	if len(updates) == 0 {
		return cv, nil
	}

	cv.UpdatedAt = time.Now()
	updates["updated_at"] = cv.UpdatedAt

	var sql strings.Builder
	fmt.Fprintln(&sql, "UPDATE configuration_versions")

	for k := range updates {
		fmt.Fprintln(&sql, "SET %s = :%[1]s", k)
	}

	fmt.Fprintf(&sql, "WHERE %s = :id", cv.Model.ID)

	_, err = db.NamedExec(sql.String(), updates)
	if err != nil {
		return nil, err
	}

	return cv, tx.Commit()
}

func (db ConfigurationVersionDB) List(workspaceID string, opts otf.ConfigurationVersionListOptions) (*otf.ConfigurationVersionList, error) {
	limit, offset := opts.GetSQLWindow()

	params := map[string]interface{}{
		"limit":        limit,
		"offset":       offset,
		"workspace_id": workspaceID,
	}

	var result []otf.ConfigurationVersion
	if err := db.Select(&result, listConfigurationVersionsSql, params); err != nil {
		return nil, err
	}

	// Convert from []otf.ConfigurationVersion to []*otf.ConfigurationVersion
	var items []*otf.ConfigurationVersion
	for _, r := range result {
		items = append(items, &r)
	}

	return &otf.ConfigurationVersionList{
		Items:      items,
		Pagination: otf.NewPagination(opts.ListOptions, len(items)),
	}, nil
}

func (db ConfigurationVersionDB) Get(opts otf.ConfigurationVersionGetOptions) (*otf.ConfigurationVersion, error) {
	return getConfigurationVersion(db.DB, opts)
}

func getConfigurationVersion(getter Getter, opts otf.ConfigurationVersionGetOptions) (*otf.ConfigurationVersion, error) {
	var condition, arg string

	switch {
	case opts.ID != nil:
		// Get config version by ID
		condition = "WHERE configuration_versions.external_id = ?"
		arg = *opts.ID
	case opts.WorkspaceID != nil:
		// Get latest config version for given workspace
		condition = "WHERE workspaces.external_id = ?"
		arg = *opts.WorkspaceID
	default:
		return nil, otf.ErrInvalidWorkspaceSpecifier
	}

	sql := fmt.Sprintf(getConfigurationVersionSql, "WHERE", condition)

	var cv otf.ConfigurationVersion
	if err := getter.Get(&cv, sql, arg); err != nil {
		return nil, err
	}

	return &cv, nil
}
