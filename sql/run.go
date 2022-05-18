package sql

import (
	"context"
	"fmt"

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

	result, err := q.InsertRun(ctx, pggen.InsertRunParams{
		ID:                     run.ID,
		PlanID:                 run.Plan.ID,
		ApplyID:                run.Apply.ID,
		IsDestroy:              run.IsDestroy,
		Refresh:                run.Refresh,
		RefreshOnly:            run.RefreshOnly,
		Status:                 string(run.Status),
		PlanStatus:             string(run.Plan.Status),
		ApplyStatus:            string(run.Apply.Status),
		ReplaceAddrs:           run.ReplaceAddrs,
		TargetAddrs:            run.TargetAddrs,
		ConfigurationVersionID: run.ConfigurationVersion.ID,
		WorkspaceID:            run.Workspace.ID,
	})
	if err != nil {
		return err
	}
	run.CreatedAt = result.CreatedAt
	run.UpdatedAt = result.UpdatedAt

	if err := insertRunStatusTimestamp(ctx, q, run); err != nil {
		return err
	}

	if err := insertPlanStatusTimestamp(ctx, q, run); err != nil {
		return err
	}

	if err := insertApplyStatusTimestamp(ctx, q, run); err != nil {
		return err
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

	// select ...for update
	var run *otf.Run
	switch {
	case opts.ID != nil:
		result, err := q.FindRunByIDForUpdate(ctx, *opts.ID)
		if err != nil {
			return nil, err
		}
		run, err = otf.UnmarshalRunDBResult(otf.RunDBResult(result))
		if err != nil {
			return nil, err
		}
	case opts.PlanID != nil:
		result, err := q.FindRunByPlanIDForUpdate(ctx, *opts.PlanID)
		if err != nil {
			return nil, err
		}
		run, err = otf.UnmarshalRunDBResult(otf.RunDBResult(result))
		if err != nil {
			return nil, err
		}
	case opts.ApplyID != nil:
		result, err := q.FindRunByApplyIDForUpdate(ctx, *opts.ApplyID)
		if err != nil {
			return nil, err
		}
		run, err = otf.UnmarshalRunDBResult(otf.RunDBResult(result))
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("invalid run get spec")
	}

	// Make copies of statuses before update
	runStatus := run.Status
	planStatus := run.Plan.Status
	applyStatus := run.Apply.Status

	if err := fn(run); err != nil {
		return nil, err
	}

	if run.Status != runStatus {
		var err error
		run.UpdatedAt, err = q.UpdateRunStatus(ctx, string(run.Status), run.ID)
		if err != nil {
			return nil, err
		}

		if err := insertRunStatusTimestamp(ctx, q, run); err != nil {
			return nil, err
		}
	}

	if run.Plan.Status != planStatus {
		var err error
		run.UpdatedAt, err = q.UpdatePlanStatus(ctx, string(run.Plan.Status), run.Plan.ID)
		if err != nil {
			return nil, err
		}

		if err := insertPlanStatusTimestamp(ctx, q, run); err != nil {
			return nil, err
		}
	}

	if run.Apply.Status != applyStatus {
		var err error
		run.UpdatedAt, err = q.UpdateApplyStatus(ctx, string(run.Apply.Status), run.Apply.ID)
		if err != nil {
			return nil, err
		}

		if err := insertApplyStatusTimestamp(ctx, q, run); err != nil {
			return nil, err
		}
	}

	return run, tx.Commit(ctx)
}

func (db RunDB) CreatePlanReport(runID string, report otf.ResourceReport) error {
	q := pggen.NewQuerier(db.Pool)
	ctx := context.Background()

	result, err := q.UpdateRunPlannedChangesByRunID(ctx, pggen.UpdateRunPlannedChangesByRunIDParams{
		ID:           runID,
		Additions:    report.Additions,
		Changes:      report.Changes,
		Destructions: report.Destructions,
	})
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return otf.ErrResourceNotFound
	}

	return err
}

func (db RunDB) CreateApplyReport(applyID string, summary otf.ResourceReport) error {
	q := pggen.NewQuerier(db.Pool)
	ctx := context.Background()

	result, err := q.UpdateRunAppliedChangesByApplyID(ctx, pggen.UpdateRunAppliedChangesByApplyIDParams{
		ID:           applyID,
		Additions:    summary.Additions,
		Changes:      summary.Changes,
		Destructions: summary.Destructions,
	})
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return otf.ErrResourceNotFound
	}

	return err
}

func (db RunDB) List(opts otf.RunListOptions) (*otf.RunList, error) {
	q := pggen.NewQuerier(db.Pool)
	batch := &pgx.Batch{}
	ctx := context.Background()

	var workspaceID string
	if opts.OrganizationName != nil && opts.WorkspaceName != nil {
		wid, err := q.FindWorkspaceIDByName(ctx, *opts.WorkspaceName, *opts.OrganizationName)
		if err != nil {
			return nil, err
		}
		workspaceID = wid
	} else if opts.WorkspaceID != nil {
		workspaceID = *opts.WorkspaceID
	} else {
		// Match any workspace ID
		workspaceID = "%"
	}

	var statuses []string
	if len(opts.Statuses) > 0 {
		statuses = convertStatusSliceToStringSlice(opts.Statuses)
	} else {
		// Match any status
		statuses = []string{"%"}
	}

	q.FindRunsBatch(batch, pggen.FindRunsParams{
		WorkspaceIds: []string{workspaceID},
		Statuses:     statuses,
		Limit:        opts.GetLimit(),
		Offset:       opts.GetOffset(),
	})
	q.CountRunsBatch(batch, []string{workspaceID}, statuses)
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

// Get retrieves a Run domain obj
func (db RunDB) Get(opts otf.RunGetOptions) (*otf.Run, error) {
	q := pggen.NewQuerier(db.Pool)
	ctx := context.Background()

	if opts.ID != nil {
		result, err := q.FindRunByID(ctx, *opts.ID)
		if err != nil {
			return nil, err
		}
		return otf.UnmarshalRunDBResult(otf.RunDBResult(result))
	} else if opts.PlanID != nil {
		result, err := q.FindRunByPlanID(ctx, *opts.PlanID)
		if err != nil {
			return nil, err
		}
		return otf.UnmarshalRunDBResult(otf.RunDBResult(result))
	} else if opts.ApplyID != nil {
		result, err := q.FindRunByApplyID(ctx, *opts.ApplyID)
		if err != nil {
			return nil, err
		}
		return otf.UnmarshalRunDBResult(otf.RunDBResult(result))
	} else {
		return nil, fmt.Errorf("no ID specified")
	}
}

// SetPlanFile writes a plan file to the db
func (db RunDB) SetPlanFile(id string, file []byte, format otf.PlanFormat) error {
	q := pggen.NewQuerier(db.Pool)
	ctx := context.Background()

	switch format {
	case otf.PlanFormatBinary:
		_, err := q.PutPlanBinByRunID(ctx, file, id)
		return err
	case otf.PlanFormatJSON:
		_, err := q.PutPlanJSONByRunID(ctx, file, id)
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
		return q.GetPlanBinByRunID(ctx, id)
	case otf.PlanFormatJSON:
		return q.GetPlanJSONByRunID(ctx, id)
	default:
		return nil, fmt.Errorf("unknown plan format: %s", string(format))
	}
}

// Delete deletes a run from the DB
func (db RunDB) Delete(id string) error {
	q := pggen.NewQuerier(db.Pool)
	ctx := context.Background()

	result, err := q.DeleteRunByID(ctx, id)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return otf.ErrResourceNotFound
	}

	return nil
}

func insertRunStatusTimestamp(ctx context.Context, q *pggen.DBQuerier, run *otf.Run) error {
	ts, err := q.InsertRunStatusTimestamp(ctx, run.ID, string(run.Status))
	if err != nil {
		return err
	}
	run.StatusTimestamps = append(run.StatusTimestamps, otf.RunStatusTimestamp{
		Status:    otf.RunStatus(ts.Status),
		Timestamp: ts.Timestamp,
	})

	return nil
}

func insertPlanStatusTimestamp(ctx context.Context, q *pggen.DBQuerier, run *otf.Run) error {
	ts, err := q.InsertPlanStatusTimestamp(ctx, run.ID, string(run.Plan.Status))
	if err != nil {
		return err
	}
	run.Plan.StatusTimestamps = append(run.Plan.StatusTimestamps, otf.PlanStatusTimestamp{
		Status:    otf.PlanStatus(ts.Status),
		Timestamp: ts.Timestamp,
	})

	return nil
}

func insertApplyStatusTimestamp(ctx context.Context, q *pggen.DBQuerier, run *otf.Run) error {
	ts, err := q.InsertApplyStatusTimestamp(ctx, run.ID, string(run.Apply.Status))
	if err != nil {
		return err
	}
	run.Apply.StatusTimestamps = append(run.Apply.StatusTimestamps, otf.ApplyStatusTimestamp{
		Status:    otf.ApplyStatus(ts.Status),
		Timestamp: ts.Timestamp,
	})

	return nil
}

func convertStatusSliceToStringSlice(statuses []otf.RunStatus) (s []string) {
	for _, status := range statuses {
		s = append(s, string(status))
	}
	return
}
