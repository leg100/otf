package sql

import (
	"context"
	"fmt"
	"strconv"

	"github.com/jackc/pgx/v4"
	"github.com/leg100/otf"
	"github.com/leg100/otf/sql/pggen"
)

// CreateRun persists a Run to the DB.
func (db *DB) CreateRun(ctx context.Context, run *otf.Run) error {
	return db.tx(ctx, func(tx *DB) error {
		_, err := tx.InsertRun(ctx, pggen.InsertRunParams{
			ID:                     String(run.ID()),
			CreatedAt:              Timestamptz(run.CreatedAt()),
			IsDestroy:              run.IsDestroy(),
			Refresh:                run.Refresh(),
			RefreshOnly:            run.RefreshOnly(),
			Status:                 String(string(run.Status())),
			ReplaceAddrs:           run.ReplaceAddrs(),
			TargetAddrs:            run.TargetAddrs(),
			AutoApply:              run.AutoApply(),
			ConfigurationVersionID: String(run.ConfigurationVersionID()),
			WorkspaceID:            String(run.WorkspaceID()),
		})
		if err != nil {
			return fmt.Errorf("inserting run: %w", err)
		}
		_, err = tx.InsertPlan(ctx, String(run.ID()), String(string(run.Plan().Status())))
		if err != nil {
			return fmt.Errorf("inserting plan: %w", err)
		}
		_, err = tx.InsertApply(ctx, String(run.ID()), String(string(run.Apply().Status())))
		if err != nil {
			return fmt.Errorf("inserting apply: %w", err)
		}
		if err := tx.insertRunStatusTimestamp(ctx, run); err != nil {
			return fmt.Errorf("inserting run status timestamp: %w", err)
		}
		if err := tx.insertPhaseStatusTimestamp(ctx, run.Plan()); err != nil {
			return fmt.Errorf("inserting plan status timestamp: %w", err)
		}
		if err := tx.insertPhaseStatusTimestamp(ctx, run.Apply()); err != nil {
			return fmt.Errorf("inserting apply status timestamp: %w", err)
		}
		return nil
	})
}

// UpdateStatus updates the run status as well as its plan and/or apply.
func (db *DB) UpdateStatus(ctx context.Context, runID string, fn func(*otf.Run) error) (*otf.Run, error) {
	var run *otf.Run
	err := db.tx(ctx, func(tx *DB) error {
		// select ...for update
		result, err := tx.FindRunByIDForUpdate(ctx, String(runID))
		if err != nil {
			return Error(err)
		}
		run, err = otf.UnmarshalRunResult(otf.RunResult(result))
		if err != nil {
			return err
		}

		// Make copies of run attributes before update
		runStatus := run.Status()
		planStatus := run.Plan().Status()
		applyStatus := run.Apply().Status()
		forceCancelAvailableAt := run.ForceCancelAvailableAt()

		if err := fn(run); err != nil {
			return err
		}

		if run.Status() != runStatus {
			_, err := tx.UpdateRunStatus(ctx, String(string(run.Status())), String(run.ID()))
			if err != nil {
				return err
			}

			if err := tx.insertRunStatusTimestamp(ctx, run); err != nil {
				return err
			}
		}

		if run.Plan().Status() != planStatus {
			_, err := tx.UpdatePlanStatusByID(ctx, String(string(run.Plan().Status())), String(run.ID()))
			if err != nil {
				return err
			}

			if err := tx.insertPhaseStatusTimestamp(ctx, run.Plan()); err != nil {
				return err
			}
		}

		if run.Apply().Status() != applyStatus {
			_, err := tx.UpdateApplyStatusByID(ctx, String(string(run.Apply().Status())), String(run.ID()))
			if err != nil {
				return err
			}

			if err := tx.insertPhaseStatusTimestamp(ctx, run.Apply()); err != nil {
				return err
			}
		}

		if run.ForceCancelAvailableAt() != forceCancelAvailableAt && run.ForceCancelAvailableAt() != nil {
			_, err := tx.UpdateRunForceCancelAvailableAt(ctx, Timestamptz(*run.ForceCancelAvailableAt()), String(run.ID()))
			if err != nil {
				return err
			}
		}

		return nil
	})
	return run, err
}

func (db *DB) CreatePlanReport(ctx context.Context, runID string, report otf.ResourceReport) error {
	_, err := db.UpdatePlannedChangesByID(ctx, pggen.UpdatePlannedChangesByIDParams{
		RunID:        String(runID),
		Additions:    report.Additions,
		Changes:      report.Changes,
		Destructions: report.Destructions,
	})
	if err != nil {
		return Error(err)
	}
	return err
}

func (db *DB) CreateApplyReport(ctx context.Context, runID string, report otf.ResourceReport) error {
	_, err := db.UpdateAppliedChangesByID(ctx, pggen.UpdateAppliedChangesByIDParams{
		RunID:        String(runID),
		Additions:    report.Additions,
		Changes:      report.Changes,
		Destructions: report.Destructions,
	})
	if err != nil {
		return Error(err)
	}
	return err
}

func (db *DB) ListRuns(ctx context.Context, opts otf.RunListOptions) (*otf.RunList, error) {
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

	var items []*otf.Run
	for _, r := range rows {
		run, err := otf.UnmarshalRunResult(otf.RunResult(r))
		if err != nil {
			return nil, err
		}
		items = append(items, run)
	}

	return &otf.RunList{
		Items:      items,
		Pagination: otf.NewPagination(opts.ListOptions, *count),
	}, nil
}

// GetRun retrieves a run using the get options
func (db *DB) GetRun(ctx context.Context, runID string) (*otf.Run, error) {
	result, err := db.FindRunByID(ctx, String(runID))
	if err != nil {
		return nil, Error(err)
	}
	return otf.UnmarshalRunResult(otf.RunResult(result))
}

// SetPlanFile writes a plan file to the db
func (db *DB) SetPlanFile(ctx context.Context, runID string, file []byte, format otf.PlanFormat) error {
	switch format {
	case otf.PlanFormatBinary:
		_, err := db.UpdatePlanBinByID(ctx, file, String(runID))
		return err
	case otf.PlanFormatJSON:
		_, err := db.UpdatePlanJSONByID(ctx, file, String(runID))
		return err
	default:
		return fmt.Errorf("unknown plan format: %s", string(format))
	}
}

// GetPlanFile retrieves a plan file for the run
func (db *DB) GetPlanFile(ctx context.Context, runID string, format otf.PlanFormat) ([]byte, error) {
	switch format {
	case otf.PlanFormatBinary:
		return db.GetPlanBinByID(ctx, String(runID))
	case otf.PlanFormatJSON:
		return db.GetPlanJSONByID(ctx, String(runID))
	default:
		return nil, fmt.Errorf("unknown plan format: %s", string(format))
	}
}

// GetLockFile retrieves the lock file for the run
func (db *DB) GetLockFile(ctx context.Context, runID string) ([]byte, error) {
	return db.Querier.GetLockFileByID(ctx, String(runID))
}

// SetLockFile sets the lock file for the run
func (db *DB) SetLockFile(ctx context.Context, runID string, lockFile []byte) error {
	_, err := db.PutLockFile(ctx, lockFile, String(runID))
	return err
}

// DeleteRun deletes a run from the DB
func (db *DB) DeleteRun(ctx context.Context, id string) error {
	_, err := db.DeleteRunByID(ctx, String(id))
	return err
}

func (db *DB) insertRunStatusTimestamp(ctx context.Context, run *otf.Run) error {
	ts, err := run.StatusTimestamp(run.Status())
	if err != nil {
		return err
	}
	_, err = db.InsertRunStatusTimestamp(ctx, pggen.InsertRunStatusTimestampParams{
		ID:        String(run.ID()),
		Status:    String(string(run.Status())),
		Timestamp: Timestamptz(ts),
	})
	return err
}

func (db *DB) insertPhaseStatusTimestamp(ctx context.Context, phase otf.Phase) error {
	ts, err := phase.StatusTimestamp(phase.Status())
	if err != nil {
		return err
	}
	_, err = db.InsertPhaseStatusTimestamp(ctx, pggen.InsertPhaseStatusTimestampParams{
		RunID:     String(phase.ID()),
		Phase:     String(string(phase.Phase())),
		Status:    String(string(phase.Status())),
		Timestamp: Timestamptz(ts),
	})
	return err
}

func convertStatusSliceToStringSlice(statuses []otf.RunStatus) (s []string) {
	for _, status := range statuses {
		s = append(s, string(status))
	}
	return
}
