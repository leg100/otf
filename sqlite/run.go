package sqlite

import (
	"database/sql"
	"strings"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"github.com/leg100/otf"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var _ otf.RunStore = (*RunDB)(nil)

// Run models a row in a runs table.
type Run struct {
	gorm.Model

	ExternalID string

	ForceCancelAvailableAt time.Time
	IsDestroy              bool
	Message                string
	Permissions            *otf.RunPermissions `gorm:"embedded;embeddedPrefix:permission_"`
	PositionInQueue        int
	Refresh                bool
	RefreshOnly            bool
	Status                 otf.RunStatus
	StatusTimestamps       *RunStatusTimestamps `gorm:"embedded;embeddedPrefix:timestamp_"`

	// Comma separated list of replace addresses
	ReplaceAddrs string
	// Comma separated list of target addresses
	TargetAddrs string

	// Run belongs to a workspace
	WorkspaceID uint

	// Run belongs to a configuration version
	ConfigurationVersionID uint
}

type RunDB struct {
	*sqlx.DB
}

func NewRunDB(db *sqlx.DB) *RunDB {
	return &RunDB{
		DB: db,
	}
}

// Create persists a Run to the DB.
func (db RunDB) Create(run *otf.Run) (*otf.Run, error) {
	run.CreatedAt = time.Now()
	run.UpdatedAt = time.Now()

	clauses := map[string]interface{}{
		"configuration_version_id":  run.ConfigurationVersion.ID,
		"created_at":                run.CreatedAt,
		"external_id":               run.ID,
		"force_cancel_available_at": run.ForceCancelAvailableAt,
		"is_destroy":                run.IsDestroy,
		"message":                   run.Message,
		"refresh":                   run.Refresh,
		"refresh_only":              run.RefreshOnly,
		"replace_addrs":             strings.Join(run.ReplaceAddrs, ","),
		"status":                    run.Status,
		"target_addrs":              strings.Join(run.TargetAddrs, ","),
		"updated_at":                run.UpdatedAt,
		"workspace_id":              run.Workspace.ID,
	}

	if run.StatusTimestamps.AppliedAt != nil {
		clauses["timestamp_applied_at"] = sql.NullTime{Time: *run.StatusTimestamps.AppliedAt, Valid: true}
	}

	result, err := sq.Insert("runs").SetMap(clauses).RunWith(db).Exec()
	if err != nil {
		return nil, err
	}

	_, err = result.LastInsertId()
	if err != nil {
		return nil, err
	}

	// plan#create(run_id)

	// apply#create(run_id)

	return run, nil
}

// Update persists an updated Run to the DB. The existing run is fetched from
// the DB, the supplied func is invoked on the run, and the updated run is
// persisted back to the DB. The returned Run includes any changes, including a
// new UpdatedAt value.
func (db RunDB) Update(id string, fn func(*otf.Run) error) (*otf.Run, error) {
	var model *Run

	err := db.Transaction(func(tx *sqlx.DB) (err error) {
		// DB -> model
		model, err = getRun(tx, otf.RunGetOptions{ID: &id})
		if err != nil {
			return err
		}

		// Update obj using client-supplied fn
		if err := model.Update(fn); err != nil {
			return err
		}

		// Save changes to fields of relations (plan, apply) too
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

func (db RunDB) List(opts otf.RunListOptions) (*otf.RunList, error) {
	var models RunList
	var count int64

	err := db.Transaction(func(tx *sqlx.DB) error {
		query := tx

		// Optionally filter by workspace
		if opts.WorkspaceID != nil {
			ws, err := getWorkspace(tx, otf.WorkspaceSpecifier{ID: opts.WorkspaceID})
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

	return &otf.RunList{
		Items:      models.ToDomain(),
		Pagination: otf.NewPagination(opts.ListOptions, int(count)),
	}, nil
}

// Get retrieves a Run domain obj
func (db RunDB) Get(opts otf.RunGetOptions) (*otf.Run, error) {
	run, err := getRun(db.DB, opts)
	if err != nil {
		return nil, err
	}
	return run.ToDomain(), nil
}

func getRun(db *sqlx.DB, opts otf.RunGetOptions) (*otf.Run, error) {
	runs := sq.Select("*").From("runs")

	switch {
	case opts.ID != nil:
		runs = runs.Where("external_id = ?", opts.ID)
	case opts.PlanID != nil:
		runs = runs.Join("plans ON plans.run_id = runs.id").Where("plans.external_id = ?", opts.PlanID)
	case opts.ApplyID != nil:
		runs = runs.Join("applies ON applies.run_id = runs.id").Where("applies.external_id = ?", opts.ApplyID)
	default:
		return nil, otf.ErrInvalidRunGetOptions
	}

	sql, args, err := runs.ToSql()
	if err != nil {
		return nil, err
	}

	m := map[string]interface{}{}
	if err := db.QueryRowx(sql, args).MapScan(m); err != nil {
		return nil, err
	}

	return &model, nil
}
