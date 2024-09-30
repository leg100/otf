package run

import (
	"context"
	"fmt"
	"sort"
	"strconv"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/configversion"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/sql"
	"github.com/leg100/otf/internal/sql/sqlc"
	"github.com/leg100/otf/internal/workspace"
)

type (
	// pgdb is a database of runs on postgres
	pgdb struct {
		*sql.DB // provides access to generated SQL queries
	}

	// pgresult is the result of a database query for a run.
	pgresult struct {
		RunID                  pgtype.Text                  `json:"run_id"`
		CreatedAt              pgtype.Timestamptz           `json:"created_at"`
		CancelSignaledAt       pgtype.Timestamptz           `json:"cancel_signaled_at"`
		IsDestroy              pgtype.Bool                  `json:"is_destroy"`
		PositionInQueue        pgtype.Int4                  `json:"position_in_queue"`
		Refresh                pgtype.Bool                  `json:"refresh"`
		RefreshOnly            pgtype.Bool                  `json:"refresh_only"`
		Source                 pgtype.Text                  `json:"source"`
		Status                 pgtype.Text                  `json:"status"`
		PlanStatus             pgtype.Text                  `json:"plan_status"`
		ApplyStatus            pgtype.Text                  `json:"apply_status"`
		ReplaceAddrs           []string                     `json:"replace_addrs"`
		TargetAddrs            []string                     `json:"target_addrs"`
		AutoApply              pgtype.Bool                  `json:"auto_apply"`
		PlanResourceReport     *sqlc.Report                 `json:"plan_resource_report"`
		PlanOutputReport       *sqlc.Report                 `json:"plan_output_report"`
		ApplyResourceReport    *sqlc.Report                 `json:"apply_resource_report"`
		ConfigurationVersionID pgtype.Text                  `json:"configuration_version_id"`
		WorkspaceID            pgtype.Text                  `json:"workspace_id"`
		PlanOnly               pgtype.Bool                  `json:"plan_only"`
		CreatedBy              pgtype.Text                  `json:"created_by"`
		TerraformVersion       pgtype.Text                  `json:"terraform_version"`
		AllowEmptyApply        pgtype.Bool                  `json:"allow_empty_apply"`
		ExecutionMode          pgtype.Text                  `json:"execution_mode"`
		Latest                 pgtype.Bool                  `json:"latest"`
		OrganizationName       pgtype.Text                  `json:"organization_name"`
		CostEstimationEnabled  pgtype.Bool                  `json:"cost_estimation_enabled"`
		IngressAttributes      *sqlc.IngressAttributes      `json:"ingress_attributes"`
		RunStatusTimestamps    []sqlc.RunStatusTimestamps   `json:"run_status_timestamps"`
		PlanStatusTimestamps   []sqlc.PhaseStatusTimestamps `json:"plan_status_timestamps"`
		ApplyStatusTimestamps  []sqlc.PhaseStatusTimestamps `json:"apply_status_timestamps"`
		RunVariables           []sqlc.RunVariables          `json:"run_variables"`
	}
)

func (result pgresult) toRun() *Run {
	run := Run{
		ID:                     result.RunID.String,
		CreatedAt:              result.CreatedAt.Time.UTC(),
		IsDestroy:              result.IsDestroy.Bool,
		PositionInQueue:        int(result.PositionInQueue.Int),
		Refresh:                result.Refresh.Bool,
		RefreshOnly:            result.RefreshOnly.Bool,
		Source:                 Source(result.Source.String),
		Status:                 Status(result.Status.String),
		ReplaceAddrs:           result.ReplaceAddrs,
		TargetAddrs:            result.TargetAddrs,
		AutoApply:              result.AutoApply.Bool,
		PlanOnly:               result.PlanOnly.Bool,
		AllowEmptyApply:        result.AllowEmptyApply.Bool,
		TerraformVersion:       result.TerraformVersion.String,
		ExecutionMode:          workspace.ExecutionMode(result.ExecutionMode.String),
		Latest:                 result.Latest.Bool,
		Organization:           result.OrganizationName.String,
		WorkspaceID:            result.WorkspaceID.String,
		ConfigurationVersionID: result.ConfigurationVersionID.String,
		CostEstimationEnabled:  result.CostEstimationEnabled.Bool,
		Plan: Phase{
			RunID:          result.RunID.String,
			PhaseType:      internal.PlanPhase,
			Status:         PhaseStatus(result.PlanStatus.String),
			ResourceReport: reportFromDB(result.PlanResourceReport),
			OutputReport:   reportFromDB(result.PlanOutputReport),
		},
		Apply: Phase{
			RunID:          result.RunID.String,
			PhaseType:      internal.ApplyPhase,
			Status:         PhaseStatus(result.ApplyStatus.String),
			ResourceReport: reportFromDB(result.ApplyResourceReport),
		},
	}
	// convert run timestamps from db result and sort them according to
	// timestamp (earliest first)
	run.StatusTimestamps = make([]StatusTimestamp, len(result.RunStatusTimestamps))
	for i, rst := range result.RunStatusTimestamps {
		run.StatusTimestamps[i] = StatusTimestamp{
			Status:    Status(rst.Status.String),
			Timestamp: rst.Timestamp.Time.UTC(),
		}
	}
	sort.Slice(run.StatusTimestamps, func(i, j int) bool {
		return run.StatusTimestamps[i].Timestamp.Before(run.StatusTimestamps[j].Timestamp)
	})
	// convert plan timestamps from db result and sort them according to
	// timestamp (earliest first)
	run.Plan.StatusTimestamps = make([]PhaseStatusTimestamp, len(result.PlanStatusTimestamps))
	for i, pst := range result.PlanStatusTimestamps {
		run.Plan.StatusTimestamps[i] = PhaseStatusTimestamp{
			Status:    PhaseStatus(pst.Status.String),
			Timestamp: pst.Timestamp.Time.UTC(),
		}
	}
	sort.Slice(run.Plan.StatusTimestamps, func(i, j int) bool {
		return run.Plan.StatusTimestamps[i].Timestamp.Before(run.Plan.StatusTimestamps[j].Timestamp)
	})
	// convert apply timestamps from db result and sort them according to
	// timestamp (earliest first)
	run.Apply.StatusTimestamps = make([]PhaseStatusTimestamp, len(result.ApplyStatusTimestamps))
	for i, ast := range result.ApplyStatusTimestamps {
		run.Apply.StatusTimestamps[i] = PhaseStatusTimestamp{
			Status:    PhaseStatus(ast.Status.String),
			Timestamp: ast.Timestamp.Time.UTC(),
		}
	}
	sort.Slice(run.Apply.StatusTimestamps, func(i, j int) bool {
		return run.Apply.StatusTimestamps[i].Timestamp.Before(run.Apply.StatusTimestamps[j].Timestamp)
	})
	if len(result.RunVariables) > 0 {
		run.Variables = make([]Variable, len(result.RunVariables))
		for i, v := range result.RunVariables {
			run.Variables[i] = Variable{Key: v.Key.String, Value: v.Value.String}
		}
	}
	if result.CreatedBy.Status == pgtype.Present {
		run.CreatedBy = &result.CreatedBy.String
	}
	if result.CancelSignaledAt.Status == pgtype.Present {
		run.CancelSignaledAt = internal.Time(result.CancelSignaledAt.Time.UTC())
	}
	if result.IngressAttributes != nil {
		run.IngressAttributes = configversion.NewIngressFromRow(result.IngressAttributes)
	}
	return &run
}

// CreateRun persists a Run to the DB.
func (db *pgdb) CreateRun(ctx context.Context, run *Run) error {
	return db.Tx(ctx, func(ctx context.Context, q *sqlc.Queries) error {
		_, err := q.InsertRun(ctx, sqlc.InsertRunParams{
			ID:                     sql.String(run.ID),
			CreatedAt:              sql.Timestamptz(run.CreatedAt),
			IsDestroy:              sql.Bool(run.IsDestroy),
			PositionInQueue:        sql.Int4(0),
			Refresh:                sql.Bool(run.Refresh),
			RefreshOnly:            sql.Bool(run.RefreshOnly),
			Source:                 sql.String(string(run.Source)),
			Status:                 sql.String(string(run.Status)),
			ReplaceAddrs:           run.ReplaceAddrs,
			TargetAddrs:            run.TargetAddrs,
			AutoApply:              sql.Bool(run.AutoApply),
			PlanOnly:               sql.Bool(run.PlanOnly),
			AllowEmptyApply:        sql.Bool(run.AllowEmptyApply),
			TerraformVersion:       sql.String(run.TerraformVersion),
			ConfigurationVersionID: sql.String(run.ConfigurationVersionID),
			WorkspaceID:            sql.String(run.WorkspaceID),
			CreatedBy:              sql.StringPtr(run.CreatedBy),
		})
		for _, v := range run.Variables {
			_, err = q.InsertRunVariable(ctx, sqlc.InsertRunVariableParams{
				RunID: sql.String(run.ID),
				Key:   sql.String(v.Key),
				Value: sql.String(v.Value),
			})
			if err != nil {
				return fmt.Errorf("inserting run variable: %w", err)
			}
		}
		if err != nil {
			return fmt.Errorf("inserting run: %w", err)
		}
		_, err = q.InsertPlan(ctx, sql.String(run.ID), sql.String(string(run.Plan.Status)))
		if err != nil {
			return fmt.Errorf("inserting plan: %w", err)
		}
		_, err = q.InsertApply(ctx, sql.String(run.ID), sql.String(string(run.Apply.Status)))
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
func (db *pgdb) UpdateStatus(ctx context.Context, runID string, fn func(*Run) error) (*Run, error) {
	var run *Run
	err := db.Tx(ctx, func(ctx context.Context, q *sqlc.Queries) error {
		// select ...for update
		result, err := q.FindRunByIDForUpdate(ctx, sql.String(runID))
		if err != nil {
			return sql.Error(err)
		}
		run = pgresult(result).toRun()

		// Make copies of run attributes before update
		runStatus := run.Status
		planStatus := run.Plan.Status
		applyStatus := run.Apply.Status
		cancelSignaledAt := run.CancelSignaledAt

		if err := fn(run); err != nil {
			return err
		}

		if run.Status != runStatus {
			_, err := q.UpdateRunStatus(ctx, sql.String(string(run.Status)), sql.String(run.ID))
			if err != nil {
				return err
			}

			if err := db.insertRunStatusTimestamp(ctx, run); err != nil {
				return err
			}
		}

		if run.Plan.Status != planStatus {
			_, err := q.UpdatePlanStatusByID(ctx, sql.String(string(run.Plan.Status)), sql.String(run.ID))
			if err != nil {
				return err
			}

			if err := db.insertPhaseStatusTimestamp(ctx, run.Plan); err != nil {
				return err
			}
		}

		if run.Apply.Status != applyStatus {
			_, err := q.UpdateApplyStatusByID(ctx, sql.String(string(run.Apply.Status)), sql.String(run.ID))
			if err != nil {
				return err
			}

			if err := db.insertPhaseStatusTimestamp(ctx, run.Apply); err != nil {
				return err
			}
		}

		if run.CancelSignaledAt != cancelSignaledAt && run.CancelSignaledAt != nil {
			_, err := q.UpdateCancelSignaledAt(ctx, sql.Timestamptz(*run.CancelSignaledAt), sql.String(run.ID))
			if err != nil {
				return err
			}
		}

		return nil
	})
	return run, err
}

func (db *pgdb) CreatePlanReport(ctx context.Context, runID string, resource, output Report) error {
	_, err := db.Conn(ctx).UpdatePlannedChangesByID(ctx, sqlc.UpdatePlannedChangesByIDParams{
		RunID:                sql.String(runID),
		ResourceAdditions:    sql.Int4(resource.Additions),
		ResourceChanges:      sql.Int4(resource.Changes),
		ResourceDestructions: sql.Int4(resource.Destructions),
		OutputAdditions:      sql.Int4(output.Additions),
		OutputChanges:        sql.Int4(output.Changes),
		OutputDestructions:   sql.Int4(output.Destructions),
	})
	if err != nil {
		return sql.Error(err)
	}
	return err
}

func (db *pgdb) CreateApplyReport(ctx context.Context, runID string, report Report) error {
	_, err := db.Conn(ctx).UpdateAppliedChangesByID(ctx, sqlc.UpdateAppliedChangesByIDParams{
		RunID:        sql.String(runID),
		Additions:    sql.Int4(report.Additions),
		Changes:      sql.Int4(report.Changes),
		Destructions: sql.Int4(report.Destructions),
	})
	if err != nil {
		return sql.Error(err)
	}
	return err
}

func (db *pgdb) ListRuns(ctx context.Context, opts ListOptions) (*resource.Page[*Run], error) {
	q := db.Conn(ctx)
	batch := &pgx.Batch{}
	organization := "%"
	if opts.Organization != nil {
		organization = *opts.Organization
	}
	workspaceName := "%"
	if opts.WorkspaceName != nil {
		workspaceName = *opts.WorkspaceName
	}
	workspaceID := "%"
	if opts.WorkspaceID != nil {
		workspaceID = *opts.WorkspaceID
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
	q.FindRunsBatch(batch, sqlc.FindRunsParams{
		OrganizationNames: []string{organization},
		WorkspaceNames:    []string{workspaceName},
		WorkspaceIds:      []string{workspaceID},
		CommitSHA:         sql.StringPtr(opts.CommitSHA),
		VCSUsername:       sql.StringPtr(opts.VCSUsername),
		Sources:           sources,
		Statuses:          statuses,
		PlanOnly:          []string{planOnly},
		Limit:             opts.GetLimit(),
		Offset:            opts.GetOffset(),
	})
	q.CountRunsBatch(batch, sqlc.CountRunsParams{
		OrganizationNames: []string{organization},
		WorkspaceNames:    []string{workspaceName},
		WorkspaceIds:      []string{workspaceID},
		CommitSHA:         sql.StringPtr(opts.CommitSHA),
		VCSUsername:       sql.StringPtr(opts.VCSUsername),
		Sources:           sources,
		Statuses:          statuses,
		PlanOnly:          []string{planOnly},
	})

	results := db.SendBatch(ctx, batch)
	defer results.Close()

	rows, err := q.FindRunsScan(results)
	if err != nil {
		return nil, err
	}
	count, err := q.CountRunsScan(results)
	if err != nil {
		return nil, err
	}

	items := make([]*Run, len(rows))
	for i, r := range rows {
		items[i] = pgresult(r).toRun()
	}
	return resource.NewPage(items, opts.PageOptions, internal.Int64(count.Int)), nil
}

// GetRun retrieves a run using the get options
func (db *pgdb) GetRun(ctx context.Context, runID string) (*Run, error) {
	result, err := db.Conn(ctx).FindRunByID(ctx, sql.String(runID))
	if err != nil {
		return nil, sql.Error(err)
	}
	return pgresult(result).toRun(), nil
}

// SetPlanFile writes a plan file to the db
func (db *pgdb) SetPlanFile(ctx context.Context, runID string, file []byte, format PlanFormat) error {
	q := db.Conn(ctx)
	switch format {
	case PlanFormatBinary:
		_, err := q.UpdatePlanBinByID(ctx, file, sql.String(runID))
		return err
	case PlanFormatJSON:
		_, err := q.UpdatePlanJSONByID(ctx, file, sql.String(runID))
		return err
	default:
		return fmt.Errorf("unknown plan format: %s", string(format))
	}
}

// GetPlanFile retrieves a plan file for the run
func (db *pgdb) GetPlanFile(ctx context.Context, runID string, format PlanFormat) ([]byte, error) {
	q := db.Conn(ctx)
	switch format {
	case PlanFormatBinary:
		return q.GetPlanBinByID(ctx, sql.String(runID))
	case PlanFormatJSON:
		return q.GetPlanJSONByID(ctx, sql.String(runID))
	default:
		return nil, fmt.Errorf("unknown plan format: %s", string(format))
	}
}

// GetLockFile retrieves the lock file for the run
func (db *pgdb) GetLockFile(ctx context.Context, runID string) ([]byte, error) {
	return db.Conn(ctx).GetLockFileByID(ctx, sql.String(runID))
}

// SetLockFile sets the lock file for the run
func (db *pgdb) SetLockFile(ctx context.Context, runID string, lockFile []byte) error {
	_, err := db.Conn(ctx).PutLockFile(ctx, lockFile, sql.String(runID))
	return err
}

// DeleteRun deletes a run from the DB
func (db *pgdb) DeleteRun(ctx context.Context, id string) error {
	_, err := db.Conn(ctx).DeleteRunByID(ctx, sql.String(id))
	return err
}

func (db *pgdb) insertRunStatusTimestamp(ctx context.Context, run *Run) error {
	ts, err := run.StatusTimestamp(run.Status)
	if err != nil {
		return err
	}
	_, err = db.Conn(ctx).InsertRunStatusTimestamp(ctx, sqlc.InsertRunStatusTimestampParams{
		ID:        sql.String(run.ID),
		Status:    sql.String(string(run.Status)),
		Timestamp: sql.Timestamptz(ts),
	})
	return err
}

func (db *pgdb) insertPhaseStatusTimestamp(ctx context.Context, phase Phase) error {
	ts, err := phase.StatusTimestamp(phase.Status)
	if err != nil {
		return err
	}
	_, err = db.Conn(ctx).InsertPhaseStatusTimestamp(ctx, sqlc.InsertPhaseStatusTimestampParams{
		RunID:     sql.String(phase.RunID),
		Phase:     sql.String(string(phase.PhaseType)),
		Status:    sql.String(string(phase.Status)),
		Timestamp: sql.Timestamptz(ts),
	})
	return err
}
