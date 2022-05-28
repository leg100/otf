package sql

import (
	"context"
	"fmt"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/leg100/otf"
	"github.com/leg100/otf/sql/pggen"
)

var _ otf.RunStore = (*RunDB)(nil)

type RunDB struct {
	*pgxpool.Pool
}

func NewRunDB(db *pgxpool.Pool) *RunDB {
	return &RunDB{
		Pool: db,
	}
}

// Create persists a Run to the DB.
func (db RunDB) Create(run *otf.Run) error {
	ctx := context.Background()
	tx, err := db.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	q := pggen.NewQuerier(tx)
	_, err = q.InsertRun(ctx, pggen.InsertRunParams{
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
	if err := insertRunStatusTimestamp(ctx, q, run); err != nil {
		return fmt.Errorf("inserting run status timestamp: %w", err)
	}
	if err := insertPlanStatusTimestamp(ctx, q, run); err != nil {
		return fmt.Errorf("inserting plan status timestamp: %w", err)
	}
	if err := insertApplyStatusTimestamp(ctx, q, run); err != nil {
		return fmt.Errorf("inserting apply status timestamp: %w", err)
	}
	return tx.Commit(ctx)
}

func (db RunDB) UpdateStatus(opts otf.RunGetOptions, fn func(*otf.Run) error) (*otf.Run, error) {
	ctx := context.Background()

	tx, err := db.Pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	q := pggen.NewQuerier(tx)

	// Get run ID first
	runID, err := getRunID(ctx, q, opts)
	if err != nil {
		return nil, err
	}
	// select ...for update
	result, err := q.FindRunByIDForUpdate(ctx, pggen.FindRunByIDForUpdateParams{
		RunID: runID,
	})
	if err != nil {
		return nil, err
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
		_, err = q.UpdateRunStatus(ctx,
			pgtype.Text{String: string(run.Status()), Status: pgtype.Present},
			pgtype.Text{String: run.ID(), Status: pgtype.Present},
		)
		if err != nil {
			return nil, err
		}

		if err := insertRunStatusTimestamp(ctx, q, run); err != nil {
			return nil, err
		}
	}

	if run.Plan.Status() != planStatus {
		var err error
		_, err = q.UpdatePlanStatus(ctx,
			pgtype.Text{String: string(run.Plan.Status()), Status: pgtype.Present},
			pgtype.Text{String: run.Plan.ID(), Status: pgtype.Present},
		)
		if err != nil {
			return nil, err
		}

		if err := insertPlanStatusTimestamp(ctx, q, run); err != nil {
			return nil, err
		}
	}

	if run.Apply.Status() != applyStatus {
		var err error
		_, err = q.UpdateApplyStatus(ctx,
			pgtype.Text{String: string(run.Apply.Status()), Status: pgtype.Present},
			pgtype.Text{String: run.Apply.ID(), Status: pgtype.Present},
		)
		if err != nil {
			return nil, err
		}

		if err := insertApplyStatusTimestamp(ctx, q, run); err != nil {
			return nil, err
		}
	}

	return run, tx.Commit(ctx)
}

func (db RunDB) CreatePlanReport(planID string, report otf.ResourceReport) error {
	q := pggen.NewQuerier(db.Pool)
	ctx := context.Background()

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

func (db RunDB) CreateApplyReport(applyID string, report otf.ResourceReport) error {
	q := pggen.NewQuerier(db.Pool)
	ctx := context.Background()

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

func (db RunDB) List(opts otf.RunListOptions) (*otf.RunList, error) {
	q := pggen.NewQuerier(db.Pool)
	batch := &pgx.Batch{}
	ctx := context.Background()

	var workspaceID pgtype.Text
	if opts.OrganizationName != nil && opts.WorkspaceName != nil {
		wid, err := q.FindWorkspaceIDByName(ctx,
			pgtype.Text{String: *opts.WorkspaceName, Status: pgtype.Present},
			pgtype.Text{String: *opts.OrganizationName, Status: pgtype.Present},
		)
		if err != nil {
			return nil, err
		}
		workspaceID = wid
	} else if opts.WorkspaceID != nil {
		workspaceID = pgtype.Text{String: *opts.WorkspaceID, Status: pgtype.Present}
	} else {
		// Match any workspace ID
		workspaceID = pgtype.Text{String: "%", Status: pgtype.Present}
	}

	var statuses []string
	if len(opts.Statuses) > 0 {
		statuses = convertStatusSliceToStringSlice(opts.Statuses)
	} else {
		// Match any status
		statuses = []string{"%"}
	}

	q.FindRunsBatch(batch, pggen.FindRunsParams{
		WorkspaceIds:                []string{workspaceID.String},
		Statuses:                    statuses,
		Limit:                       opts.GetLimit(),
		Offset:                      opts.GetOffset(),
		IncludeConfigurationVersion: includeConfigurationVersion(opts.Include),
		IncludeWorkspace:            includeWorkspace(opts.Include),
	})
	q.CountRunsBatch(batch, []string{workspaceID.String}, statuses)
	results := db.Pool.SendBatch(ctx, batch)
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
func (db RunDB) Get(opts otf.RunGetOptions) (*otf.Run, error) {
	q := pggen.NewQuerier(db.Pool)
	ctx := context.Background()
	// Get run ID first
	runID, err := getRunID(ctx, q, opts)
	if err != nil {
		return nil, err
	}
	// ...now get full run
	result, err := q.FindRunByID(ctx, pggen.FindRunByIDParams{
		RunID:                       runID,
		IncludeConfigurationVersion: includeConfigurationVersion(opts.Include),
		IncludeWorkspace:            includeWorkspace(opts.Include),
	})
	if err != nil {
		return nil, err
	}
	return otf.UnmarshalRunDBResult(otf.RunDBResult(result))
}

// SetPlanFile writes a plan file to the db
func (db RunDB) SetPlanFile(planID string, file []byte, format otf.PlanFormat) error {
	q := pggen.NewQuerier(db.Pool)
	ctx := context.Background()

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
func (db RunDB) GetPlanFile(id string, format otf.PlanFormat) ([]byte, error) {
	q := pggen.NewQuerier(db.Pool)
	ctx := context.Background()

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
func (db RunDB) Delete(id string) error {
	q := pggen.NewQuerier(db.Pool)
	ctx := context.Background()

	result, err := q.DeleteRunByID(ctx, pgtype.Text{String: id, Status: pgtype.Present})
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return otf.ErrResourceNotFound
	}

	return nil
}

func getRunID(ctx context.Context, q *pggen.DBQuerier, opts otf.RunGetOptions) (pgtype.Text, error) {
	if opts.PlanID != nil {
		return q.FindRunIDByPlanID(ctx, pgtype.Text{String: *opts.PlanID, Status: pgtype.Present})
	} else if opts.ApplyID != nil {
		return q.FindRunIDByApplyID(ctx, pgtype.Text{String: *opts.ApplyID, Status: pgtype.Present})
	} else if opts.ID != nil {
		return pgtype.Text{String: *opts.ID, Status: pgtype.Present}, nil
	} else {
		return pgtype.Text{}, fmt.Errorf("no ID specified")
	}
}

func insertRunStatusTimestamp(ctx context.Context, q *pggen.DBQuerier, run *otf.Run) error {
	ts, err := run.StatusTimestamp(run.Status())
	if err != nil {
		return err
	}
	_, err = q.InsertRunStatusTimestamp(ctx, pggen.InsertRunStatusTimestampParams{
		ID:        pgtype.Text{String: run.ID(), Status: pgtype.Present},
		Status:    pgtype.Text{String: string(run.Status()), Status: pgtype.Present},
		Timestamp: ts,
	})
	return err
}

func insertPlanStatusTimestamp(ctx context.Context, q *pggen.DBQuerier, run *otf.Run) error {
	ts, err := run.PlanStatusTimestamp(run.Plan.Status())
	if err != nil {
		return err
	}
	_, err = q.InsertPlanStatusTimestamp(ctx, pggen.InsertPlanStatusTimestampParams{
		ID:        pgtype.Text{String: run.ID(), Status: pgtype.Present},
		Status:    pgtype.Text{String: string(run.Plan.Status()), Status: pgtype.Present},
		Timestamp: ts,
	})
	return err
}

func insertApplyStatusTimestamp(ctx context.Context, q *pggen.DBQuerier, run *otf.Run) error {
	ts, err := run.ApplyStatusTimestamp(run.Apply.Status())
	if err != nil {
		return err
	}
	_, err = q.InsertApplyStatusTimestamp(ctx, pggen.InsertApplyStatusTimestampParams{
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
