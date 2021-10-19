package sql

import (
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/leg100/otf"
	"github.com/mitchellh/copystructure"
)

var (
	_ otf.ConfigurationVersionStore = (*ConfigurationVersionDB)(nil)

	configurationVersionColumns = []string{
		"configuration_version_id",
		"created_at",
		"updated_at",
		"auto_queue_runs",
		"source",
		"speculative",
		"status",
		"status_timestamps",
		"blob_id",
	}

	insertConfigurationVersionSQL = fmt.Sprintf("INSERT INTO configuration_versions (%s, workspace_id) VALUES (%s, :workspaces.workspace_id)",
		strings.Join(configurationVersionColumns, ", "),
		strings.Join(otf.PrefixSlice(configurationVersionColumns, ":"), ", "))
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
	sql, args, err := db.BindNamed(insertConfigurationVersionSQL, cv)
	if err != nil {
		return nil, err
	}
	_, err = db.Exec(sql, args...)
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

	updated, err := update(db.Mapper, tx, "configuration_versions", "configuration_version_id", before.(*otf.ConfigurationVersion), cv)
	if err != nil {
		return nil, err
	}

	if updated {
		return cv, tx.Commit()
	}

	return cv, tx.Commit()
}

func (db ConfigurationVersionDB) List(workspaceID string, opts otf.ConfigurationVersionListOptions) (*otf.ConfigurationVersionList, error) {
	selectBuilder := psql.Select().
		From("configuration_versions").
		Join("workspaces USING (workspace_id)").
		Where("workspaces.workspace_id = ?", workspaceID)

	var count int
	if err := selectBuilder.Columns("count(*)").RunWith(db).QueryRow().Scan(&count); err != nil {
		return nil, fmt.Errorf("counting total rows: %w", err)
	}

	selectBuilder = selectBuilder.
		Columns(asColumnList("configuration_versions", false, configurationVersionColumns...)).
		Columns(asColumnList("workspaces", true, workspaceColumns...)).
		Limit(opts.GetLimit()).
		Offset(opts.GetOffset())

	sql, args, err := selectBuilder.ToSql()
	if err != nil {
		return nil, err
	}

	var items []*otf.ConfigurationVersion
	if err := db.Select(&items, sql, args...); err != nil {
		return nil, err
	}

	return &otf.ConfigurationVersionList{
		Items:      items,
		Pagination: otf.NewPagination(opts.ListOptions, count),
	}, nil
}

func (db ConfigurationVersionDB) Get(opts otf.ConfigurationVersionGetOptions) (*otf.ConfigurationVersion, error) {
	return getConfigurationVersion(db.DB, opts)
}

// Delete deletes a configuration version from the DB
func (db ConfigurationVersionDB) Delete(id string) error {
	tx := db.MustBegin()
	defer tx.Rollback()

	cv, err := getConfigurationVersion(tx, otf.ConfigurationVersionGetOptions{ID: otf.String(id)})
	if err != nil {
		return err
	}

	_, err = tx.Exec("DELETE FROM configuration_versions WHERE id = $1", cv.ID)
	if err != nil {
		return fmt.Errorf("unable to delete configuration_version: %w", err)
	}

	return tx.Commit()
}

func getConfigurationVersion(getter Getter, opts otf.ConfigurationVersionGetOptions) (*otf.ConfigurationVersion, error) {
	selectBuilder := psql.Select(asColumnList("configuration_versions", false, configurationVersionColumns...)).
		Columns(asColumnList("workspaces", true, workspaceColumns...)).
		Join("workspaces USING (workspace_id)").
		From("configuration_versions")

	switch {
	case opts.ID != nil:
		// Get config version by ID
		selectBuilder = selectBuilder.Where("configuration_versions.configuration_version_id = ?", *opts.ID)
	case opts.WorkspaceID != nil:
		// Get latest config version for given workspace
		selectBuilder = selectBuilder.Where("workspaces.workspace_id = ?", *opts.WorkspaceID)
	default:
		return nil, otf.ErrInvalidWorkspaceSpecifier
	}

	sql, args, err := selectBuilder.ToSql()
	if err != nil {
		return nil, err
	}

	var cv otf.ConfigurationVersion
	if err := getter.Get(&cv, sql, args...); err != nil {
		return nil, databaseError(err)
	}

	return &cv, nil
}
