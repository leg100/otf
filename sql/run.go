package sql

import (
	"fmt"
	"strings"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"github.com/leg100/otf"
	"github.com/mitchellh/copystructure"
)

var (
	_ otf.RunStore = (*RunDB)(nil)

	runColumnsWithoutID = []string{"created_at", "updated_at", "external_id", "is_destroy", "position_in_queue", "refresh", "refresh_only", "status", "status_timestamps", "replace_addrs", "target_addrs"}
	runColumns          = append(runColumnsWithoutID, "id")

	planColumnsWithoutID = []string{"created_at", "updated_at", "external_id", "resource_additions", "resource_changes", "resource_destructions", "status", "status_timestamps", "logs_blob_id", "plan_file_blob_id", "plan_json_blob_id", "run_id"}
	planColumns          = append(planColumnsWithoutID, "id")

	applyColumnsWithoutID = []string{"created_at", "updated_at", "external_id", "resource_additions", "resource_changes", "resource_destructions", "status", "status_timestamps", "logs_blob_id", "run_id"}
	applyColumns          = append(applyColumnsWithoutID, "id")

	insertRunSQL = fmt.Sprintf("INSERT INTO runs (%s, workspace_id, configuration_version_id) VALUES (%s, :workspaces.id, :configuration_versions.id) RETURNING id",
		strings.Join(runColumnsWithoutID, ", "),
		strings.Join(otf.PrefixSlice(runColumnsWithoutID, ":"), ", "))

	insertPlanSQL = fmt.Sprintf("INSERT INTO plans (%s) VALUES (%s) RETURNING id",
		strings.Join(planColumnsWithoutID, ", "),
		strings.Join(otf.PrefixSlice(planColumnsWithoutID, ":"), ", "))

	insertApplySQL = fmt.Sprintf("INSERT INTO applies (%s) VALUES (%s) RETURNING id",
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
	sql, args, err := tx.BindNamed(insertRunSQL, run)
	if err != nil {
		return nil, err
	}
	if err := tx.Get(&run.Model.ID, sql, args...); err != nil {
		return nil, err
	}

	// Insert plan
	run.Plan.RunID = run.Model.ID
	sql, args, err = tx.BindNamed(insertPlanSQL, run.Plan)
	if err != nil {
		return nil, err
	}
	if err := tx.Get(&run.Plan.Model.ID, sql, args...); err != nil {
		return nil, err
	}

	// Insert apply
	run.Apply.RunID = run.Model.ID
	sql, args, err = tx.BindNamed(insertApplySQL, run.Apply)
	if err != nil {
		return nil, err
	}
	if err := tx.Get(&run.Apply.Model.ID, sql, args...); err != nil {
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

	runUpdated, err := update(db.Mapper, tx, "runs", before.(*otf.Run), run)
	if err != nil {
		return nil, err
	}

	planUpdated, err := update(db.Mapper, tx, "plans", before.(*otf.Run).Plan, run.Plan)
	if err != nil {
		return nil, err
	}

	applyUpdated, err := update(db.Mapper, tx, "applies", before.(*otf.Run).Apply, run.Apply)
	if err != nil {
		return nil, err
	}

	if runUpdated || planUpdated || applyUpdated {
		return run, tx.Commit()
	}

	return run, nil
}

func (db RunDB) List(opts otf.RunListOptions) (*otf.RunList, error) {
	selectBuilder := psql.Select().
		From("runs").
		Join("plans ON plans.run_id = runs.id").
		Join("applies ON applies.run_id = runs.id").
		Join("configuration_versions ON configuration_versions.id = runs.configuration_version_id").
		Join("workspaces ON workspaces.id = runs.workspace_id")

	// Optionally filter by workspace
	if opts.WorkspaceID != nil {
		selectBuilder = selectBuilder.Where("workspaces.external_id = ?", *opts.WorkspaceID)
	}

	// Optionally filter by statuses
	if len(opts.Statuses) > 0 {
		selectBuilder = selectBuilder.Where(sq.Eq{"runs.status": opts.Statuses})
	}

	var count int
	if err := selectBuilder.Columns("count(1)").RunWith(db).QueryRow().Scan(&count); err != nil {
		return nil, fmt.Errorf("counting total rows: %w", err)
	}

	selectBuilder = selectBuilder.
		Columns(asColumnList("runs", false, runColumns...)).
		Columns(asColumnList("plans", true, planColumns...)).
		Columns(asColumnList("applies", true, applyColumns...)).
		Columns(asColumnList("configuration_versions", true, configurationVersionColumns...)).
		Columns(asColumnList("workspaces", true, workspaceColumns...)).
		Limit(opts.GetLimit()).
		Offset(opts.GetOffset())

	sql, args, err := selectBuilder.ToSql()
	if err != nil {
		return nil, err
	}

	var items []*otf.Run
	if err := db.Select(&items, sql, args...); err != nil {
		return nil, fmt.Errorf("unable to scan runs from DB: %w", err)
	}

	return &otf.RunList{
		Items:      items,
		Pagination: otf.NewPagination(opts.ListOptions, count),
	}, nil
}

// Get retrieves a Run domain obj
func (db RunDB) Get(opts otf.RunGetOptions) (*otf.Run, error) {
	return getRun(db.DB, opts)
}

// Delete deletes a run from the DB
func (db RunDB) Delete(id string) error {
	tx := db.MustBegin()
	defer tx.Rollback()

	run, err := getRun(tx, otf.RunGetOptions{ID: otf.String(id)})
	if err != nil {
		return err
	}

	_, err = tx.Exec("DELETE FROM runs WHERE id = $1", run.Model.ID)
	if err != nil {
		return fmt.Errorf("unable to delete run: %w", err)
	}

	return tx.Commit()
}

func getRun(db Getter, opts otf.RunGetOptions) (*otf.Run, error) {
	selectBuilder := psql.Select(asColumnList("runs", false, runColumns...)).
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
		return nil, databaseError(err)
	}

	return &run, nil
}
