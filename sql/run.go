package sql

import (
	"context"
	"fmt"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/leg100/otf"
	"github.com/leg100/otf/sql/pggen"
)

// CreateRun persists a Run to the DB. Should be wrapped in a transaction.
func (db *DB) CreateRun(ctx context.Context, run *otf.Run) error {
	_, err := db.InsertRun(ctx, pggen.InsertRunParams{
		ID:                     pgtype.Text{String: run.ID(), Status: pgtype.Present},
		CreatedAt:              run.CreatedAt(),
		IsDestroy:              run.IsDestroy(),
		Refresh:                run.Refresh(),
		RefreshOnly:            run.RefreshOnly(),
		Status:                 pgtype.Text{String: string(run.Status()), Status: pgtype.Present},
		ReplaceAddrs:           run.ReplaceAddrs(),
		TargetAddrs:            run.TargetAddrs(),
		ConfigurationVersionID: pgtype.Text{String: run.ConfigurationVersion.ID(), Status: pgtype.Present},
		WorkspaceID:            pgtype.Text{String: run.Workspace.ID(), Status: pgtype.Present},
	})
	if err != nil {
		return err
	}
	_, err = db.InsertPlan(ctx, pggen.InsertPlanParams{
		PlanID:       pgtype.Text{String: run.Plan.ID(), Status: pgtype.Present},
		RunID:        pgtype.Text{String: run.ID(), Status: pgtype.Present},
		Status:       pgtype.Text{String: string(run.Plan.Status()), Status: pgtype.Present},
		Additions:    0,
		Changes:      0,
		Destructions: 0,
	})
	if err != nil {
		return err
	}
	_, err = db.InsertApply(ctx, pggen.InsertApplyParams{
		ApplyID:      pgtype.Text{String: run.Apply.ID(), Status: pgtype.Present},
		RunID:        pgtype.Text{String: run.ID(), Status: pgtype.Present},
		Status:       pgtype.Text{String: string(run.Apply.Status()), Status: pgtype.Present},
		Additions:    0,
		Changes:      0,
		Destructions: 0,
	})
	if err != nil {
		return err
	}
	if err := db.insertRunStatusTimestamp(ctx, run); err != nil {
		return fmt.Errorf("inserting run status timestamp: %w", err)
	}
	if err := db.insertPlanStatusTimestamp(ctx, run.Plan); err != nil {
		return fmt.Errorf("inserting plan status timestamp: %w", err)
	}
	if err := db.insertApplyStatusTimestamp(ctx, run.Apply); err != nil {
		return fmt.Errorf("inserting apply status timestamp: %w", err)
	}
	return nil
}

// UpdateStatus updates the run status as well as its plan and/or apply. Wrap in
// a tx.
func (db *DB) UpdateStatus(ctx context.Context, opts otf.RunGetOptions, fn func(*otf.Run) error) (*otf.Run, error) {
	// Get run ID first
	runID, err := db.getRunID(ctx, opts)
	if err != nil {
		return nil, databaseError(err)
	}
	// select ...for update
	result, err := db.FindRunByIDForUpdate(ctx, pggen.FindRunByIDForUpdateParams{
		RunID: runID,
	})
	if err != nil {
		return nil, databaseError(err)
	}
	run, err := otf.UnmarshalRunDBResult(otf.RunDBResult(result))
	if err != nil {
		return nil, err
	}

	// Make copies of statuses before update
	runStatus := run.Status()
	planStatus := run.Plan.Status()
	applyStatus := run.Apply.Status()

	if err := fn(run); err != nil {
		return nil, err
	}

	if run.Status() != runStatus {
		var err error
		_, err = db.UpdateRunStatus(ctx,
			pgtype.Text{String: string(run.Status()), Status: pgtype.Present},
			pgtype.Text{String: run.ID(), Status: pgtype.Present},
		)
		if err != nil {
			return nil, err
		}

		if err := db.insertRunStatusTimestamp(ctx, run); err != nil {
			return nil, err
		}
	}

	if run.Plan.Status() != planStatus {
		var err error
		_, err = db.UpdatePlanStatus(ctx,
			pgtype.Text{String: string(run.Plan.Status()), Status: pgtype.Present},
			pgtype.Text{String: run.Plan.ID(), Status: pgtype.Present},
		)
		if err != nil {
			return nil, err
		}

		if err := db.insertPlanStatusTimestamp(ctx, run.Plan); err != nil {
			return nil, err
		}
	}

	if run.Apply.Status() != applyStatus {
		var err error
		_, err = db.UpdateApplyStatus(ctx,
			pgtype.Text{String: string(run.Apply.Status()), Status: pgtype.Present},
			pgtype.Text{String: run.Apply.ID(), Status: pgtype.Present},
		)
		if err != nil {
			return nil, err
		}

		if err := db.insertApplyStatusTimestamp(ctx, run.Apply); err != nil {
			return nil, err
		}
	}
	return run, nil
}

func (db *DB) CreatePlanReport(ctx context.Context, planID string, report otf.ResourceReport) error {
	_, err := db.UpdatePlannedChangesByID(ctx, pggen.UpdatePlannedChangesByIDParams{
		PlanID:       pgtype.Text{String: planID, Status: pgtype.Present},
		Additions:    report.Additions,
		Changes:      report.Changes,
		Destructions: report.Destructions,
	})
	if err != nil {
		return databaseError(err)
	}
	return err
}

func (db *DB) CreateApplyReport(ctx context.Context, applyID string, report otf.ResourceReport) error {
	_, err := db.UpdateAppliedChangesByID(ctx, pggen.UpdateAppliedChangesByIDParams{
		ApplyID:      pgtype.Text{String: applyID, Status: pgtype.Present},
		Additions:    report.Additions,
		Changes:      report.Changes,
		Destructions: report.Destructions,
	})
	if err != nil {
		return databaseError(err)
	}
	return err
}

func (db *DB) ListRuns(ctx context.Context, opts otf.RunListOptions) (*otf.RunList, error) {
	batch := &pgx.Batch{}
	organizationName := "%"
	if opts.OrganizationName != nil {
		organizationName = *opts.OrganizationName
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
	db.FindRunsBatch(batch, pggen.FindRunsParams{
		OrganizationNames:           []string{organizationName},
		WorkspaceNames:              []string{workspaceName},
		WorkspaceIds:                []string{workspaceID},
		Statuses:                    statuses,
		Limit:                       opts.GetLimit(),
		Offset:                      opts.GetOffset(),
		IncludeConfigurationVersion: includeConfigurationVersion(opts.Include),
		IncludeWorkspace:            includeWorkspace(opts.Include),
	})
	db.CountRunsBatch(batch, pggen.CountRunsParams{
		OrganizationNames: []string{organizationName},
		WorkspaceNames:    []string{workspaceName},
		WorkspaceIds:      []string{workspaceID},
		Statuses:          statuses,
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
		run, err := otf.UnmarshalRunDBResult(otf.RunDBResult(r))
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
func (db *DB) GetRun(ctx context.Context, opts otf.RunGetOptions) (*otf.Run, error) {
	// Get run ID first
	runID, err := db.getRunID(ctx, opts)
	if err != nil {
		return nil, databaseError(err)
	}
	// ...now get full run
	result, err := db.FindRunByID(ctx, pggen.FindRunByIDParams{
		RunID:                       runID,
		IncludeConfigurationVersion: includeConfigurationVersion(opts.Include),
		IncludeWorkspace:            includeWorkspace(opts.Include),
	})
	if err != nil {
		return nil, databaseError(err)
	}
	return otf.UnmarshalRunDBResult(otf.RunDBResult(result))
}

// SetPlanFile writes a plan file to the db
func (db *DB) SetPlanFile(ctx context.Context, planID string, file []byte, format otf.PlanFormat) error {
	switch format {
	case otf.PlanFormatBinary:
		_, err := db.UpdatePlanBinByID(ctx, file, pgtype.Text{String: planID, Status: pgtype.Present})
		return err
	case otf.PlanFormatJSON:
		_, err := db.UpdatePlanJSONByID(ctx, file, pgtype.Text{String: planID, Status: pgtype.Present})
		return err
	default:
		return fmt.Errorf("unknown plan format: %s", string(format))
	}
}

// GetPlanFile retrieves a plan file for the run
func (db *DB) GetPlanFile(ctx context.Context, runID string, format otf.PlanFormat) ([]byte, error) {
	switch format {
	case otf.PlanFormatBinary:
		return db.GetPlanBinByID(ctx, pgtype.Text{String: runID, Status: pgtype.Present})
	case otf.PlanFormatJSON:
		return db.GetPlanJSONByID(ctx, pgtype.Text{String: runID, Status: pgtype.Present})
	default:
		return nil, fmt.Errorf("unknown plan format: %s", string(format))
	}
}

// DeleteRun deletes a run from the DB
func (db *DB) DeleteRun(ctx context.Context, id string) error {
	result, err := db.DeleteRunByID(ctx, pgtype.Text{String: id, Status: pgtype.Present})
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return otf.ErrResourceNotFound
	}

	return nil
}

func (db *DB) getRunID(ctx context.Context, opts otf.RunGetOptions) (pgtype.Text, error) {
	if opts.PlanID != nil {
		return db.FindRunIDByPlanID(ctx, pgtype.Text{String: *opts.PlanID, Status: pgtype.Present})
	} else if opts.ApplyID != nil {
		return db.FindRunIDByApplyID(ctx, pgtype.Text{String: *opts.ApplyID, Status: pgtype.Present})
	} else if opts.ID != nil {
		return pgtype.Text{String: *opts.ID, Status: pgtype.Present}, nil
	} else {
		return pgtype.Text{}, fmt.Errorf("no ID specified")
	}
}

func (db *DB) insertRunStatusTimestamp(ctx context.Context, run *otf.Run) error {
	ts, err := run.StatusTimestamp(run.Status())
	if err != nil {
		return err
	}
	_, err = db.InsertRunStatusTimestamp(ctx, pggen.InsertRunStatusTimestampParams{
		ID:        pgtype.Text{String: run.ID(), Status: pgtype.Present},
		Status:    pgtype.Text{String: string(run.Status()), Status: pgtype.Present},
		Timestamp: ts,
	})
	return err
}

func (db *DB) insertPlanStatusTimestamp(ctx context.Context, plan *otf.Plan) error {
	ts, err := plan.StatusTimestamp(plan.Status())
	if err != nil {
		return err
	}
	_, err = db.InsertPlanStatusTimestamp(ctx, pggen.InsertPlanStatusTimestampParams{
		ID:        pgtype.Text{String: plan.ID(), Status: pgtype.Present},
		Status:    pgtype.Text{String: string(plan.Status()), Status: pgtype.Present},
		Timestamp: ts,
	})
	return err
}

func (db *DB) insertApplyStatusTimestamp(ctx context.Context, apply *otf.Apply) error {
	ts, err := apply.StatusTimestamp(apply.Status())
	if err != nil {
		return err
	}
	_, err = db.InsertApplyStatusTimestamp(ctx, pggen.InsertApplyStatusTimestampParams{
		ID:        pgtype.Text{String: apply.ID(), Status: pgtype.Present},
		Status:    pgtype.Text{String: string(apply.Status()), Status: pgtype.Present},
		Timestamp: ts,
	})
	return err
}

func convertStatusSliceToStringSlice(statuses []otf.RunStatus) (s []string) {
	for _, status := range statuses {
		s = append(s, string(status))
	}
	return
}
