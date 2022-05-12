package sql

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/leg100/otf"
)

var _ otf.RunStore = (*RunDB)(nil)

type RunDB struct {
	*pgx.Conn
}

type Timestamps interface {
	GetCreatedAt() time.Time
	GetUpdatedAt() time.Time
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

	_, err = q.InsertRun(ctx, InsertRunParams{
		ID:                     &run.ID,
		IsDestroy:              &run.IsDestroy,
		Refresh:                &run.Refresh,
		RefreshOnly:            &run.RefreshOnly,
		Status:                 otf.String(string(run.Status)),
		ReplaceAddrs:           run.ReplaceAddrs,
		TargetAddrs:            run.TargetAddrs,
		ConfigurationVersionID: &run.ConfigurationVersion.ID,
		WorkspaceID:            &run.Workspace.ID,
	})
	if err != nil {
		return nil, err
	}

	// Insert timestamp for current run status
	_, err = q.InsertRunStatusTimestamp(ctx, &run.ID, otf.String(string(run.Status)))
	if err != nil {
		return nil, err
	}

	// Insert plan
	_, err = q.InsertPlan(ctx, InsertPlanParams{
		ID: &run.Plan.ID,
	})
	if err != nil {
		return nil, err
	}

	// Insert timestamp for current plan status
	_, err = q.InsertPlanStatusTimestamp(ctx, &run.Plan.ID, otf.String(string(run.Plan.Status)))
	if err != nil {
		return nil, err
	}

	// Insert apply
	_, err = q.InsertApply(ctx, InsertApplyParams{
		ID: &run.Apply.ID,
	})
	if err != nil {
		return nil, err
	}

	// Insert timestamp for current apply status
	_, err = q.InsertApplyStatusTimestamp(ctx, &run.Apply.ID, otf.String(string(run.Apply.Status)))
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	// Return newly created run to caller
	return getRun(ctx, q, otf.RunGetOptions{ID: &run.ID})
}

func (db RunDB) UpdateStatus(id string, fn func(*otf.Run) error) (*otf.Run, error) {
	ctx := context.Background()

	tx, err := db.Conn.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	q := NewQuerier(tx)

	// select ...for update
	result, err := q.FindRunByIDForUpdate(ctx, &id)
	if err != nil {
		return nil, err
	}
	run := convertRun(result)

	// Make copies of statuses before update
	runStatus := run.Status
	planStatus := run.Plan.Status
	applyStatus := run.Apply.Status

	if err := fn(run); err != nil {
		return nil, err
	}

	if run.Status != runStatus {
		result, err := q.UpdateRunStatus(ctx, otf.String(string(run.Status)), &run.ID)
		if err != nil {
			return nil, err
		}
		addResultToRun(run, result)

		ts, err := q.InsertRunStatusTimestamp(ctx, &run.ID, otf.String(string(run.Status)))
		if err != nil {
			return nil, err
		}
		run.StatusTimestamps = append(run.StatusTimestamps, convertRunStatusTimestamp(ts))
	}

	if run.Plan.Status != planStatus {
		result, err := q.UpdatePlanStatus(ctx, otf.String(string(run.Plan.Status)), &run.Plan.ID)
		if err != nil {
			return nil, err
		}
		addResultToPlan(run.Plan, result)

		ts, err := q.InsertPlanStatusTimestamp(ctx, &run.Plan.ID, otf.String(string(run.Plan.Status)))
		if err != nil {
			return nil, err
		}
		run.Plan.StatusTimestamps = append(run.Plan.StatusTimestamps, convertPlanStatusTimestamp(ts))
	}

	if run.Apply.Status != applyStatus {
		result, err := q.UpdateApplyStatus(ctx, otf.String(string(run.Apply.Status)), &run.Apply.ID)
		if err != nil {
			return nil, err
		}
		addResultToApply(run.Apply, result)

		ts, err := q.InsertApplyStatusTimestamp(ctx, &run.Apply.ID, otf.String(string(run.Apply.Status)))
		if err != nil {
			return nil, err
		}
		run.Apply.StatusTimestamps = append(run.Apply.StatusTimestamps, convertApplyStatusTimestamp(ts))
	}

	return run, tx.Commit(ctx)
}

func (db RunDB) UpdatePlanResources(id string, summary otf.Resources) error {
	q := NewQuerier(db.Conn)
	ctx := context.Background()

	_, err := q.UpdatePlanResources(ctx, UpdatePlanResourcesParams{
		RunID:                &id,
		ResourceAdditions:    int32(summary.ResourceAdditions),
		ResourceChanges:      int32(summary.ResourceChanges),
		ResourceDestructions: int32(summary.ResourceDestructions),
	})
	return err
}

func (db RunDB) UpdateApplyResources(id string, summary otf.Resources) error {
	q := NewQuerier(db.Conn)
	ctx := context.Background()

	_, err := q.UpdateApplyResources(ctx, UpdateApplyResourcesParams{
		RunID:                &id,
		ResourceAdditions:    int32(summary.ResourceAdditions),
		ResourceChanges:      int32(summary.ResourceChanges),
		ResourceDestructions: int32(summary.ResourceDestructions),
	})
	return err
}

func (db RunDB) List(opts otf.RunListOptions) (*otf.RunList, error) {
	q := NewQuerier(db.Conn)
	ctx := context.Background()

	var err error
	var result interface{}

	if opts.WorkspaceID != nil {
		result, err = q.FindRunsByWorkspaceID(ctx, FindRunsByWorkspaceIDParams{
			WorkspaceID: opts.WorkspaceID,
			Limit:       opts.GetLimit(),
			Offset:      opts.GetOffset(),
		})
	} else if opts.OrganizationName != nil && opts.WorkspaceName != nil {
		result, err = q.FindRunsByWorkspaceName(ctx, FindRunsByWorkspaceNameParams{
			OrganizationName: opts.OrganizationName,
			WorkspaceName:    opts.WorkspaceName,
			Limit:            opts.GetLimit(),
			Offset:           opts.GetOffset(),
		})
	} else if len(opts.Statuses) > 0 {
		result, err = q.FindRunsByStatuses(ctx, FindRunsByStatusesParams{
			Statuses: convertToStringSlice(opts.Statuses),
			Limit:    opts.GetLimit(),
			Offset:   opts.GetOffset(),
		})
	} else {
		return nil, fmt.Errorf("no list filter specified")
	}
	if err != nil {
		return nil, err
	}

	var items []*otf.Run
	for _, r := range convertToInterfaceSlice(result) {
		items = append(items, convertRun(r.(runResult)))
	}

	return &otf.RunList{
		Items:      items,
		Pagination: otf.NewPagination(opts.ListOptions, getCount(result)),
	}, nil
}

// Get retrieves a Run domain obj
func (db RunDB) Get(opts otf.RunGetOptions) (*otf.Run, error) {
	q := NewQuerier(db.Conn)
	return getRun(context.Background(), q, opts)
}

// SetPlanFile writes a plan file to the db
func (db RunDB) SetPlanFile(id string, file []byte, format otf.PlanFormat) error {
	q := NewQuerier(db.Conn)
	ctx := context.Background()

	switch format {
	case otf.PlanFormatBinary:
		_, err := q.PutPlanBinByRunID(ctx, file, &id)
		return err
	case otf.PlanFormatJSON:
		_, err := q.PutPlanJSONByRunID(ctx, file, &id)
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
		return q.GetPlanBinByRunID(ctx, &id)
	case otf.PlanFormatJSON:
		return q.GetPlanJSONByRunID(ctx, &id)
	default:
		return nil, fmt.Errorf("unknown plan format: %s", string(format))
	}
}

// Delete deletes a run from the DB
func (db RunDB) Delete(id string) error {
	q := NewQuerier(db.Conn)
	ctx := context.Background()

	result, err := q.DeleteRunByID(ctx, &id)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return otf.ErrResourceNotFound
	}

	return nil
}

func getRun(ctx context.Context, q *DBQuerier, opts otf.RunGetOptions) (*otf.Run, error) {
	if opts.ID != nil {
		result, err := q.FindRunByID(ctx, opts.ID)
		if err != nil {
			return nil, err
		}
		return convertRun(result), nil
	} else if opts.PlanID != nil {
		result, err := q.FindRunByPlanID(ctx, opts.PlanID)
		if err != nil {
			return nil, err
		}
		return convertRun(result), nil
	} else if opts.ApplyID != nil {
		result, err := q.FindRunByApplyID(ctx, opts.ApplyID)
		if err != nil {
			return nil, err
		}
		return convertRun(result), nil
	} else {
		return nil, fmt.Errorf("no ID specified")
	}
}

func convertToStringSlice(i interface{}) (s []string) {
	slice := reflect.ValueOf(i)
	for i := 0; i < slice.Len(); i++ {
		v := slice.Index(i)
		s = append(s, v.Interface().(string))
	}
	return
}
