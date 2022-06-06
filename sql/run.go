package sql

import (
	"context"
	"fmt"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/leg100/otf"
	"github.com/leg100/otf/sql/pggen"
)

var _ otf.RunStore = (*RunDB)(nil)

type RunDB struct {
	conn
	pggen.Querier
}

func newRunDB(conn conn) *RunDB {
	return &RunDB{
		conn:    conn,
		Querier: pggen.NewQuerier(conn),
	}
}

func (db RunDB) Tx(ctx context.Context, callback func(otf.RunStore) error) error {
	tx, err := db.Begin(ctx)
	if err != nil {
		return err
	}
	if err := callback(newRunDB(tx)); err != nil {
		return tx.Rollback(ctx)
	}
	return tx.Commit(ctx)
}

// Create persists a Run to the DB. Should be wrapped in a transaction.
func (db RunDB) Create(ctx context.Context, run *otf.Run) error {
	_, err := db.InsertRun(ctx, pggen.InsertRunParams{
		ID:                     pgtype.Text{String: run.ID(), Status: pgtype.Present},
		CreatedAt:              run.CreatedAt(),
		PlanID:                 pgtype.Text{String: run.Plan.ID(), Status: pgtype.Present},
		ApplyID:                pgtype.Text{String: run.Apply.ID(), Status: pgtype.Present},
		IsDestroy:              run.IsDestroy(),
		Refresh:                run.Refresh(),
		RefreshOnly:            run.RefreshOnly(),
		Status:                 pgtype.Text{String: string(run.Status()), Status: pgtype.Present},
		PlanStatus:             pgtype.Text{String: string(run.Plan.Status()), Status: pgtype.Present},
		ApplyStatus:            pgtype.Text{String: string(run.Apply.Status()), Status: pgtype.Present},
		ReplaceAddrs:           run.ReplaceAddrs(),
		TargetAddrs:            run.TargetAddrs(),
		PlannedAdditions:       0,
		PlannedChanges:         0,
		PlannedDestructions:    0,
		AppliedAdditions:       0,
		AppliedChanges:         0,
		AppliedDestructions:    0,
		ConfigurationVersionID: pgtype.Text{String: run.ConfigurationVersion.ID(), Status: pgtype.Present},
		WorkspaceID:            pgtype.Text{String: run.Workspace.ID(), Status: pgtype.Present},
	})
	if err != nil {
		return err
	}
	if err := db.insertRunStatusTimestamp(ctx, run); err != nil {
		return fmt.Errorf("inserting run status timestamp: %w", err)
	}
	if err := db.insertPlanStatusTimestamp(ctx, run); err != nil {
		return fmt.Errorf("inserting plan status timestamp: %w", err)
	}
	if err := db.insertApplyStatusTimestamp(ctx, run); err != nil {
		return fmt.Errorf("inserting apply status timestamp: %w", err)
	}
	return nil
}

// UpdateStatus updates the run status as well as its plan and/or apply. Wrap in
// a tx.
func (db RunDB) UpdateStatus(ctx context.Context, opts otf.RunGetOptions, fn func(*otf.Run) error) (*otf.Run, error) {
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

		if err := db.insertPlanStatusTimestamp(ctx, run); err != nil {
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

		if err := db.insertApplyStatusTimestamp(ctx, run); err != nil {
			return nil, err
		}
	}
	return run, nil
}

func (db RunDB) CreatePlanReport(ctx context.Context, planID string, report otf.ResourceReport) error {
	q := pggen.NewQuerier(db)
	_, err := q.UpdateRunPlannedChangesByPlanID(ctx, pggen.UpdateRunPlannedChangesByPlanIDParams{
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

func (db RunDB) CreateApplyReport(ctx context.Context, applyID string, report otf.ResourceReport) error {
	q := pggen.NewQuerier(db)
	_, err := q.UpdateRunAppliedChangesByApplyID(ctx, pggen.UpdateRunAppliedChangesByApplyIDParams{
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

func (db RunDB) List(ctx context.Context, opts otf.RunListOptions) (*otf.RunList, error) {
	q := pggen.NewQuerier(db)
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
	q.FindRunsBatch(batch, pggen.FindRunsParams{
		OrganizationNames:           []string{organizationName},
		WorkspaceNames:              []string{workspaceName},
		WorkspaceIds:                []string{workspaceID},
		Statuses:                    statuses,
		Limit:                       opts.GetLimit(),
		Offset:                      opts.GetOffset(),
		IncludeConfigurationVersion: includeConfigurationVersion(opts.Include),
		IncludeWorkspace:            includeWorkspace(opts.Include),
	})
	q.CountRunsBatch(batch, pggen.CountRunsParams{
		OrganizationNames: []string{organizationName},
		WorkspaceNames:    []string{workspaceName},
		WorkspaceIds:      []string{workspaceID},
		Statuses:          statuses,
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

// Get retrieves a run using the get options
func (db RunDB) Get(ctx context.Context, opts otf.RunGetOptions) (*otf.Run, error) {
	q := pggen.NewQuerier(db)
	// Get run ID first
	runID, err := db.getRunID(ctx, opts)
	if err != nil {
		return nil, databaseError(err)
	}
	// ...now get full run
	result, err := q.FindRunByID(ctx, pggen.FindRunByIDParams{
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
func (db RunDB) SetPlanFile(ctx context.Context, planID string, file []byte, format otf.PlanFormat) error {
	q := pggen.NewQuerier(db)
	switch format {
	case otf.PlanFormatBinary:
		_, err := q.UpdateRunPlanBinByPlanID(ctx,
			file,
			pgtype.Text{String: planID, Status: pgtype.Present},
		)
		return err
	case otf.PlanFormatJSON:
		_, err := q.UpdateRunPlanJSONByPlanID(ctx,
			file,
			pgtype.Text{String: planID, Status: pgtype.Present},
		)
		return err
	default:
		return fmt.Errorf("unknown plan format: %s", string(format))
	}
}

// GetPlanFile retrieves a plan file for the run
func (db RunDB) GetPlanFile(ctx context.Context, id string, format otf.PlanFormat) ([]byte, error) {
	q := pggen.NewQuerier(db)
	switch format {
	case otf.PlanFormatBinary:
		return q.GetPlanBinByRunID(ctx, pgtype.Text{String: id, Status: pgtype.Present})
	case otf.PlanFormatJSON:
		return q.GetPlanJSONByRunID(ctx, pgtype.Text{String: id, Status: pgtype.Present})
	default:
		return nil, fmt.Errorf("unknown plan format: %s", string(format))
	}
}

// Delete deletes a run from the DB
func (db RunDB) Delete(ctx context.Context, id string) error {
	q := pggen.NewQuerier(db)
	result, err := q.DeleteRunByID(ctx, pgtype.Text{String: id, Status: pgtype.Present})
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return otf.ErrResourceNotFound
	}

	return nil
}

func (db RunDB) getRunID(ctx context.Context, opts otf.RunGetOptions) (pgtype.Text, error) {
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

func (db RunDB) insertRunStatusTimestamp(ctx context.Context, run *otf.Run) error {
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

func (db RunDB) insertPlanStatusTimestamp(ctx context.Context, run *otf.Run) error {
	ts, err := run.PlanStatusTimestamp(run.Plan.Status())
	if err != nil {
		return err
	}
	_, err = db.InsertPlanStatusTimestamp(ctx, pggen.InsertPlanStatusTimestampParams{
		ID:        pgtype.Text{String: run.ID(), Status: pgtype.Present},
		Status:    pgtype.Text{String: string(run.Plan.Status()), Status: pgtype.Present},
		Timestamp: ts,
	})
	return err
}

func (db RunDB) insertApplyStatusTimestamp(ctx context.Context, run *otf.Run) error {
	ts, err := run.ApplyStatusTimestamp(run.Apply.Status())
	if err != nil {
		return err
	}
	_, err = db.InsertApplyStatusTimestamp(ctx, pggen.InsertApplyStatusTimestampParams{
		ID:        pgtype.Text{String: run.ID(), Status: pgtype.Present},
		Status:    pgtype.Text{String: string(run.Apply.Status()), Status: pgtype.Present},
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
