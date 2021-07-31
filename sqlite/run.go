package sqlite

import (
	"github.com/leg100/ots"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var _ ots.RunStore = (*RunDB)(nil)

type RunDB struct {
	*gorm.DB
}

func NewRunDB(db *gorm.DB) *RunDB {
	return &RunDB{
		DB: db,
	}
}

// CreateRun persists a Run to the DB.
func (db RunDB) Create(domain *ots.Run) (*ots.Run, error) {
	model := &Run{}
	model.FromDomain(domain)

	if result := db.Omit("Workspace", "ConfigurationVersion").Create(model); result.Error != nil {
		return nil, result.Error
	}

	return model.ToDomain(), nil
}

// UpdateRun persists an updated Run to the DB. The existing run is fetched from
// the DB, the supplied func is invoked on the run, and the updated run is
// persisted back to the DB. The returned Run includes any changes, including a
// new UpdatedAt value.
func (db RunDB) Update(id string, fn func(*ots.Run) error) (*ots.Run, error) {
	var model *Run

	err := db.Transaction(func(tx *gorm.DB) (err error) {
		// DB -> model
		model, err := getRun(tx, ots.RunGetOptions{ID: &id})
		if err != nil {
			return err
		}

		// Update obj using client-supplied fn
		if err := model.Update(fn); err != nil {
			return err
		}

		if result := tx.Session(&gorm.Session{FullSaveAssociations: true}).Save(model); result.Error != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return model.ToDomain(), nil
}

func (db RunDB) List(opts ots.RunListOptions) (*ots.RunList, error) {
	var models RunList
	var count int64

	err := db.Transaction(func(tx *gorm.DB) error {
		query := tx

		// Optionally filter by workspace
		if opts.WorkspaceID != nil {
			ws, err := getWorkspace(tx, ots.WorkspaceSpecifier{ID: opts.WorkspaceID})
			if err != nil {
				return err
			}

			query = query.Where("workspace_id = ?", ws.Model.ID)
		}

		// Optionally filter by statuses
		if len(opts.Statuses) > 0 {
			query = query.Where("status IN ?", opts.Statuses)
		}

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

	return &ots.RunList{
		Items:      models.ToDomain(),
		Pagination: ots.NewPagination(opts.ListOptions, int(count)),
	}, nil
}

// Get retrieves a Run domain obj
func (db RunDB) Get(opts ots.RunGetOptions) (*ots.Run, error) {
	run, err := getRun(db.DB, opts)
	if err != nil {
		return nil, err
	}
	return run.ToDomain(), nil
}

func getRun(db *gorm.DB, opts ots.RunGetOptions) (*Run, error) {
	var model Run

	query := db.Preload(clause.Associations)

	switch {
	case opts.ID != nil:
		query = query.Where("external_id = ?", opts.ID)
	case opts.PlanID != nil:
		query = query.Joins("JOIN plans ON plans.run_id = runs.id").Where("plans.external_id = ?", opts.PlanID)
	case opts.ApplyID != nil:
		query = query.Joins("JOIN applies ON applies.run_id = runs.id").Where("applies.external_id = ?", opts.ApplyID)
	default:
		return nil, ots.ErrInvalidRunGetOptions
	}

	if result := query.First(&model); result.Error != nil {
		return nil, result.Error
	}

	return &model, nil
}
