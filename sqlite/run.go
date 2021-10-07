package sqlite

import (
	"fmt"
	"strings"
	"time"

	"github.com/jinzhu/copier"
	"github.com/jmoiron/sqlx"
	"github.com/leg100/otf"
)

var _ otf.RunStore = (*RunDB)(nil)

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
	tx := db.MustBegin()
	defer tx.Rollback()

	// Insert run
	result, err := tx.NamedExec(insertRunSql, run)
	if err != nil {
		return nil, err
	}
	run.Model.ID, err = result.LastInsertId()
	if err != nil {
		return nil, err
	}

	// Insert plan
	result, err = tx.NamedExec(insertPlanSql, run.Plan)
	if err != nil {
		return nil, err
	}

	// Insert apply
	result, err = tx.NamedExec(insertApplySql, run.Apply)
	if err != nil {
		return nil, err
	}

	return run, nil
}

// Update persists an updated Run to the DB. The existing run is fetched from
// the DB, the supplied func is invoked on the run, and the updated run is
// persisted back to the DB. The returned Run includes any changes, including a
// new UpdatedAt value.
func (db RunDB) Update(id string, fn func(*otf.Run) error) (*otf.Run, error) {
	tx := db.MustBegin()
	defer tx.Rollback()

	run, err := getRun(tx, otf.RunGetOptions{ID: &id})
	if err != nil {
		return nil, err
	}

	before := otf.Run{}
	copier.Copy(&before, run)

	// Update obj using client-supplied fn
	if err := fn(run); err != nil {
		return nil, err
	}

	updates := FindUpdates(db.Mapper, before, run)
	if len(updates) == 0 {
		return run, nil
	}

	run.UpdatedAt = time.Now()
	updates["updated_at"] = run.UpdatedAt

	var sql strings.Builder
	fmt.Fprintln(&sql, "UPDATE runs")

	for k := range updates {
		fmt.Fprintln(&sql, "SET %s = :%[1]s", k)
	}

	fmt.Fprintf(&sql, "WHERE %s = :id", run.Model.ID)

	_, err = db.NamedExec(sql.String(), updates)
	if err != nil {
		return nil, err
	}

	return run, tx.Commit()
}

func (db RunDB) List(opts otf.RunListOptions) (*otf.RunList, error) {
	type listRunParams struct {
		Limit               int
		Offset              int
		WorkspaceExternalID string
		Statuses            []otf.RunStatus
	}

	params := listRunParams{}

	var sql strings.Builder
	fmt.Fprintln(&sql, "SELECT", getRunColumns, "FROM", "runs", getRunJoins)

	var conditions []string

	// Optionally filter by workspace
	if opts.WorkspaceID != nil {
		conditions = append(conditions, "workspaces.external_id = :workspace_external_id")
		params.WorkspaceExternalID = *opts.WorkspaceID
	}

	// Optionally filter by statuses
	if len(opts.Statuses) > 0 {
		conditions = append(conditions, "run.status IN (:statuses)")
		params.Statuses = opts.Statuses
	}

	fmt.Fprintln(&sql, "WHERE", strings.Join(conditions, " AND "))

	if opts.PageSize > 0 {
		params.Limit = opts.PageSize
	}

	if opts.PageNumber > 0 {
		params.Offset = (opts.PageNumber - 1) * opts.PageSize
	}

	query, args, err := sqlx.Named(sql.String(), params)
	if err != nil {
		return nil, err
	}
	query, args, err = sqlx.In(query, args)
	if err != nil {
		return nil, err
	}

	var result []otf.Run
	if err := db.Select(&result, query, args); err != nil {
		return nil, err
	}

	// Convert from []otf.Run to []*otf.Run
	var items []*otf.Run
	for _, r := range result {
		items = append(items, &r)
	}

	return &otf.RunList{
		Items:      items,
		Pagination: otf.NewPagination(opts.ListOptions, len(items)),
	}, nil
}

// Get retrieves a Run domain obj
func (db RunDB) Get(opts otf.RunGetOptions) (*otf.Run, error) {
	return getRun(db.DB, opts)
}

func getRun(db Getter, opts otf.RunGetOptions) (*otf.Run, error) {
	type getRunParams struct {
		ID      string
		PlanID  string
		ApplyID string
	}
	params := getRunParams{}

	var sql strings.Builder
	fmt.Fprintln(&sql, "SELECT", getRunColumns, "FROM", "runs", getRunJoins)

	var condition string

	switch {
	case opts.ID != nil:
		condition = "runs.external_id = ?"
		params.ID = *opts.ID
	case opts.PlanID != nil:
		condition = "plans.external_id = ?"
		params.PlanID = *opts.PlanID
	case opts.ApplyID != nil:
		condition = "applies.external_id = ?"
		params.ApplyID = *opts.ApplyID
	default:
		return nil, otf.ErrInvalidRunGetOptions
	}

	fmt.Fprintln(&sql, "WHERE", condition)

	var run otf.Run
	if err := db.Get(&run, sql.String(), params); err != nil {
		return nil, err
	}

	return &run, nil
}
