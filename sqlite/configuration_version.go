package sqlite

import (
	"github.com/leg100/otf"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var _ otf.ConfigurationVersionStore = (*ConfigurationVersionDB)(nil)

type ConfigurationVersionDB struct {
	*gorm.DB
}

func NewConfigurationVersionDB(db *gorm.DB) *ConfigurationVersionDB {
	return &ConfigurationVersionDB{
		DB: db,
	}
}

func (db ConfigurationVersionDB) Create(cv *otf.ConfigurationVersion) (*otf.ConfigurationVersion, error) {
	model := &ConfigurationVersion{}
	model.FromDomain(cv)

	if result := db.Omit("Workspace").Create(model); result.Error != nil {
		return nil, result.Error
	}

	return model.ToDomain(), nil
}

// Update persists an updated ConfigurationVersion to the DB. The existing run
// is fetched from the DB, the supplied func is invoked on the run, and the
// updated run is persisted back to the DB. The returned ConfigurationVersion
// includes any changes, including a new UpdatedAt value.
func (db ConfigurationVersionDB) Update(id string, fn func(*otf.ConfigurationVersion) error) (*otf.ConfigurationVersion, error) {
	var model *ConfigurationVersion

	err := db.Transaction(func(tx *gorm.DB) (err error) {
		// Get existing model obj from DB
		model, err = getConfigurationVersion(tx, otf.ConfigurationVersionGetOptions{ID: &id})
		if err != nil {
			return err
		}

		// Update domain obj using client-supplied fn
		if err := model.Update(fn); err != nil {
			return err
		}

		if result := tx.Save(&model); result.Error != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return model.ToDomain(), nil
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

	return &otf.ConfigurationVersionList{
		Items:      models.ToDomain(),
		Pagination: otf.NewPagination(opts.ListOptions, int(count)),
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
