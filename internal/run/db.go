package run

import (
	"context"
	"fmt"
	"strconv"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/configversion"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/sql"
	"github.com/leg100/otf/internal/sql/pggen"
	"github.com/leg100/otf/internal/workspace"
)

type (
	// pgdb is a database of runs on postgres
	pgdb struct {
		*sql.DB // provides access to generated SQL queries
	}

	// pgresult is the result of a database query for a run.
	pgresult struct {
		RunID                  pgtype.Text                   `json:"run_id"`
		CreatedAt              pgtype.Timestamptz            `json:"created_at"`
		ForceCancelAvailableAt pgtype.Timestamptz            `json:"force_cancel_available_at"`
		IsDestroy              bool                          `json:"is_destroy"`
		PositionInQueue        pgtype.Int4                   `json:"position_in_queue"`
		Refresh                bool                          `json:"refresh"`
		RefreshOnly            bool                          `json:"refresh_only"`
		Source                 pgtype.Text                   `json:"source"`
		Status                 pgtype.Text                   `json:"status"`
		PlanStatus             pgtype.Text                   `json:"plan_status"`
		ApplyStatus            pgtype.Text                   `json:"apply_status"`
		ReplaceAddrs           []string                      `json:"replace_addrs"`
		TargetAddrs            []string                      `json:"target_addrs"`
		AutoApply              bool                          `json:"auto_apply"`
		PlanResourceReport     *pggen.Report                 `json:"plan_resource_report"`
		PlanOutputReport       *pggen.Report                 `json:"plan_output_report"`
		ApplyResourceReport    *pggen.Report                 `json:"apply_resource_report"`
		ConfigurationVersionID pgtype.Text                   `json:"configuration_version_id"`
		WorkspaceID            pgtype.Text                   `json:"workspace_id"`
		PlanOnly               bool                          `json:"plan_only"`
		CreatedBy              pgtype.Text                   `json:"created_by"`
		ExecutionMode          pgtype.Text                   `json:"execution_mode"`
		Latest                 bool                          `json:"latest"`
		OrganizationName       pgtype.Text                   `json:"organization_name"`
		CostEstimationEnabled  bool                          `json:"cost_estimation_enabled"`
		IngressAttributes      *pggen.IngressAttributes      `json:"ingress_attributes"`
		RunStatusTimestamps    []pggen.RunStatusTimestamps   `json:"run_status_timestamps"`
		PlanStatusTimestamps   []pggen.PhaseStatusTimestamps `json:"plan_status_timestamps"`
		ApplyStatusTimestamps  []pggen.PhaseStatusTimestamps `json:"apply_status_timestamps"`
	}
)

func (result pgresult) toRun() *Run {
	run := Run{
		ID:                     result.RunID.String,
		CreatedAt:              result.CreatedAt.Time.UTC(),
		IsDestroy:              result.IsDestroy,
		PositionInQueue:        int(result.PositionInQueue.Int),
		Refresh:                result.Refresh,
		RefreshOnly:            result.RefreshOnly,
		Source:                 RunSource(result.Source.String),
		Status:                 internal.RunStatus(result.Status.String),
		StatusTimestamps:       unmarshalRunStatusTimestampRows(result.RunStatusTimestamps),
		ReplaceAddrs:           result.ReplaceAddrs,
		TargetAddrs:            result.TargetAddrs,
		AutoApply:              result.AutoApply,
		PlanOnly:               result.PlanOnly,
		ExecutionMode:          workspace.ExecutionMode(result.ExecutionMode.String),
		Latest:                 result.Latest,
		Organization:           result.OrganizationName.String,
		WorkspaceID:            result.WorkspaceID.String,
		ConfigurationVersionID: result.ConfigurationVersionID.String,
		CostEstimationEnabled:  result.CostEstimationEnabled,
		Plan: Phase{
			RunID:            result.RunID.String,
			PhaseType:        internal.PlanPhase,
			Status:           PhaseStatus(result.PlanStatus.String),
			StatusTimestamps: unmarshalPlanStatusTimestampRows(result.PlanStatusTimestamps),
			ResourceReport:   reportFromDB(result.PlanResourceReport),
			OutputReport:     reportFromDB(result.PlanOutputReport),
		},
		Apply: Phase{
			RunID:            result.RunID.String,
			PhaseType:        internal.ApplyPhase,
			Status:           PhaseStatus(result.ApplyStatus.String),
			StatusTimestamps: unmarshalApplyStatusTimestampRows(result.ApplyStatusTimestamps),
			ResourceReport:   reportFromDB(result.ApplyResourceReport),
		},
	}
	if result.CreatedBy.Status == pgtype.Present {
		run.CreatedBy = &result.CreatedBy.String
	}
	if result.ForceCancelAvailableAt.Status == pgtype.Present {
		run.ForceCancelAvailableAt = internal.Time(result.ForceCancelAvailableAt.Time.UTC())
	}
	if result.IngressAttributes != nil {
		run.IngressAttributes = configversion.NewIngressFromRow(result.IngressAttributes)
	}
	return &run
}

// CreateRun persists a Run to the DB.
func (db *pgdb) CreateRun(ctx context.Context, run *Run) error {
	return db.Tx(ctx, func(ctx context.Context, q pggen.Querier) error {
		_, err := q.InsertRun(ctx, pggen.InsertRunParams{
			ID:                     sql.String(run.ID),
			CreatedAt:              sql.Timestamptz(run.CreatedAt),
			IsDestroy:              run.IsDestroy,
			PositionInQueue:        sql.Int4(0),
			Refresh:                run.Refresh,
			RefreshOnly:            run.RefreshOnly,
			Source:                 sql.String(string(run.Source)),
			Status:                 sql.String(string(run.Status)),
			ReplaceAddrs:           run.ReplaceAddrs,
			TargetAddrs:            run.TargetAddrs,
			AutoApply:              run.AutoApply,
			PlanOnly:               run.PlanOnly,
			ConfigurationVersionID: sql.String(run.ConfigurationVersionID),
			WorkspaceID:            sql.String(run.WorkspaceID),
			CreatedBy:              sql.StringPtr(run.CreatedBy),
		})
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
	err := db.Tx(ctx, func(ctx context.Context, q pggen.Querier) error {
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
		forceCancelAvailableAt := run.ForceCancelAvailableAt

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

		if run.ForceCancelAvailableAt != forceCancelAvailableAt && run.ForceCancelAvailableAt != nil {
			_, err := q.UpdateRunForceCancelAvailableAt(ctx, sql.Timestamptz(*run.ForceCancelAvailableAt), sql.String(run.ID))
			if err != nil {
				return err
			}
		}

		return nil
	})
	return run, err
}

func (db *pgdb) CreatePlanReport(ctx context.Context, runID string, resource, output Report) error {
	_, err := db.Conn(ctx).UpdatePlannedChangesByID(ctx, pggen.UpdatePlannedChangesByIDParams{
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
	_, err := db.Conn(ctx).UpdateAppliedChangesByID(ctx, pggen.UpdateAppliedChangesByIDParams{
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
	q.FindRunsBatch(batch, pggen.FindRunsParams{
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
	q.CountRunsBatch(batch, pggen.CountRunsParams{
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

	var items []*Run
	for _, r := range rows {
		items = append(items, pgresult(r).toRun())
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
	_, err = db.Conn(ctx).InsertRunStatusTimestamp(ctx, pggen.InsertRunStatusTimestampParams{
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
	_, err = db.Conn(ctx).InsertPhaseStatusTimestamp(ctx, pggen.InsertPhaseStatusTimestampParams{
		RunID:     sql.String(phase.RunID),
		Phase:     sql.String(string(phase.PhaseType)),
		Status:    sql.String(string(phase.Status)),
		Timestamp: sql.Timestamptz(ts),
	})
	return err
}

func unmarshalRunStatusTimestampRows(rows []pggen.RunStatusTimestamps) (timestamps []RunStatusTimestamp) {
	for _, ty := range rows {
		timestamps = append(timestamps, RunStatusTimestamp{
			Status:    internal.RunStatus(ty.Status.String),
			Timestamp: ty.Timestamp.Time.UTC(),
		})
	}
	return timestamps
}

func unmarshalPlanStatusTimestampRows(rows []pggen.PhaseStatusTimestamps) (timestamps []PhaseStatusTimestamp) {
	for _, ty := range rows {
		timestamps = append(timestamps, PhaseStatusTimestamp{
			Status:    PhaseStatus(ty.Status.String),
			Timestamp: ty.Timestamp.Time.UTC(),
		})
	}
	return timestamps
}

func unmarshalApplyStatusTimestampRows(rows []pggen.PhaseStatusTimestamps) (timestamps []PhaseStatusTimestamp) {
	for _, ty := range rows {
		timestamps = append(timestamps, PhaseStatusTimestamp{
			Status:    PhaseStatus(ty.Status.String),
			Timestamp: ty.Timestamp.Time.UTC(),
		})
	}
	return timestamps
}
