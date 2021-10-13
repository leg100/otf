package sqlite

import (
	"fmt"
	"strings"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"github.com/leg100/otf"
	"github.com/mitchellh/copystructure"
)

var (
	_ otf.ConfigurationVersionStore = (*ConfigurationVersionDB)(nil)

	configurationVersionColumnsWithoutID = []string{"created_at", "updated_at", "external_id", "auto_queue_runs", "source", "speculative", "status", "status_timestamps", "blob_id"}

	configurationVersionColumns = append(configurationVersionColumnsWithoutID, "id")

	insertConfigurationVersionSQL = fmt.Sprintf("INSERT INTO configuration_versions (%s, workspace_id) VALUES (%s, :workspaces.id)",
		strings.Join(configurationVersionColumnsWithoutID, ", "),
		strings.Join(otf.PrefixSlice(configurationVersionColumnsWithoutID, ":"), ", "))
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
	// Insert
	result, err := db.NamedExec(insertConfigurationVersionSQL, cv)
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

	// Make a copy for comparison with the updated obj
	before, err := copystructure.Copy(cv)
	if err != nil {
		return nil, err
	}

	// Update obj using client-supplied fn
	if err := fn(cv); err != nil {
		return nil, err
	}

	updates := FindUpdates(db.Mapper, before.(*otf.ConfigurationVersion), cv)
	if len(updates) == 0 {
		return cv, nil
	}

	cv.UpdatedAt = time.Now()
	updates["updated_at"] = cv.UpdatedAt

	sql := sq.Update("configuration_versions").Where("id = ?", cv.Model.ID)

	query, args, err := sql.SetMap(updates).ToSql()
	if err != nil {
		return nil, err
	}

	_, err = tx.Exec(query, args...)
	if err != nil {
		return nil, fmt.Errorf("executing SQL statement: %s: %w", query, err)
	}

	return cv, tx.Commit()
}

func (db ConfigurationVersionDB) List(workspaceID string, opts otf.ConfigurationVersionListOptions) (*otf.ConfigurationVersionList, error) {
	selectBuilder := sq.Select(asColumnList("configuration_versions", false, configurationVersionColumns...)).
		Columns(asColumnList("workspaces", true, workspaceColumns...)).
		From("configuration_versions").
		Join("workspaces ON workspaces.id == configuration_versions.workspace_id").
		Where("workspaces.external_id = ?", workspaceID).
		Limit(opts.GetLimit()).
		Offset(opts.GetOffset())

	sql, args, err := selectBuilder.ToSql()
	if err != nil {
		return nil, err
	}

	var result []otf.ConfigurationVersion
	if err := db.Select(&result, sql, args...); err != nil {
		return nil, err
	}

	// Convert from []otf.ConfigurationVersion to []*otf.ConfigurationVersion
	var items []*otf.ConfigurationVersion
	for i := 0; i < len(result); i++ {
		items = append(items, &result[i])
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
	selectBuilder := sq.Select(asColumnList("configuration_versions", false, configurationVersionColumns...)).
		Columns(asColumnList("workspaces", true, workspaceColumns...)).
		Join("workspaces ON workspaces.id == configuration_versions.workspace_id").
		From("configuration_versions")

	switch {
	case opts.ID != nil:
		// Get config version by ID
		selectBuilder = selectBuilder.Where("configuration_versions.external_id = ?", *opts.ID)
	case opts.WorkspaceID != nil:
		// Get latest config version for given workspace
		selectBuilder = selectBuilder.Where("workspaces.external_id = ?", *opts.WorkspaceID)
	default:
		return nil, otf.ErrInvalidWorkspaceSpecifier
	}

	sql, args, err := selectBuilder.ToSql()
	if err != nil {
		return nil, err
	}

	var cv otf.ConfigurationVersion
	if err := getter.Get(&cv, sql, args...); err != nil {
		return nil, err
	}

	return &cv, nil
}
