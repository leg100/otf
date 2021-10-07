package sqlite

import (
	"github.com/leg100/otf"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	_ otf.ConfigurationVersionStore = (*ConfigurationVersionDB)(nil)

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
	:created_at            AS :created_at
    :updated_at            AS :updated_at
    :external_id           AS :external_id
	:auto_queue_runs       AS :auto_queue_runs
    :source                AS :source
    :speculative           AS :speculative
    :status                AS :status
    :status_timestamps     AS :status_timestamps
    :blob_id               AS :blob_id
    :workspace_id          AS :workspace_id)
`
)

type ConfigurationVersionDB struct {
	*gorm.DB
}

func NewConfigurationVersionDB(db *gorm.DB) *ConfigurationVersionDB {
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
	fmt.Fprintln(&sql, "UPDATE workspaces")

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
	var models ConfigurationVersionList
	var count int64

	err := db.Transaction(func(tx *gorm.DB) error {
		ws, err := getWorkspace(tx, otf.WorkspaceSpecifier{ID: &workspaceID})
		if err != nil {
			return err
		}

		query := tx.Where("workspace_id = ?", ws.ID)

	type listParams struct {
		Limit            int
		Offset           int
		WorkspaceID string
	}

	params := listParams{}

	var sql strings.Builder
	fmt.Fprintln(&sql, "SELECT", getConfigurationVersionColumns, "FROM", "configuration_versions", getWorkspaceJoins)

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

func (db ConfigurationVersionDB) Get(opts otf.ConfigurationVersionGetOptions) (*otf.ConfigurationVersion, error) {
	cv, err := getConfigurationVersion(db.DB, opts)
	if err != nil {
		return nil, err
	}
	return cv.ToDomain(), nil
}

func getConfigurationVersion(db *gorm.DB, opts otf.ConfigurationVersionGetOptions) (*ConfigurationVersion, error) {
	var model ConfigurationVersion

	query := db.Preload(clause.Associations)

	switch {
	case opts.ID != nil:
		// Get config version by ID
		query = query.Where("external_id = ?", *opts.ID)
	case opts.WorkspaceID != nil:
		// Get latest config version for given workspace
		ws, err := getWorkspace(db, otf.WorkspaceSpecifier{ID: opts.WorkspaceID})
		if err != nil {
			return nil, err
		}
		query = query.Where("workspace_id = ?", ws.ID).Order("created_at desc")
	default:
		return nil, otf.ErrInvalidConfigurationVersionGetOptions
	}

	if result := query.First(&model); result.Error != nil {
		return nil, result.Error
	}

	return &model, nil
}
