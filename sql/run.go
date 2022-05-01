package sql

import (
	"context"
	"fmt"
	"reflect"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/leg100/otf"
	"github.com/mitchellh/copystructure"
)

var _ otf.RunStore = (*RunDB)(nil)

type RunDB struct {
	*pgx.Conn
}

type Timestamps interface {
	GetCreatedAt() pgtype.Timestamptz
	GetUpdatedAt() pgtype.Timestamptz
}

type RunRow interface {
	GetRunID() *string
	GetIsDestroy() *bool
	GetWorkspaceID() *string
	GetStatus() *string
	GetRunStatusTimestamps() []RunStatusTimestamps
	GetConfigurationVersionID() *string
	GetReplaceAddrs() []string
	GetTargetAddrs() []string
	GetPlan() Plans

	Timestamps
}

type RunListRow interface {
	RunRow

	GetFullCount() *int
}

type PlanRow interface {
	GetPlanID() *string
	Timestamps
	GetStatus() *string
}

type ApplyRow interface {
	GetApplyID() *string
	Timestamps
	GetStatus() *string
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

	// Insert status timestamps
	for _, ts := range run.StatusTimestamps {
		_, err := q.InsertRunStatusTimestamp(ctx, &run.ID, otf.String(string(ts.Status)))
		if err != nil {
			return nil, err
		}
	}

	// Insert plan
	_, err = q.InsertPlan(ctx, InsertPlanParams{
		ID: &run.Plan.ID,
	})
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

	tx.Commit()

	// Get new run
	return getRun(db.Conn, otf.RunGetOptions{ID: &run.ID})
}

// Update persists an updated Run to the DB. The existing run is fetched from
// the DB, the supplied func is invoked on the run, and the updated run is
// persisted back to the DB. The returned Run includes any changes, including a
// new UpdatedAt value.
func (db RunDB) Update(opts otf.RunGetOptions, fn func(*otf.Run) error) (*otf.Run, error) {
	ctx := context.Background()

	tx, err := db.Conn.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	run, err := getRun(tx, opts)
	if err != nil {
		return nil, err
	}

	// Make copy before client updates it
	cp, err := copystructure.Copy(run)
	if err != nil {
		return nil, err
	}
	before := cp.(*otf.Run)

	// Update obj using client-supplied fn
	if err := fn(run); err != nil {
		return nil, err
	}

	q := NewQuerier(tx)

	var modified bool

	if before.Status != run.Status {
		_, err := q.UpdateRunStatus(ctx, otf.String(string(run.Status)), &run.ID)
		if err != nil {
			return nil, err
		}
		modified = true
	}

	if before.Plan.Status != run.Plan.Status {
		_, err := q.UpdatePlanStatus(ctx, otf.String(string(run.Plan.Status)), &run.Plan.ID)
		if err != nil {
			return nil, err
		}
		modified = true
	}

	if before.Apply.Status != run.Apply.Status {
		_, err := q.UpdateApplyStatus(ctx, otf.String(string(run.Apply.Status)), &run.Apply.ID)
		if err != nil {
			return nil, err
		}
		modified = true
	}

	if modified {
		if err := tx.Commit(); err != nil {
			return nil, err
		}
		return getRun(db.Conn, opts)
	}

	return run, nil
}

func (db RunDB) List(opts otf.RunListOptions) (*otf.RunList, error) {
	q := NewQuerier(db.Conn)

	ctx := context.Background()

	var result interface{}
	var err error

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

	rows := reflect.ValueOf(result)
	for i := 0; i < rows.Len(); i++ {
		v := rows.Index(i)
		rr := v.Interface().(RunRow)
		items = append(items, convertRun(rr))
	}

	var count int
	if rows.Len() > 0 {
		count = *rows.Index(0).Interface().(RunListRow).GetFullCount()
	}

	return &otf.RunList{
		Items:      items,
		Pagination: otf.NewPagination(opts.ListOptions, count),
	}, nil
}

// Get retrieves a Run domain obj
func (db RunDB) Get(opts otf.RunGetOptions) (*otf.Run, error) {
	return getRun(db.Conn, opts)
}

func (db RunDB) GetPlanFile(id string) ([]byte, error) {
	q := NewQuerier(db.Conn)
	ctx := context.Background()

	return q.GetPlanFileByRunID(ctx, &id)
}

func (db RunDB) GetPlanFile(id string) ([]byte, error) {
	q := NewQuerier(db.Conn)
	ctx := context.Background()

	return q.GetPlanFileByRunID(ctx, &id)
}

func (db RunDB) GetPlanJSON(id string) ([]byte, error) {
	q := NewQuerier(db.Conn)
	ctx := context.Background()

	return q.GetPlanJSONByRunID(ctx, &id)
}

// Delete deletes a run from the DB
func (db RunDB) Delete(id string) error {
	q := NewQuerier(db.Conn)

	ctx := context.Background()

	_, err := q.DeleteRunByID(ctx, &id)
	return err
}

func getRun(db pgx.Conn, opts otf.RunGetOptions) (*otf.Run, error) {
	q := NewQuerier(db.Conn)

	ctx := context.Background()

	var result interface{}
	var err error

	if opts.ID != nil {
		result, err = q.FindRunByID(ctx, opts.ID)
	} else if opts.PlanID != nil {
		result, err = q.FindRunByPlanID(ctx, opts.PlanID)
	} else if opts.ApplyID != nil {
		result, err = q.FindRunByApplyID(ctx, opts.ApplyID)
	} else {
		return nil, fmt.Errorf("no ID specified")
	}
	if err != nil {
		return nil, err
	}

	// TODO: handle errors
	run := convertRun(reflect.ValueOf(result).Interface().(RunRow))

	return run, nil
}

func convertRun(row RunRow) *otf.Run {
	run := otf.Run{
		ID:               *row.GetRunID(),
		Timestamps:       convertTimestamps(row),
		Status:           otf.RunStatus(*row.GetStatus()),
		StatusTimestamps: convertRunStatusTimestamps(row.GetRunStatusTimestamps()),
		IsDestroy:        *row.GetIsDestroy(),
		ReplaceAddrs:     row.GetReplaceAddrs(),
		TargetAddrs:      row.GetTargetAddrs(),
		Plan:             convertPlan(row.GetPlan()),
	}

	return &run
}

func convertPlan(row PlanRow) *otf.Plan {
	return &otf.Plan{
		ID:         *row.GetPlanID(),
		Timestamps: convertTimestamps(row),
		Status:     otf.PlanStatus(*row.GetStatus()),
	}
}

func convertApply(row ApplyRow) *otf.Apply {
	return &otf.Apply{
		ID:         *row.GetApplyID(),
		Timestamps: convertTimestamps(row),
		Status:     otf.ApplyStatus(*row.GetStatus()),
	}
}

func convertTimestamps(ts Timestamps) otf.Timestamps {
	return otf.Timestamps{
		CreatedAt: ts.GetCreatedAt(),
		UpdatedAt: ts.GetUpdatedAt(),
	}
}

func convertRunStatusTimestamps(rows []RunStatusTimestamps) []otf.RunStatusTimestamp {
	timestamps := make([]otf.RunStatusTimestamp, len(rows))
	for _, r := range rows {
		timestamps = append(timestamps, otf.RunStatusTimestamp{
			Status:    otf.RunStatus(*r.Status),
			Timestamp: r.Timestamp,
		})
	}
	return timestamps
}

func convertToStringSlice(i interface{}) (s []string) {
	slice := reflect.ValueOf(i)
	for i := 0; i < slice.Len(); i++ {
		v := slice.Index(i)
		s = append(s, v.Interface().(string))
	}
	return
}
