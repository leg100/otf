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
	_ otf.RunStore = (*RunDB)(nil)

	runColumnsWithoutID = []string{"created_at", "updated_at", "external_id", "force_cancel_available_at", "is_destroy", "position_in_queue", "refresh", "refresh_only", "status", "status_timestamps", "replace_addrs", "target_addrs"}
	runColumns          = append(runColumnsWithoutID, "id")

	planColumnsWithoutID = []string{"created_at", "updated_at", "resource_additions", "resource_changes", "resource_destructions", "status", "status_timestamps", "logs_blob_id", "plan_file_blob_id", "plan_json_blob_id", "run_id"}
	planColumns          = append(planColumnsWithoutID, "id")

	applyColumnsWithoutID = []string{"created_at", "updated_at", "resource_additions", "resource_changes", "resource_destructions", "status", "status_timestamps", "logs_blob_id", "run_id"}
	applyColumns          = append(applyColumnsWithoutID, "id")

	insertRunSQL = fmt.Sprintf("INSERT INTO runs (%s, workspace_id, configuration_version_id) VALUES (%s, :workspaces.id, :configuration_versions.id)",
		strings.Join(runColumnsWithoutID, ", "),
		strings.Join(otf.PrefixSlice(runColumnsWithoutID, ":"), ", "))

	insertPlanSQL = fmt.Sprintf("INSERT INTO plans (%s) VALUES (%s)",
		strings.Join(planColumnsWithoutID, ", "),
		strings.Join(otf.PrefixSlice(planColumnsWithoutID, ":"), ", "))

	insertApplySQL = fmt.Sprintf("INSERT INTO applies (%s) VALUES (%s)",
		strings.Join(applyColumnsWithoutID, ", "),
		strings.Join(otf.PrefixSlice(applyColumnsWithoutID, ":"), ", "))
)

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
	result, err := tx.NamedExec(insertRunSQL, run)
	if err != nil {
		return nil, err
	}
	run.Model.ID, err = result.LastInsertId()
	if err != nil {
		return nil, err
	}

	// Insert plan
	run.Plan.RunID = run.Model.ID
	result, err = tx.NamedExec(insertPlanSQL, run.Plan)
	if err != nil {
		return nil, err
	}
	run.Plan.Model.ID, err = result.LastInsertId()
	if err != nil {
		return nil, err
	}

	// Insert apply
	run.Apply.RunID = run.Model.ID
	result, err = tx.NamedExec(insertApplySQL, run.Apply)
	if err != nil {
		return nil, err
	}
	run.Apply.Model.ID, err = result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return run, tx.Commit()
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

	// Make a copy for comparison with the updated obj
	before, err := copystructure.Copy(run)
	if err != nil {
		return nil, err
	}

	// Update obj using client-supplied fn
	if err := fn(run); err != nil {
		return nil, err
	}

	updates := FindUpdates(db.Mapper, before.(*otf.Run), run)
	if len(updates) == 0 {
		return run, nil
	}

	run.UpdatedAt = time.Now()
	updates["updated_at"] = run.UpdatedAt

	sql := sq.Update("runs").Where("id = ?", run.Model.ID)

	query, args, err := sql.SetMap(updates).ToSql()
	if err != nil {
		return nil, err
	}

	_, err = tx.Exec(query, args...)
	if err != nil {
		return nil, fmt.Errorf("executing SQL statement: %s: %w", query, err)
	}

	return run, tx.Commit()
}

func (db RunDB) List(opts otf.RunListOptions) (*otf.RunList, error) {
	selectBuilder := sq.Select(asColumnList("runs", false, runColumns...)).
		Columns(asColumnList("plans", true, planColumns...)).
		Columns(asColumnList("applies", true, applyColumns...)).
		Columns(asColumnList("configuration_versions", true, configurationVersionColumns...)).
		Columns(asColumnList("workspaces", true, workspaceColumns...)).
		From("runs").
		Join("plans ON plans.run_id = runs.id").
		Join("applies ON applies.run_id = runs.id").
		Join("configuration_versions ON configuration_versions.id = runs.configuration_version_id").
		Join("workspaces ON workspaces.id = runs.workspace_id").
		Limit(opts.GetLimit()).
		Offset(opts.GetOffset())

	// Optionally filter by workspace
	if opts.WorkspaceID != nil {
		selectBuilder.Where("workspaces.external_id = ?", *opts.WorkspaceID)
	}

	// Optionally filter by statuses
	if len(opts.Statuses) > 0 {
		selectBuilder.Where(sq.Eq{"runs.status": opts.Statuses})
	}

	query, args, err := selectBuilder.ToSql()
	if err != nil {
		return nil, err
	}

	var result []otf.Run
	if err := db.Select(&result, query, args...); err != nil {
		return nil, fmt.Errorf("unable to scan runs from DB: %w", err)
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
	selectBuilder := sq.Select(asColumnList("runs", false, runColumns...)).
		Columns(asColumnList("plans", true, planColumns...)).
		Columns(asColumnList("applies", true, applyColumns...)).
		Columns(asColumnList("configuration_versions", true, configurationVersionColumns...)).
		Columns(asColumnList("workspaces", true, workspaceColumns...)).
		From("runs").
		Join("plans ON plans.run_id = runs.id").
		Join("applies ON applies.run_id = runs.id").
		Join("configuration_versions ON configuration_versions.id = runs.configuration_version_id").
		Join("workspaces ON workspaces.id = runs.workspace_id")

	switch {
	case opts.ID != nil:
		selectBuilder = selectBuilder.Where("runs.external_id = ?", *opts.ID)
	case opts.PlanID != nil:
		selectBuilder = selectBuilder.Where("plans.external_id = ?", *opts.PlanID)
	case opts.ApplyID != nil:
		selectBuilder = selectBuilder.Where("applies.external_id = ?", *opts.ApplyID)
	default:
		return nil, otf.ErrInvalidRunGetOptions
	}

	sql, args, err := selectBuilder.ToSql()
	if err != nil {
		return nil, err
	}

	var run otf.Run
	if err := db.Get(&run, sql, args...); err != nil {
		return nil, err
	}

	return &run, nil
}
