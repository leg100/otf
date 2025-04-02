package run

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/configversion"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/runstatus"
	"github.com/leg100/otf/internal/sql"
	"github.com/leg100/otf/internal/workspace"
)

// pgdb is a database of runs on postgres
type pgdb struct {
	*sql.DB // provides access to generated SQL queries
}

// CreateRun persists a Run to the DB.
func (db *pgdb) CreateRun(ctx context.Context, run *Run) error {
	return db.Tx(ctx, func(ctx context.Context, conn sql.Connection) error {
		_, err := db.Exec(ctx, `
INSERT INTO runs (
    run_id,
    created_at,
    is_destroy,
    position_in_queue,
    refresh,
    refresh_only,
    source,
    status,
    replace_addrs,
    target_addrs,
    auto_apply,
    plan_only,
    configuration_version_id,
    workspace_id,
    created_by,
    terraform_version,
    allow_empty_apply
) VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6,
    $7,
    $8,
    $9,
    $10,
    $11,
    $12,
    $13,
    $14,
    $15,
    $16,
    $17
)`,

			run.ID,
			run.CreatedAt,
			run.IsDestroy,
			0,
			run.Refresh,
			run.RefreshOnly,
			run.Source,
			run.Status,
			run.ReplaceAddrs,
			run.TargetAddrs,
			run.AutoApply,
			run.PlanOnly,
			run.ConfigurationVersionID,
			run.WorkspaceID,
			run.CreatedBy,
			run.TerraformVersion,
			run.AllowEmptyApply,
		)
		for _, v := range run.Variables {
			_, err := db.Exec(ctx, `INSERT INTO run_variables ( run_id, key, value) VALUES ( $1, $2, $3)`,
				run.ID,
				v.Key,
				v.Value,
			)
			if err != nil {
				return fmt.Errorf("inserting run variable: %w", err)
			}
		}
		if err != nil {
			return fmt.Errorf("inserting run: %w", err)
		}
		_, err = db.Exec(ctx, `INSERT INTO plans (run_id, status) VALUES ($1, $2)`,
			run.ID,
			run.Plan.Status,
		)
		if err != nil {
			return fmt.Errorf("inserting plan: %w", err)
		}
		_, err = db.Exec(ctx, `INSERT INTO applies (run_id, status) VALUES ($1, $2)`,
			run.ID,
			run.Apply.Status,
		)
		if err != nil {
			return fmt.Errorf("inserting apply: %w", err)
		}
		if err := db.insertRunStatusTimestamp(ctx, run); err != nil {
			return fmt.Errorf("inserting run status timestamp: %w", err)
		}
		if err := db.insertPhaseStatusTimestamp(ctx, run.Plan); err != nil {
			return fmt.Errorf("inserting plan status timestamp: %w", err)
		}
		if err := db.insertPhaseStatusTimestamp(ctx, run.Apply); err != nil {
			return fmt.Errorf("inserting apply status timestamp: %w", err)
		}
		return nil
	})
}

// UpdateStatus updates the run status as well as its plan and/or apply.
func (db *pgdb) UpdateStatus(ctx context.Context, runID resource.TfeID, fn func(context.Context, *Run) error) (*Run, error) {
	var (
		runStatus        runstatus.Status
		planStatus       PhaseStatus
		applyStatus      PhaseStatus
		cancelSignaledAt *time.Time
	)
	return sql.Updater(
		ctx,
		db.DB,
		func(ctx context.Context, conn sql.Connection) (*Run, error) {
			row := db.Query(ctx, `
SELECT
    runs.run_id,
    runs.created_at,
    runs.cancel_signaled_at,
    runs.is_destroy,
    runs.position_in_queue,
    runs.refresh,
    runs.refresh_only,
    runs.source,
    runs.status,
    plans.status        AS plan_status,
    applies.status      AS apply_status,
    runs.replace_addrs,
    runs.target_addrs,
    runs.auto_apply,
    plans.resource_report::"report" AS plan_resource_report,
    plans.output_report::"report" AS plan_output_report,
    applies.resource_report::"report" AS apply_resource_report,
    runs.configuration_version_id,
    runs.workspace_id,
    runs.plan_only,
    runs.created_by,
    runs.terraform_version,
    runs.allow_empty_apply,
    workspaces.execution_mode AS execution_mode,
    CASE WHEN workspaces.latest_run_id = runs.run_id THEN true
         ELSE false
    END AS latest,
    workspaces.organization_name,
    organizations.cost_estimation_enabled,
    rst.run_status_timestamps,
    pst.plan_status_timestamps,
    ast.apply_status_timestamps,
    rv.run_variables,
    ia::"ingress_attributes" AS ingress_attributes
FROM runs
JOIN plans USING (run_id)
JOIN applies USING (run_id)
JOIN workspaces ON runs.workspace_id = workspaces.workspace_id
JOIN organizations ON workspaces.organization_name = organizations.name
JOIN (configuration_versions cv LEFT JOIN ingress_attributes ia USING (configuration_version_id)) USING (configuration_version_id)
LEFT JOIN (
    SELECT
        run_id,
        array_agg(rv.*)::run_variables[] AS run_variables
    FROM run_variables rv
    GROUP BY run_id
) AS rv ON rv.run_id = runs.run_id
LEFT JOIN (
    SELECT
        run_id,
        array_agg(rst.*)::run_status_timestamps[] AS run_status_timestamps
    FROM run_status_timestamps rst
    GROUP BY run_id
) AS rst ON rst.run_id = runs.run_id
LEFT JOIN (
    SELECT
        run_id,
        array_agg(pst.*)::phase_status_timestamps[] AS plan_status_timestamps
    FROM phase_status_timestamps pst
    WHERE pst.phase = 'plan'
    GROUP BY run_id
) AS pst ON pst.run_id = runs.run_id
LEFT JOIN (
    SELECT
        run_id,
        array_agg(ast.*)::phase_status_timestamps[] AS apply_status_timestamps
    FROM phase_status_timestamps ast
    WHERE ast.phase = 'apply'
    GROUP BY run_id
) AS ast ON ast.run_id = runs.run_id
WHERE runs.run_id = $1
FOR UPDATE OF runs, plans, applies
`, runID)
			run, err := sql.CollectOneRow(row, scan)
			if err != nil {
				return nil, err
			}
			// Make copies of run attributes before update
			runStatus = run.Status
			planStatus = run.Plan.Status
			applyStatus = run.Apply.Status
			cancelSignaledAt = run.CancelSignaledAt
			return run, nil
		},
		fn,
		func(ctx context.Context, conn sql.Connection, run *Run) error {
			if run.Status != runStatus {
				_, err := db.Exec(ctx, `
UPDATE runs
SET
    status = $1
WHERE run_id = $2
`,
					run.Status,
					run.ID,
				)
				if err != nil {
					return err
				}

				if err := db.insertRunStatusTimestamp(ctx, run); err != nil {
					return err
				}
			}

			if run.Plan.Status != planStatus {
				_, err := db.Exec(ctx, `
UPDATE plans
SET status = $1
WHERE run_id = $2
`,
					run.Plan.Status,
					run.ID,
				)
				if err != nil {
					return err
				}

				if err := db.insertPhaseStatusTimestamp(ctx, run.Plan); err != nil {
					return err
				}
			}

			if run.Apply.Status != applyStatus {
				_, err := db.Exec(ctx, `
UPDATE applies
SET status = $1
WHERE run_id = $2
`,
					run.Apply.Status,
					run.ID,
				)
				if err != nil {
					return err
				}

				if err := db.insertPhaseStatusTimestamp(ctx, run.Apply); err != nil {
					return err
				}
			}

			if run.CancelSignaledAt != cancelSignaledAt && run.CancelSignaledAt != nil {
				_, err := db.Exec(ctx, `
UPDATE runs
SET
    cancel_signaled_at = $1
WHERE run_id = $2
`,
					*run.CancelSignaledAt,
					run.ID,
				)
				if err != nil {
					return err
				}
			}

			return nil
		},
	)
}

func (db *pgdb) CreatePlanReport(ctx context.Context, runID resource.TfeID, resource, output Report) error {
	_, err := db.Exec(ctx, `
UPDATE plans
SET resource_report = (
        $1,
        $2,
        $3
    ),
    output_report = (
        $4,
        $5,
        $6
    )
WHERE run_id = $7
`,
		resource.Additions,
		resource.Changes,
		resource.Destructions,
		output.Additions,
		output.Changes,
		output.Destructions,
		runID,
	)
	return err
}

func (db *pgdb) CreateApplyReport(ctx context.Context, runID resource.TfeID, report Report) error {
	_, err := db.Exec(ctx, `
UPDATE applies
SET resource_report = (
    $1,
    $2,
    $3
)
WHERE run_id = $4
`,
		report.Additions,
		report.Changes,
		report.Destructions,
		runID,
	)
	return err
}

func (db *pgdb) ListRuns(ctx context.Context, opts ListOptions) (*resource.Page[*Run], error) {
	organization := "%"
	if opts.Organization != nil {
		organization = opts.Organization.String()
	}
	workspaceName := "%"
	if opts.WorkspaceName != nil {
		workspaceName = *opts.WorkspaceName
	}
	workspaceID := "%"
	if opts.WorkspaceID != nil {
		workspaceID = opts.WorkspaceID.String()
	}
	sources := []string{"%"}
	if len(opts.Sources) > 0 {
		sources = internal.ToStringSlice(opts.Sources)
	}
	statuses := []string{"%"}
	if len(opts.Statuses) > 0 {
		statuses = internal.ToStringSlice(opts.Statuses)
	}
	planOnly := "%"
	if opts.PlanOnly != nil {
		planOnly = strconv.FormatBool(*opts.PlanOnly)
	}
	rows := db.Query(ctx, `
SELECT
    runs.run_id,
    runs.created_at,
    runs.cancel_signaled_at,
    runs.is_destroy,
    runs.position_in_queue,
    runs.refresh,
    runs.refresh_only,
    runs.source,
    runs.status,
    plans.status      AS plan_status,
    applies.status      AS apply_status,
    runs.replace_addrs,
    runs.target_addrs,
    runs.auto_apply,
    plans.resource_report::"report" AS plan_resource_report,
    plans.output_report::"report" AS plan_output_report,
    applies.resource_report::"report" AS apply_resource_report,
    runs.configuration_version_id,
    runs.workspace_id,
    runs.plan_only,
    runs.created_by,
    runs.terraform_version,
    runs.allow_empty_apply,
    workspaces.execution_mode AS execution_mode,
    CASE WHEN workspaces.latest_run_id = runs.run_id THEN true
         ELSE false
    END AS latest,
    workspaces.organization_name,
    organizations.cost_estimation_enabled,
    rst.run_status_timestamps,
    pst.plan_status_timestamps,
    ast.apply_status_timestamps,
    rv.run_variables,
    ia::"ingress_attributes" AS ingress_attributes
FROM runs
JOIN plans USING (run_id)
JOIN applies USING (run_id)
JOIN workspaces ON runs.workspace_id = workspaces.workspace_id
JOIN organizations ON workspaces.organization_name = organizations.name
JOIN (configuration_versions cv LEFT JOIN ingress_attributes ia USING (configuration_version_id)) USING (configuration_version_id)
LEFT JOIN (
    SELECT
        run_id,
        array_agg(rv.*)::run_variables[] AS run_variables
    FROM run_variables rv
    GROUP BY run_id
) AS rv ON rv.run_id = runs.run_id
LEFT JOIN (
    SELECT
        run_id,
        array_agg(rst.*)::run_status_timestamps[] AS run_status_timestamps
    FROM run_status_timestamps rst
    GROUP BY run_id
) AS rst ON rst.run_id = runs.run_id
LEFT JOIN (
    SELECT
        run_id,
        array_agg(pst.*)::phase_status_timestamps[] AS plan_status_timestamps
    FROM phase_status_timestamps pst
    WHERE pst.phase = 'plan'
    GROUP BY run_id
) AS pst ON pst.run_id = runs.run_id
LEFT JOIN (
    SELECT
        run_id,
        array_agg(ast.*)::phase_status_timestamps[] AS apply_status_timestamps
    FROM phase_status_timestamps ast
    WHERE ast.phase = 'apply'
    GROUP BY run_id
) AS ast ON ast.run_id = runs.run_id
WHERE
    workspaces.organization_name LIKE ANY($1::text[])
AND workspaces.workspace_id      LIKE ANY($2::text[])
AND workspaces.name              LIKE ANY($3::text[])
AND runs.source                  LIKE ANY($4::text[])
AND runs.status                  LIKE ANY($5::text[])
AND runs.plan_only::text         LIKE ANY($6::text[])
AND (($7::text IS NULL) OR ia.commit_sha = $7)
AND (($8::text IS NULL) OR ia.sender_username = $8)
ORDER BY runs.created_at DESC
LIMIT $9::int
OFFSET $10::int
`,
		[]string{organization},
		[]string{workspaceID},
		[]string{workspaceName},
		sources,
		statuses,
		[]string{planOnly},
		opts.CommitSHA,
		opts.VCSUsername,
		sql.GetLimit(opts.PageOptions),
		sql.GetOffset(opts.PageOptions),
	)
	items, err := sql.CollectRows(rows, scan)
	if err != nil {
		return nil, fmt.Errorf("querying runs: %w", err)
	}
	count, err := db.Int(ctx, `
SELECT count(*)
FROM runs
JOIN workspaces USING(workspace_id)
JOIN (configuration_versions LEFT JOIN ingress_attributes ia USING (configuration_version_id)) USING (configuration_version_id)
WHERE
    workspaces.organization_name LIKE ANY($1::text[])
AND workspaces.workspace_id      LIKE ANY($2::text[])
AND workspaces.name              LIKE ANY($3::text[])
AND runs.source                  LIKE ANY($4::text[])
AND runs.status                  LIKE ANY($5::text[])
AND runs.plan_only::text         LIKE ANY($6::text[])
AND (($7::text IS NULL) OR ia.commit_sha = $7)
AND (($8::text IS NULL) OR ia.sender_username = $8)
`,

		[]string{organization},
		[]string{workspaceID},
		[]string{workspaceName},
		sources,
		statuses,
		[]string{planOnly},
		opts.CommitSHA,
		opts.VCSUsername,
	)
	if err != nil {
		return nil, fmt.Errorf("counting runs: %w", err)
	}
	return resource.NewPage(items, opts.PageOptions, internal.Int64(count)), nil
}

// get retrieves a run using the get options
func (db *pgdb) get(ctx context.Context, runID resource.ID) (*Run, error) {
	rows := db.Query(ctx, `
SELECT
    runs.run_id,
    runs.created_at,
    runs.cancel_signaled_at,
    runs.is_destroy,
    runs.position_in_queue,
    runs.refresh,
    runs.refresh_only,
    runs.source,
    runs.status,
    plans.status      AS plan_status,
    applies.status      AS apply_status,
    runs.replace_addrs,
    runs.target_addrs,
    runs.auto_apply,
    plans.resource_report::"report" AS plan_resource_report,
    plans.output_report::"report" AS plan_output_report,
    applies.resource_report::"report" AS apply_resource_report,
    runs.configuration_version_id,
    runs.workspace_id,
    runs.plan_only,
    runs.created_by,
    runs.terraform_version,
    runs.allow_empty_apply,
    workspaces.execution_mode AS execution_mode,
    CASE WHEN workspaces.latest_run_id = runs.run_id THEN true
         ELSE false
    END AS latest,
    workspaces.organization_name,
    organizations.cost_estimation_enabled,
    rst.run_status_timestamps,
    pst.plan_status_timestamps,
    ast.apply_status_timestamps,
    rv.run_variables,
    ia::"ingress_attributes" AS ingress_attributes
FROM runs
JOIN plans USING (run_id)
JOIN applies USING (run_id)
JOIN workspaces ON runs.workspace_id = workspaces.workspace_id
JOIN organizations ON workspaces.organization_name = organizations.name
JOIN (configuration_versions cv LEFT JOIN ingress_attributes ia USING (configuration_version_id)) USING (configuration_version_id)
LEFT JOIN (
    SELECT
        run_id,
        array_agg(rv.*)::run_variables[] AS run_variables
    FROM run_variables rv
    GROUP BY run_id
) AS rv ON rv.run_id = runs.run_id
LEFT JOIN (
    SELECT
        run_id,
        array_agg(rst.*)::run_status_timestamps[] AS run_status_timestamps
    FROM run_status_timestamps rst
    GROUP BY run_id
) AS rst ON rst.run_id = runs.run_id
LEFT JOIN (
    SELECT
        run_id,
        array_agg(pst.*)::phase_status_timestamps[] AS plan_status_timestamps
    FROM phase_status_timestamps pst
    WHERE pst.phase = 'plan'
    GROUP BY run_id
) AS pst ON pst.run_id = runs.run_id
LEFT JOIN (
    SELECT
        run_id,
        array_agg(ast.*)::phase_status_timestamps[] AS apply_status_timestamps
    FROM phase_status_timestamps ast
    WHERE ast.phase = 'apply'
    GROUP BY run_id
) AS ast ON ast.run_id = runs.run_id
WHERE runs.run_id = $1
`,
		runID)
	return sql.CollectOneRow(rows, scan)
}

// SetPlanFile writes a plan file to the db
func (db *pgdb) SetPlanFile(ctx context.Context, runID resource.TfeID, file []byte, format PlanFormat) error {
	switch format {
	case PlanFormatBinary:
		_, err := db.Exec(ctx, `UPDATE plans SET plan_bin = $1 WHERE run_id = $2`, file, runID)
		return err
	case PlanFormatJSON:
		_, err := db.Exec(ctx, `UPDATE plans SET plan_json = $1 WHERE run_id = $2`, file, runID)
		return err
	default:
		return fmt.Errorf("unknown plan format: %s", string(format))
	}
}

// GetPlanFile retrieves a plan file for the run
func (db *pgdb) GetPlanFile(ctx context.Context, runID resource.TfeID, format PlanFormat) ([]byte, error) {
	var row pgx.Rows
	switch format {
	case PlanFormatBinary:
		row = db.Query(ctx, `SELECT plan_bin FROM plans WHERE run_id = $1`, runID)
	case PlanFormatJSON:
		row = db.Query(ctx, `SELECT plan_json FROM plans WHERE run_id = $1`, runID)
	default:
		return nil, fmt.Errorf("unknown plan format: %s", string(format))
	}
	return sql.CollectOneType[[]byte](row)
}

// GetLockFile retrieves the lock file for the run
func (db *pgdb) GetLockFile(ctx context.Context, runID resource.TfeID) ([]byte, error) {
	row := db.Query(ctx, `SELECT lock_file FROM runs WHERE run_id = $1`, runID)
	return sql.CollectOneType[[]byte](row)
}

// SetLockFile sets the lock file for the run
func (db *pgdb) SetLockFile(ctx context.Context, runID resource.TfeID, lockFile []byte) error {
	_, err := db.Exec(ctx, `UPDATE runs SET lock_file = $1 WHERE run_id = $2`, lockFile, runID)
	return err
}

// DeleteRun deletes a run from the DB
func (db *pgdb) DeleteRun(ctx context.Context, id resource.TfeID) error {
	_, err := db.Exec(ctx, `DELETE FROM runs WHERE run_id = $1`, id)
	return err
}

func (db *pgdb) insertRunStatusTimestamp(ctx context.Context, run *Run) error {
	ts, err := run.StatusTimestamp(run.Status)
	if err != nil {
		return err
	}
	_, err = db.Exec(ctx, `
INSERT INTO run_status_timestamps (
    run_id,
    status,
    timestamp
) VALUES (
    $1,
    $2,
    $3
)`,
		run.ID,
		run.Status,
		ts,
	)
	return err
}

func (db *pgdb) insertPhaseStatusTimestamp(ctx context.Context, phase Phase) error {
	ts, err := phase.StatusTimestamp(phase.Status)
	if err != nil {
		return err
	}
	_, err = db.Exec(ctx, `
INSERT INTO phase_status_timestamps (
    run_id,
    phase,
    status,
    timestamp
) VALUES (
    $1,
    $2,
    $3,
    $4
)`,
		phase.RunID,
		phase.PhaseType,
		phase.Status,
		ts,
	)
	return err
}

func scan(row pgx.CollectableRow) (*Run, error) {
	type (
		statusTimestampModel struct {
			RunID     resource.TfeID `db:"run_id"`
			Status    runstatus.Status
			Timestamp time.Time
		}
		phaseStatusTimestampModel struct {
			RunID     resource.TfeID `db:"run_id"`
			Phase     internal.PhaseType
			Status    PhaseStatus
			Timestamp time.Time
		}
		runVariableModel struct {
			RunID resource.TfeID `db:"run_id"`
			Key   string
			Value string
		}
		model struct {
			ID                     resource.TfeID                        `db:"run_id"`
			CreatedAt              time.Time                             `db:"created_at"`
			IsDestroy              bool                                  `db:"is_destroy"`
			CancelSignaledAt       *time.Time                            `db:"cancel_signaled_at"`
			Organization           resource.OrganizationName             `db:"organization_name"`
			Refresh                bool                                  `db:"refresh"`
			RefreshOnly            bool                                  `db:"refresh_only"`
			ReplaceAddrs           []string                              `db:"replace_addrs"`
			PositionInQueue        int                                   `db:"position_in_queue"`
			TargetAddrs            []string                              `db:"target_addrs"`
			TerraformVersion       string                                `db:"terraform_version"`
			AllowEmptyApply        bool                                  `db:"allow_empty_apply"`
			AutoApply              bool                                  `db:"auto_apply"`
			PlanOnly               bool                                  `db:"plan_only"`
			Source                 Source                                `db:"source"`
			Status                 runstatus.Status                      `db:"status"`
			PlanStatus             PhaseStatus                           `db:"plan_status"`
			ApplyStatus            PhaseStatus                           `db:"apply_status"`
			WorkspaceID            resource.TfeID                        `db:"workspace_id"`
			ConfigurationVersionID resource.TfeID                        `db:"configuration_version_id"`
			ExecutionMode          workspace.ExecutionMode               `db:"execution_mode"`
			RunVariables           []runVariableModel                    `db:"run_variables"`
			PlanResourceReport     *Report                               `db:"plan_resource_report"`
			PlanOutputReport       *Report                               `db:"plan_output_report"`
			ApplyResourceReport    *Report                               `db:"apply_resource_report"`
			RunStatusTimestamps    []statusTimestampModel                `db:"run_status_timestamps"`
			PlanStatusTimestamps   []phaseStatusTimestampModel           `db:"plan_status_timestamps"`
			ApplyStatusTimestamps  []phaseStatusTimestampModel           `db:"apply_status_timestamps"`
			Latest                 bool                                  `db:"latest"`
			IngressAttributes      *configversion.IngressAttributesModel `db:"ingress_attributes"`
			CreatedBy              *string                               `db:"created_by"`
			CostEstimationEnabled  bool                                  `db:"cost_estimation_enabled"`
		}
	)
	m, err := pgx.RowToStructByName[model](row)
	if err != nil {
		return nil, err
	}
	run := &Run{
		ID:                     m.ID,
		CreatedAt:              m.CreatedAt,
		IsDestroy:              m.IsDestroy,
		CancelSignaledAt:       m.CancelSignaledAt,
		Organization:           m.Organization,
		Refresh:                m.Refresh,
		RefreshOnly:            m.RefreshOnly,
		ReplaceAddrs:           m.ReplaceAddrs,
		PositionInQueue:        m.PositionInQueue,
		TargetAddrs:            m.TargetAddrs,
		TerraformVersion:       m.TerraformVersion,
		AllowEmptyApply:        m.AllowEmptyApply,
		AutoApply:              m.AutoApply,
		PlanOnly:               m.PlanOnly,
		Source:                 m.Source,
		Status:                 m.Status,
		WorkspaceID:            m.WorkspaceID,
		ConfigurationVersionID: m.ConfigurationVersionID,
		ExecutionMode:          m.ExecutionMode,
		Plan: Phase{
			RunID:            m.ID,
			PhaseType:        internal.PlanPhase,
			Status:           m.PlanStatus,
			StatusTimestamps: make([]PhaseStatusTimestamp, len(m.PlanStatusTimestamps)),
			ResourceReport:   m.PlanResourceReport,
			OutputReport:     m.PlanOutputReport,
		},
		Apply: Phase{
			RunID:            m.ID,
			PhaseType:        internal.ApplyPhase,
			Status:           m.ApplyStatus,
			StatusTimestamps: make([]PhaseStatusTimestamp, len(m.ApplyStatusTimestamps)),
			ResourceReport:   m.ApplyResourceReport,
		},
		StatusTimestamps:      make([]StatusTimestamp, len(m.RunStatusTimestamps)),
		Latest:                m.Latest,
		CreatedBy:             m.CreatedBy,
		CostEstimationEnabled: m.CostEstimationEnabled,
	}
	if m.IngressAttributes != nil {
		run.IngressAttributes = m.IngressAttributes.ToIngressAttributes()
	}
	if len(m.RunVariables) > 0 {
		run.Variables = make([]Variable, len(m.RunVariables))
		for i, model := range m.RunVariables {
			run.Variables[i] = Variable{Key: model.Key, Value: model.Value}
		}
	}
	for i, model := range m.RunStatusTimestamps {
		run.StatusTimestamps[i] = StatusTimestamp{
			Status:    model.Status,
			Timestamp: model.Timestamp,
		}
	}
	for i, model := range m.PlanStatusTimestamps {
		run.Plan.StatusTimestamps[i] = PhaseStatusTimestamp{
			Phase:     internal.PlanPhase,
			Status:    model.Status,
			Timestamp: model.Timestamp,
		}
	}
	for i, model := range m.ApplyStatusTimestamps {
		run.Apply.StatusTimestamps[i] = PhaseStatusTimestamp{
			Phase:     internal.ApplyPhase,
			Status:    model.Status,
			Timestamp: model.Timestamp,
		}
	}
	// sort timestamps (earliest first)
	//
	// TODO: use ORDER BY in database queries instead
	sort.Slice(run.StatusTimestamps, func(i, j int) bool {
		return run.StatusTimestamps[i].Timestamp.Before(run.StatusTimestamps[j].Timestamp)
	})
	sort.Slice(run.Plan.StatusTimestamps, func(i, j int) bool {
		return run.Plan.StatusTimestamps[i].Timestamp.Before(run.Plan.StatusTimestamps[j].Timestamp)
	})
	sort.Slice(run.Apply.StatusTimestamps, func(i, j int) bool {
		return run.Apply.StatusTimestamps[i].Timestamp.Before(run.Apply.StatusTimestamps[j].Timestamp)
	})
	return run, err
}
