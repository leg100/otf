package sqlite

import (
	"fmt"
	"strings"
	"time"

	"github.com/jinzhu/copier"
	"github.com/jmoiron/sqlx"
	"github.com/leg100/otf"
)

var (
	_ otf.RunStore = (*RunDB)(nil)

	insertRunSql = `INSERT INTO runs (created_at, updated_at, external_id, force_cancel_available_at, is_destroy, message, position_in_queue, refresh, refresh_only, status, status_timestamps, replace_addrs, target_addrs, workspace_id, configuration_version_id)
VALUES (:created_at,:updated_at,:external_id,:force_cancel_available_at,?,?,?,?,?,?,?,?,?,?)
`

	insertPlanSql = `INSERT INTO plans (created_at, updated_at, external_id, logs_blob_id, plan_file_blob_id, plan_json_blob_id, status, status_timestamps, run_id
VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?)
`

	insertApplySql = `INSERT INTO applies (created_at, updated_at, external_id, logs_blob_id, status, status_timestamps, run_id
VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?)
`

	runColumns = []string{"id", "created_at", "updated_at", "external_id", "force_cancel_available_at", "is_destroy", "position_in_queue", "refresh", "refresh_only", "status", "status_timestamps", "replace_addrs", "target_addrs", "workspace_id", "configuration_version_id"}

	planColumns = []string{"id", "created_at", "updated_at", "resource_additions", "resource_changes", "resource_destructions", "status", "status_timestamps", "logs_blob_id", "plan_file_blob_id", "plan_json_blob_id", "run_id"}

	applyColumns = []string{"id", "created_at", "updated_at", "resource_additions", "resource_changes", "resource_destructions", "status", "status_timestamps", "logs_blob_id", "run_id"}

	listRunsSql = fmt.Sprintf(`SELECT %s, %s, %s, %s, %s
FROM runs
JOIN plans ON plans.run_id = runs.id
JOIN applies ON applies.run_id = runs.id
JOIN configuration_versions ON configuration_versions.id = runs.configuration_version_id
JOIN workspaces ON workspaces.id = runs.workspace_id
`, asColumnList("runs", false, runColumns...), asColumnList("plans", true, planColumns...), asColumnList("applies", true, applyColumns...), asColumnList("configuration_versions", true, configurationVersionColumns...), asColumnList("workspaces", true, workspaceColumns...))

	getRunSql = fmt.Sprintf(`SELECT %s, %s, %s, %s, %s
FROM runs
JOIN plans ON plans.run_id = runs.id
JOIN applies ON applies.run_id = runs.id
JOIN configuration_versions ON configuration_versions.id = runs.configuration_version_id
JOIN workspaces ON workspaces.id = runs.workspace_id
`, asColumnList("runs", false, runColumns...), asColumnList("plans", true, planColumns...), asColumnList("applies", true, applyColumns...), asColumnList("configuration_versions", true, configurationVersionColumns...), asColumnList("workspaces", true, workspaceColumns...))
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
	result, err := tx.NamedExec(insertRunSql, run)
	if err != nil {
		return nil, err
	}
	run.Model.ID, err = result.LastInsertId()
	if err != nil {
		return nil, err
	}

	// Insert plan
	run.Plan.RunID = run.Model.ID
	result, err = tx.NamedExec(insertPlanSql, run.Plan)
	if err != nil {
		return nil, err
	}

	// Insert apply
	run.Apply.RunID = run.Model.ID
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
		fmt.Fprintf(&sql, "SET %s = :%[1]s\n", k)
	}

	fmt.Fprintf(&sql, "WHERE %d = :id", run.Model.ID)

	_, err = db.NamedExec(sql.String(), updates)
	if err != nil {
		return nil, err
	}

	return run, tx.Commit()
}

func (db RunDB) List(opts otf.RunListOptions) (*otf.RunList, error) {
	limit, offset := opts.GetSQLWindow()

	params := map[string]interface{}{
		"limit":  limit,
		"offset": offset,
	}

	var conditions []string

	// Optionally filter by workspace
	if opts.WorkspaceID != nil {
		conditions = append(conditions, "workspaces.external_id = :workspace_external_id")
		params["workspace_external_id"] = *opts.WorkspaceID
	}

	// Optionally filter by statuses
	if len(opts.Statuses) > 0 {
		conditions = append(conditions, "runs.status IN (:statuses)")
		params["statuses"] = opts.Statuses
	}

	sql := listRunsSql
	if len(conditions) > 0 {
		sql += fmt.Sprintln("WHERE", strings.Join(conditions, " AND "))
	}
	sql += " LIMIT :limit OFFSET :offset "

	query, args, err := sqlx.Named(sql, params)
	if err != nil {
		return nil, err
	}
	query, args, err = sqlx.In(query, args...)
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
	var condition, arg string

	switch {
	case opts.ID != nil:
		condition = "runs.external_id = :id"
		arg = *opts.ID
	case opts.PlanID != nil:
		condition = "plans.external_id = :plan_id"
		arg = *opts.PlanID
	case opts.ApplyID != nil:
		condition = "applies.external_id = :apply_id"
		arg = *opts.ApplyID
	default:
		return nil, otf.ErrInvalidRunGetOptions
	}

	sql := fmt.Sprintf(getRunSql, "WHERE", condition)

	var run otf.Run
	if err := db.Get(&run, sql, arg); err != nil {
		return nil, err
	}

	return &run, nil
}
