package sql

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v4"
	"github.com/leg100/otf"
)

var _ otf.RunStore = (*RunDB)(nil)

type RunDB struct {
	*pgx.Conn
}

func NewRunDB(db *pgx.Conn) *RunDB {
	return &RunDB{
		Conn: db,
	}
}

// Create persists a Run to the DB.
func (db RunDB) Create(run *otf.Run) (*otf.Run, error) {
	ctx := context.Background()

	tx, err := db.Conn.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	q := NewQuerier(tx)

	result, err := q.InsertRun(ctx, InsertRunParams{
		ID:                     run.ID,
		PlanID:                 run.Plan.ID,
		ApplyID:                run.Apply.ID,
		IsDestroy:              run.IsDestroy,
		Refresh:                run.Refresh,
		RefreshOnly:            run.RefreshOnly,
		Status:                 run.Status,
		PlanStatus:             run.Plan.Status,
		ApplyStatus:            run.Apply.Status,
		ReplaceAddrs:           run.ReplaceAddrs,
		TargetAddrs:            run.TargetAddrs,
		ConfigurationVersionID: run.ConfigurationVersion.ID,
		WorkspaceID:            run.Workspace.ID,
	})
	if err != nil {
		return nil, err
	}
	run.CreatedAt = result.CreatedAt
	run.UpdatedAt = result.UpdatedAt

	if err := insertRunStatusTimestamp(ctx, q, run); err != nil {
		return nil, err
	}

	if err := insertPlanStatusTimestamp(ctx, q, run); err != nil {
		return nil, err
	}

	if err := insertApplyStatusTimestamp(ctx, q, run); err != nil {
		return nil, err
	}

	return run, tx.Commit(ctx)
}

func (db RunDB) UpdateStatus(opts otf.RunGetOptions, fn func(*otf.Run) error) (*otf.Run, error) {
	ctx := context.Background()

	tx, err := db.Conn.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	q := NewQuerier(tx)

	// select ...for update
	var result interface{}
	switch {
	case opts.ID != nil:
		result, err = q.FindRunByIDForUpdate(ctx, *opts.ID)
	case opts.PlanID != nil:
		result, err = q.FindRunByPlanIDForUpdate(ctx, *opts.PlanID)
	case opts.ApplyID != nil:
		result, err = q.FindRunByApplyIDForUpdate(ctx, *opts.ApplyID)
	default:
		return nil, fmt.Errorf("invalid run get spec")
	}
	if err != nil {
		return nil, err
	}
	run, err := otf.UnmarshalRunFromDB(result)
	if err != nil {
		return nil, err
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
		run.UpdatedAt, err = q.UpdateRunStatus(ctx, run.Status, run.ID)
		if err != nil {
			return nil, err
		}

		if err := insertRunStatusTimestamp(ctx, q, run); err != nil {
			return nil, err
		}
	}

	if run.Plan.Status != planStatus {
		var err error
		run.UpdatedAt, err = q.UpdatePlanStatus(ctx, run.Plan.Status, run.Plan.ID)
		if err != nil {
			return nil, err
		}

		if err := insertPlanStatusTimestamp(ctx, q, run); err != nil {
			return nil, err
		}
	}

	if run.Apply.Status != applyStatus {
		var err error
		run.UpdatedAt, err = q.UpdateApplyStatus(ctx, run.Apply.Status, run.Apply.ID)
		if err != nil {
			return nil, err
		}

		if err := insertApplyStatusTimestamp(ctx, q, run); err != nil {
			return nil, err
		}
	}

	return run, tx.Commit(ctx)
}

func (db RunDB) CreatePlanReport(planID string, summary otf.ResourceReport) error {
	q := NewQuerier(db.Conn)
	ctx := context.Background()

	_, err := q.UpdateRunPlannedChangesByPlanID(ctx, UpdateRunPlannedChangesByPlanIDParams{
		ID:           planID,
		Additions:    int32(summary.Additions),
		Changes:      int32(summary.Changes),
		Destructions: int32(summary.Destructions),
	})
	return err
}

func (db RunDB) CreateApplyReport(applyID string, summary otf.ResourceReport) error {
	q := NewQuerier(db.Conn)
	ctx := context.Background()

	_, err := q.UpdateRunAppliedChangesByApplyID(ctx, UpdateRunAppliedChangesByApplyIDParams{
		ID:           applyID,
		Additions:    int32(summary.Additions),
		Changes:      int32(summary.Changes),
		Destructions: int32(summary.Destructions),
	})
	return err
}

func (db RunDB) List(opts otf.RunListOptions) (*otf.RunList, error) {
	q := NewQuerier(db.Conn)
	ctx := context.Background()

	var rows interface{}
	var err error
	if opts.WorkspaceID != nil {
		rows, err = q.FindRunsByWorkspaceID(ctx, FindRunsByWorkspaceIDParams{
			WorkspaceID: *opts.WorkspaceID,
			Limit:       opts.GetLimit(),
			Offset:      opts.GetOffset(),
		})
	} else if opts.OrganizationName != nil && opts.WorkspaceName != nil {
		rows, err = q.FindRunsByWorkspaceName(ctx, FindRunsByWorkspaceNameParams{
			OrganizationName: *opts.OrganizationName,
			WorkspaceName:    *opts.WorkspaceName,
			Limit:            opts.GetLimit(),
			Offset:           opts.GetOffset(),
		})
	} else if len(opts.Statuses) > 0 {
		rows, err = q.FindRunsByStatuses(ctx, FindRunsByStatusesParams{
			Statuses: convertStatusSliceToStringSlice(opts.Statuses),
			Limit:    opts.GetLimit(),
			Offset:   opts.GetOffset(),
		})
	} else {
		return nil, fmt.Errorf("no list filter specified")
	}
	if err != nil {
		return nil, err
	}
	runs, count, err := otf.UnmarshalRunListFromDB(rows)
	if err != nil {
		return nil, err
	}

	return &otf.RunList{
		Items:      runs,
		Pagination: otf.NewPagination(opts.ListOptions, count),
	}, nil
}

// Get retrieves a Run domain obj
func (db RunDB) Get(opts otf.RunGetOptions) (*otf.Run, error) {
	q := NewQuerier(db.Conn)
	ctx := context.Background()

	if opts.ID != nil {
		result, err := q.FindRunByID(ctx, *opts.ID)
		if err != nil {
			return nil, err
		}
		return otf.UnmarshalRunFromDB(result)
	} else if opts.PlanID != nil {
		result, err := q.FindRunByPlanID(ctx, *opts.PlanID)
		if err != nil {
			return nil, err
		}
		return otf.UnmarshalRunFromDB(result)
	} else if opts.ApplyID != nil {
		result, err := q.FindRunByApplyID(ctx, *opts.ApplyID)
		if err != nil {
			return nil, err
		}
		return otf.UnmarshalRunFromDB(result)
	} else {
		return nil, fmt.Errorf("no ID specified")
	}
}

// SetPlanFile writes a plan file to the db
func (db RunDB) SetPlanFile(id string, file []byte, format otf.PlanFormat) error {
	q := NewQuerier(db.Conn)
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
	q := NewQuerier(db.Conn)
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
	q := NewQuerier(db.Conn)
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

func insertRunStatusTimestamp(ctx context.Context, q *DBQuerier, run *otf.Run) error {
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

func insertPlanStatusTimestamp(ctx context.Context, q *DBQuerier, run *otf.Run) error {
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

func insertApplyStatusTimestamp(ctx context.Context, q *DBQuerier, run *otf.Run) error {
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
