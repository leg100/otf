package run

import (
	"context"
	"fmt"
	"strconv"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/leg100/otf"
	"github.com/leg100/otf/sql"
	"github.com/leg100/otf/sql/pggen"
	"github.com/leg100/otf/workspace"
)

// pgdb is a database of runs on postgres
type pgdb struct {
	otf.DB // provides access to generated SQL queries
}

func newDB(db otf.DB) *pgdb {
	return &pgdb{db}
}

// CreateRun persists a Run to the DB.
func (db *pgdb) CreateRun(ctx context.Context, run *Run) error {
	return db.tx(ctx, func(tx *pgdb) error {
		_, err := tx.InsertRun(ctx, pggen.InsertRunParams{
			ID:                     sql.String(run.ID),
			CreatedAt:              sql.Timestamptz(run.CreatedAt),
			IsDestroy:              run.IsDestroy,
			Refresh:                run.Refresh,
			RefreshOnly:            run.RefreshOnly,
			Status:                 sql.String(string(run.Status)),
			ReplaceAddrs:           run.ReplaceAddrs,
			TargetAddrs:            run.TargetAddrs,
			AutoApply:              run.AutoApply,
			ConfigurationVersionID: sql.String(run.ConfigurationVersionID),
			WorkspaceID:            sql.String(run.WorkspaceID),
		})
		if err != nil {
			return fmt.Errorf("inserting run: %w", err)
		}
		_, err = tx.InsertPlan(ctx, sql.String(run.ID), sql.String(string(run.Plan.Status)))
		if err != nil {
			return fmt.Errorf("inserting plan: %w", err)
		}
		_, err = tx.InsertApply(ctx, sql.String(run.ID), sql.String(string(run.Apply.Status)))
		if err != nil {
			return fmt.Errorf("inserting apply: %w", err)
		}
		if err := tx.insertRunStatusTimestamp(ctx, run); err != nil {
			return fmt.Errorf("inserting run status timestamp: %w", err)
		}
		if err := tx.insertPhaseStatusTimestamp(ctx, run.Plan); err != nil {
			return fmt.Errorf("inserting plan status timestamp: %w", err)
		}
		if err := tx.insertPhaseStatusTimestamp(ctx, run.Apply); err != nil {
			return fmt.Errorf("inserting apply status timestamp: %w", err)
		}
		return nil
	})
}

// UpdateStatus updates the run status as well as its plan and/or apply.
func (db *pgdb) UpdateStatus(ctx context.Context, runID string, fn func(*Run) error) (*Run, error) {
	var run *Run
	err := db.tx(ctx, func(tx *pgdb) error {
		// select ...for update
		result, err := tx.FindRunByIDForUpdate(ctx, sql.String(runID))
		if err != nil {
			return sql.Error(err)
		}
		run, err = UnmarshalRunResult(RunResult(result))
		if err != nil {
			return err
		}

		// Make copies of run attributes before update
		runStatus := run.Status
		planStatus := run.Plan.Status
		applyStatus := run.Apply.Status
		forceCancelAvailableAt := run.ForceCancelAvailableAt

		if err := fn(run); err != nil {
			return err
		}

		if run.Status != runStatus {
			_, err := tx.UpdateRunStatus(ctx, sql.String(string(run.Status)), sql.String(run.ID))
			if err != nil {
				return err
			}

			if err := tx.insertRunStatusTimestamp(ctx, run); err != nil {
				return err
			}
		}

		if run.Plan.Status != planStatus {
			_, err := tx.UpdatePlanStatusByID(ctx, sql.String(string(run.Plan.Status)), sql.String(run.ID))
			if err != nil {
				return err
			}

			if err := tx.insertPhaseStatusTimestamp(ctx, run.Plan); err != nil {
				return err
			}
		}

		if run.Apply.Status != applyStatus {
			_, err := tx.UpdateApplyStatusByID(ctx, sql.String(string(run.Apply.Status)), sql.String(run.ID))
			if err != nil {
				return err
			}

			if err := tx.insertPhaseStatusTimestamp(ctx, run.Apply); err != nil {
				return err
			}
		}

		if run.ForceCancelAvailableAt != forceCancelAvailableAt && run.ForceCancelAvailableAt != nil {
			_, err := tx.UpdateRunForceCancelAvailableAt(ctx, sql.Timestamptz(*run.ForceCancelAvailableAt), sql.String(run.ID))
			if err != nil {
				return err
			}
		}

		return nil
	})
	return run, err
}

func (db *pgdb) CreatePlanReport(ctx context.Context, runID string, report ResourceReport) error {
	_, err := db.UpdatePlannedChangesByID(ctx, pggen.UpdatePlannedChangesByIDParams{
		RunID:        sql.String(runID),
		Additions:    report.Additions,
		Changes:      report.Changes,
		Destructions: report.Destructions,
	})
	if err != nil {
		return sql.Error(err)
	}
	return err
}

func (db *pgdb) CreateApplyReport(ctx context.Context, runID string, report ResourceReport) error {
	_, err := db.UpdateAppliedChangesByID(ctx, pggen.UpdateAppliedChangesByIDParams{
		RunID:        sql.String(runID),
		Additions:    report.Additions,
		Changes:      report.Changes,
		Destructions: report.Destructions,
	})
	if err != nil {
		return sql.Error(err)
	}
	return err
}

func (db *pgdb) ListRuns(ctx context.Context, opts RunListOptions) (*RunList, error) {
	batch := &pgx.Batch{}
	organizationName := "%"
	if opts.Organization != nil {
		organizationName = *opts.Organization
	}
	workspaceName := "%"
	if opts.WorkspaceName != nil {
		workspaceName = *opts.WorkspaceName
	}
	workspaceID := "%"
	if opts.WorkspaceID != nil {
		workspaceID = *opts.WorkspaceID
	}
	statuses := []string{"%"}
	if len(opts.Statuses) > 0 {
		statuses = convertStatusSliceToStringSlice(opts.Statuses)
	}
	speculative := "%"
	if opts.Speculative != nil {
		speculative = strconv.FormatBool(*opts.Speculative)
	}
	db.FindRunsBatch(batch, pggen.FindRunsParams{
		OrganizationNames: []string{organizationName},
		WorkspaceNames:    []string{workspaceName},
		WorkspaceIds:      []string{workspaceID},
		Statuses:          statuses,
		Speculative:       []string{speculative},
		Limit:             opts.GetLimit(),
		Offset:            opts.GetOffset(),
	})
	db.CountRunsBatch(batch, pggen.CountRunsParams{
		OrganizationNames: []string{organizationName},
		WorkspaceNames:    []string{workspaceName},
		WorkspaceIds:      []string{workspaceID},
		Statuses:          statuses,
		Speculative:       []string{speculative},
	})

	results := db.SendBatch(ctx, batch)
	defer results.Close()

	rows, err := db.FindRunsScan(results)
	if err != nil {
		return nil, err
	}
	count, err := db.CountRunsScan(results)
	if err != nil {
		return nil, err
	}

	var items []*Run
	for _, r := range rows {
		run, err := UnmarshalRunResult(RunResult(r))
		if err != nil {
			return nil, err
		}
		items = append(items, run)
	}

	return &RunList{
		Items:      items,
		Pagination: otf.NewPagination(opts.ListOptions, *count),
	}, nil
}

// GetRun retrieves a run using the get options
func (db *pgdb) GetRun(ctx context.Context, runID string) (*Run, error) {
	result, err := db.FindRunByID(ctx, sql.String(runID))
	if err != nil {
		return nil, sql.Error(err)
	}
	return UnmarshalRunResult(RunResult(result))
}

// SetPlanFile writes a plan file to the db
func (db *pgdb) SetPlanFile(ctx context.Context, runID string, file []byte, format PlanFormat) error {
	switch format {
	case PlanFormatBinary:
		_, err := db.UpdatePlanBinByID(ctx, file, sql.String(runID))
		return err
	case PlanFormatJSON:
		_, err := db.UpdatePlanJSONByID(ctx, file, sql.String(runID))
		return err
	default:
		return fmt.Errorf("unknown plan format: %s", string(format))
	}
}

// GetPlanFile retrieves a plan file for the run
func (db *pgdb) GetPlanFile(ctx context.Context, runID string, format PlanFormat) ([]byte, error) {
	switch format {
	case PlanFormatBinary:
		return db.GetPlanBinByID(ctx, sql.String(runID))
	case PlanFormatJSON:
		return db.GetPlanJSONByID(ctx, sql.String(runID))
	default:
		return nil, fmt.Errorf("unknown plan format: %s", string(format))
	}
}

// GetLockFile retrieves the lock file for the run
func (db *pgdb) GetLockFile(ctx context.Context, runID string) ([]byte, error) {
	return db.GetLockFileByID(ctx, sql.String(runID))
}

// SetLockFile sets the lock file for the run
func (db *pgdb) SetLockFile(ctx context.Context, runID string, lockFile []byte) error {
	_, err := db.PutLockFile(ctx, lockFile, sql.String(runID))
	return err
}

// DeleteRun deletes a run from the DB
func (db *pgdb) DeleteRun(ctx context.Context, id string) error {
	_, err := db.DeleteRunByID(ctx, sql.String(id))
	return err
}

func (db *pgdb) insertRunStatusTimestamp(ctx context.Context, run *Run) error {
	ts, err := run.StatusTimestamp(run.Status)
	if err != nil {
		return err
	}
	_, err = db.InsertRunStatusTimestamp(ctx, pggen.InsertRunStatusTimestampParams{
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
	_, err = db.InsertPhaseStatusTimestamp(ctx, pggen.InsertPhaseStatusTimestampParams{
		RunID:     sql.String(phase.RunID),
		Phase:     sql.String(string(phase.PhaseType)),
		Status:    sql.String(string(phase.Status)),
		Timestamp: sql.Timestamptz(ts),
	})
	return err
}

// tx constructs a new pgdb within a transaction.
func (db *pgdb) tx(ctx context.Context, callback func(*pgdb) error) error {
	return db.Tx(ctx, func(tx otf.DB) error {
		return callback(newDB(tx))
	})
}

func convertStatusSliceToStringSlice(statuses []otf.RunStatus) (s []string) {
	for _, status := range statuses {
		s = append(s, string(status))
	}
	return
}

// RunResult represents the result of a database query for a run.
type RunResult struct {
	RunID                  pgtype.Text                   `json:"run_id"`
	CreatedAt              pgtype.Timestamptz            `json:"created_at"`
	ForceCancelAvailableAt pgtype.Timestamptz            `json:"force_cancel_available_at"`
	IsDestroy              bool                          `json:"is_destroy"`
	PositionInQueue        int                           `json:"position_in_queue"`
	Refresh                bool                          `json:"refresh"`
	RefreshOnly            bool                          `json:"refresh_only"`
	Status                 pgtype.Text                   `json:"status"`
	PlanStatus             pgtype.Text                   `json:"plan_status"`
	ApplyStatus            pgtype.Text                   `json:"apply_status"`
	ReplaceAddrs           []string                      `json:"replace_addrs"`
	TargetAddrs            []string                      `json:"target_addrs"`
	AutoApply              bool                          `json:"auto_apply"`
	PlannedChanges         *pggen.Report                 `json:"planned_changes"`
	AppliedChanges         *pggen.Report                 `json:"applied_changes"`
	ConfigurationVersionID pgtype.Text                   `json:"configuration_version_id"`
	WorkspaceID            pgtype.Text                   `json:"workspace_id"`
	Speculative            bool                          `json:"speculative"`
	ExecutionMode          pgtype.Text                   `json:"execution_mode"`
	Latest                 bool                          `json:"latest"`
	OrganizationName       pgtype.Text                   `json:"organization_name"`
	IngressAttributes      *pggen.IngressAttributes      `json:"ingress_attributes"`
	RunStatusTimestamps    []pggen.RunStatusTimestamps   `json:"run_status_timestamps"`
	PlanStatusTimestamps   []pggen.PhaseStatusTimestamps `json:"plan_status_timestamps"`
	ApplyStatusTimestamps  []pggen.PhaseStatusTimestamps `json:"apply_status_timestamps"`
}

func UnmarshalRunResult(result RunResult) (*Run, error) {
	run := Run{
		ID:                     result.RunID.String,
		CreatedAt:              result.CreatedAt.Time.UTC(),
		IsDestroy:              result.IsDestroy,
		PositionInQueue:        result.PositionInQueue,
		Refresh:                result.Refresh,
		RefreshOnly:            result.RefreshOnly,
		Status:                 otf.RunStatus(result.Status.String),
		StatusTimestamps:       unmarshalRunStatusTimestampRows(result.RunStatusTimestamps),
		ReplaceAddrs:           result.ReplaceAddrs,
		TargetAddrs:            result.TargetAddrs,
		AutoApply:              result.AutoApply,
		Speculative:            result.Speculative,
		ExecutionMode:          workspace.ExecutionMode(result.ExecutionMode.String),
		Latest:                 result.Latest,
		Organization:           result.OrganizationName.String,
		WorkspaceID:            result.WorkspaceID.String,
		ConfigurationVersionID: result.ConfigurationVersionID.String,
		Plan: Phase{
			RunID:            result.RunID.String,
			Status:           PhaseStatus(result.PlanStatus.String),
			StatusTimestamps: unmarshalPlanStatusTimestampRows(result.PlanStatusTimestamps),
			ResourceReport:   (*ResourceReport)(result.PlannedChanges),
		},
		Apply: Phase{
			RunID:            result.RunID.String,
			Status:           PhaseStatus(result.ApplyStatus.String),
			StatusTimestamps: unmarshalApplyStatusTimestampRows(result.ApplyStatusTimestamps),
			ResourceReport:   (*ResourceReport)(result.AppliedChanges),
		},
	}
	if result.ForceCancelAvailableAt.Status == pgtype.Present {
		run.ForceCancelAvailableAt = otf.Time(result.ForceCancelAvailableAt.Time.UTC())
	}
	if result.IngressAttributes != nil {
		run.Commit = &result.IngressAttributes.CommitSHA.String
	}
	return &run, nil
}

func unmarshalRunStatusTimestampRows(rows []pggen.RunStatusTimestamps) (timestamps []RunStatusTimestamp) {
	for _, ty := range rows {
		timestamps = append(timestamps, RunStatusTimestamp{
			Status:    otf.RunStatus(ty.Status.String),
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
