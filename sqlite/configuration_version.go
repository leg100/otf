package sqlite

import (
	"github.com/leg100/ots"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var _ ots.ConfigurationVersionRepository = (*ConfigurationVersionDB)(nil)

type ConfigurationVersionDB struct {
	*gorm.DB
}

func NewConfigurationVersionDB(db *gorm.DB) *ConfigurationVersionDB {
	db.AutoMigrate(&ots.ConfigurationVersion{})

	return &ConfigurationVersionDB{
		DB: db,
	}
}

func (db ConfigurationVersionDB) Create(cv *ots.ConfigurationVersion) (*ots.ConfigurationVersion, error) {
	if result := db.DB.Create(cv); result.Error != nil {
		return nil, result.Error
	}

	return cv, nil
}

// Update persists an updated ConfigurationVersion to the DB. The existing run
// is fetched from the DB, the supplied func is invoked on the run, and the
// updated run is persisted back to the DB. The returned ConfigurationVersion
// includes any changes, including a new UpdatedAt value.
func (db ConfigurationVersionDB) Update(id string, fn func(*ots.ConfigurationVersion) error) (*ots.ConfigurationVersion, error) {
	var cv *ots.ConfigurationVersion

	err := db.Transaction(func(tx *gorm.DB) (err error) {
		// Get existing model obj from DB
		cv, err = getConfigurationVersion(tx, ots.ConfigurationVersionGetOptions{ID: &id})
		if err != nil {
			return err
		}

		// Update domain obj using client-supplied fn
		if err := fn(cv); err != nil {
			return err
		}

		if result := tx.Session(&gorm.Session{FullSaveAssociations: true}).Save(&cv); result.Error != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return cv, nil
}

func (db ConfigurationVersionDB) List(workspaceID string, opts ots.ConfigurationVersionListOptions) (*ots.ConfigurationVersionList, error) {
	var models []ots.ConfigurationVersion
	var count int64

	err := db.Transaction(func(tx *gorm.DB) error {
		ws, err := getWorkspace(tx, ots.WorkspaceSpecifier{ID: &workspaceID})
		if err != nil {
			return err
		}

		query := tx.Where("workspace_id = ?", ws.InternalID)

		if result := query.Model(&models).Count(&count); result.Error != nil {
			return result.Error
		}

		if result := query.Preload(clause.Associations).Scopes(paginate(opts.ListOptions)).Find(&models); result.Error != nil {
			return result.Error
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &ots.ConfigurationVersionList{
		Items:      configurationVersionListToPointerList(models),
		Pagination: ots.NewPagination(opts.ListOptions, int(count)),
	}, nil
}

func (db ConfigurationVersionDB) Get(opts ots.ConfigurationVersionGetOptions) (*ots.ConfigurationVersion, error) {
	cv, err := getConfigurationVersion(db.DB, opts)
	if err != nil {
		return nil, err
	}
	return cv, nil
}

func getConfigurationVersion(db *gorm.DB, opts ots.ConfigurationVersionGetOptions) (*ots.ConfigurationVersion, error) {
	var model ots.ConfigurationVersion
	var query *gorm.DB

	switch {
	case opts.ID != nil:
		// Get config version by ID
		query = db.Where("external_id = ?", *opts.ID)
	case opts.WorkspaceID != nil:
		// Get latest config version for given workspace
		ws, err := getWorkspace(db, ots.WorkspaceSpecifier{ID: opts.WorkspaceID})
		if err != nil {
			return nil, err
		}
		query = db.Where("workspace_id = ?", ws.InternalID).Order("created_at desc")
	default:
		return nil, ots.ErrInvalidConfigurationVersionGetOptions
	}

	if result := query.Preload(clause.Associations).First(&model); result.Error != nil {
		return nil, result.Error
	}

	return &model, nil
}

func configurationVersionListToPointerList(cvl []ots.ConfigurationVersion) (pl []*ots.ConfigurationVersion) {
	for i := range cvl {
		pl = append(pl, &cvl[i])
	}
	return
}
