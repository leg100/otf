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
	db.AutoMigrate(&ots.Run{}, &ots.Apply{}, &ots.Plan{})

	return &RunDB{
		DB: db,
	}
}

// CreateRun persists a Run to the DB.
func (db RunDB) Create(run *ots.Run) (*ots.Run, error) {
	if result := db.Omit("Workspace", "ConfigurationVersion").Create(run); result.Error != nil {
		return nil, result.Error
	}

	return run, nil
}

// UpdateRun persists an updated Run to the DB. The existing run is fetched from
// the DB, the supplied func is invoked on the run, and the updated run is
// persisted back to the DB. The returned Run includes any changes, including a
// new UpdatedAt value.
func (db RunDB) Update(id string, fn func(*ots.Run) error) (*ots.Run, error) {
	var run *ots.Run

	err := db.Transaction(func(tx *gorm.DB) (err error) {
		// Get existing model obj from DB
		run, err = getRun(tx, ots.RunGetOptions{ID: &id})
		if err != nil {
			return err
		}

		// Update obj using client-supplied fn
		if err := fn(run); err != nil {
			return err
		}

		if result := tx.Session(&gorm.Session{FullSaveAssociations: true}).Save(run); result.Error != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return run, nil
}

func (db RunDB) List(opts ots.RunListOptions) (*ots.RunList, error) {
	var models []ots.Run
	var count int64

	err := db.Transaction(func(tx *gorm.DB) error {
		query := tx

		// Optionally filter by workspace
		if opts.WorkspaceID != nil {
			ws, err := getWorkspace(tx, ots.WorkspaceSpecifier{ID: opts.WorkspaceID})
			if err != nil {
				return err
			}

			query = query.Where("workspace_id = ?", ws.InternalID)
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
		Items:      runListToPointerList(models),
		Pagination: ots.NewPagination(opts.ListOptions, int(count)),
	}, nil
}

func (db RunDB) Get(opts ots.RunGetOptions) (*ots.Run, error) {
	run, err := getRun(db.DB, opts)
	if err != nil {
		return nil, err
	}
	return run, nil
}

func getRun(db *gorm.DB, opts ots.RunGetOptions) (*ots.Run, error) {
	var model ots.Run

	query := db.Preload(clause.Associations)

	switch {
	case opts.ID != nil:
		query = query.Where("external_id = ?", opts.ID)
	case opts.ApplyID != nil:
		query = query.Joins("JOIN applies ON applies.id = runs.apply_id").Where("applies.external_id = ?", opts.ApplyID)
	default:
		return nil, ots.ErrInvalidRunGetOptions
	}

	if result := query.Find(&model); result.Error != nil {
		return nil, result.Error
	}

	return &model, nil
}

func runListToPointerList(rl []ots.Run) (pl []*ots.Run) {
	for i := range rl {
		pl = append(pl, &rl[i])
	}
	return
}
