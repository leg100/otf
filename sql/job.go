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

var _ otf.JobStore = (*JobDB)(nil)

type JobDB struct {
	*pgxpool.Pool
}

func NewJobDB(db *pgxpool.Pool) *JobDB {
	return &JobDB{
		Pool: db,
	}
}

// Create persists a Job to the DB.
func (db JobDB) Create(ctx context.Context, job *otf.Job) error {
	tx, err := db.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	q := pggen.NewQuerier(tx)
	_, err = q.InsertJob(ctx, pggen.InsertJobParams{
		ID:                     pgtype.Text{String: job.ID(), Status: pgtype.Present},
		CreatedAt:              job.CreatedAt(),
		IsDestroy:              job.IsDestroy(),
		Refresh:                job.Refresh(),
		RefreshOnly:            job.RefreshOnly(),
		Status:                 pgtype.Text{String: string(job.Status()), Status: pgtype.Present},
		ReplaceAddrs:           job.ReplaceAddrs(),
		TargetAddrs:            job.TargetAddrs(),
		ConfigurationVersionID: pgtype.Text{String: job.ConfigurationVersion.ID(), Status: pgtype.Present},
		WorkspaceID:            pgtype.Text{String: job.Workspace.ID(), Status: pgtype.Present},
	})
	if err != nil {
		return err
	}

	// find out which doer the job is referrring to and update its job id.

	_, err = q.InsertPlan(ctx, pggen.InsertPlanParams{
		PlanID:       pgtype.Text{String: job.Plan.ID(), Status: pgtype.Present},
		JobID:        pgtype.Text{String: job.ID(), Status: pgtype.Present},
		Status:       pgtype.Text{String: string(job.Plan.Status()), Status: pgtype.Present},
		Additions:    0,
		Changes:      0,
		Destructions: 0,
	})
	if err != nil {
		return err
	}
	_, err = q.InsertApply(ctx, pggen.InsertApplyParams{
		ApplyID:      pgtype.Text{String: job.Apply.ID(), Status: pgtype.Present},
		JobID:        pgtype.Text{String: job.ID(), Status: pgtype.Present},
		Status:       pgtype.Text{String: string(job.Apply.Status()), Status: pgtype.Present},
		Additions:    0,
		Changes:      0,
		Destructions: 0,
	})
	if err != nil {
		return err
	}
	if err := insertJobStatusTimestamp(ctx, q, job); err != nil {
		return fmt.Errorf("inserting run status timestamp: %w", err)
	}
	if err := insertPlanStatusTimestamp(ctx, q, job.Plan); err != nil {
		return fmt.Errorf("inserting plan status timestamp: %w", err)
	}
	if err := insertApplyStatusTimestamp(ctx, q, job.Apply); err != nil {
		return fmt.Errorf("inserting apply status timestamp: %w", err)
	}
	return tx.Commit(ctx)
}

func (db *DB) UpdateJobStatus(ctx context.Context, jobID string, status otf.JobStatus) (*otf.Job, error) {
	_, err := db.conn.UpdateJobStatus(ctx, jobID, status)
	if run.Apply.Status() != applyStatus {
		var err error
		_, err = q.UpdateApplyStatus(ctx,
			pgtype.Text{String: string(run.Apply.Status()), Status: pgtype.Present},
			pgtype.Text{String: run.Apply.ID(), Status: pgtype.Present},
		)
		if err != nil {
			return nil, err
		}

		if err := insertApplyStatusTimestamp(ctx, q, run.Apply); err != nil {
			return nil, err
		}
	}

	return run, tx.Commit(ctx)
}

func (db JobDB) CreatePlanReport(ctx context.Context, planID string, report otf.ResourceReport) error {
	q := pggen.NewQuerier(db.Pool)
	_, err := q.UpdatePlannedChangesByID(ctx, pggen.UpdatePlannedChangesByIDParams{
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

func (db JobDB) CreateApplyReport(ctx context.Context, applyID string, report otf.ResourceReport) error {
	q := pggen.NewQuerier(db.Pool)
	_, err := q.UpdateAppliedChangesByID(ctx, pggen.UpdateAppliedChangesByIDParams{
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

func (db JobDB) List(ctx context.Context, opts otf.JobListOptions) (*otf.JobList, error) {
	q := pggen.NewQuerier(db.Pool)
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
	q.FindJobsBatch(batch, pggen.FindJobsParams{
		OrganizationNames:           []string{organizationName},
		WorkspaceNames:              []string{workspaceName},
		WorkspaceIds:                []string{workspaceID},
		Statuses:                    statuses,
		Limit:                       opts.GetLimit(),
		Offset:                      opts.GetOffset(),
		IncludeConfigurationVersion: includeConfigurationVersion(opts.Include),
		IncludeWorkspace:            includeWorkspace(opts.Include),
	})
	q.CountJobsBatch(batch, pggen.CountJobsParams{
		OrganizationNames: []string{organizationName},
		WorkspaceNames:    []string{workspaceName},
		WorkspaceIds:      []string{workspaceID},
		Statuses:          statuses,
	})
	results := db.Pool.SendBatch(ctx, batch)
	defer results.Close()

	rows, err := q.FindJobsScan(results)
	if err != nil {
		return nil, err
	}
	count, err := q.CountJobsScan(results)
	if err != nil {
		return nil, err
	}

	var items []*otf.Job
	for _, r := range rows {
		run, err := otf.UnmarshalJobDBResult(otf.JobDBResult(r))
		if err != nil {
			return nil, err
		}
		items = append(items, run)
	}

	return &otf.JobList{
		Items:      items,
		Pagination: otf.NewPagination(opts.ListOptions, *count),
	}, nil
}

// Get retrieves a run using the get options
func (db JobDB) Get(ctx context.Context, opts otf.JobGetOptions) (*otf.Job, error) {
	q := pggen.NewQuerier(db.Pool)
	// Get run ID first
	runID, err := getJobID(ctx, q, opts)
	if err != nil {
		return nil, databaseError(err)
	}
	// ...now get full run
	result, err := q.FindJobByID(ctx, pggen.FindJobByIDParams{
		JobID:                       runID,
		IncludeConfigurationVersion: includeConfigurationVersion(opts.Include),
		IncludeWorkspace:            includeWorkspace(opts.Include),
	})
	if err != nil {
		return nil, databaseError(err)
	}
	return otf.UnmarshalJobDBResult(otf.JobDBResult(result))
}

// SetPlanFile writes a plan file to the db
func (db JobDB) SetPlanFile(ctx context.Context, planID string, file []byte, format otf.PlanFormat) error {
	q := pggen.NewQuerier(db.Pool)
	switch format {
	case otf.PlanFormatBinary:
		_, err := q.UpdatePlanBinByID(ctx,
			file,
			pgtype.Text{String: planID, Status: pgtype.Present},
		)
		return err
	case otf.PlanFormatJSON:
		_, err := q.UpdatePlanJSONByID(ctx,
			file,
			pgtype.Text{String: planID, Status: pgtype.Present},
		)
		return err
	default:
		return fmt.Errorf("unknown plan format: %s", string(format))
	}
}

// GetPlanFile retrieves a plan file for the run
func (db JobDB) GetPlanFile(ctx context.Context, runID string, format otf.PlanFormat) ([]byte, error) {
	q := pggen.NewQuerier(db.Pool)
	switch format {
	case otf.PlanFormatBinary:
		return q.GetPlanBinByID(ctx, pgtype.Text{String: runID, Status: pgtype.Present})
	case otf.PlanFormatJSON:
		return q.GetPlanJSONByID(ctx, pgtype.Text{String: runID, Status: pgtype.Present})
	default:
		return nil, fmt.Errorf("unknown plan format: %s", string(format))
	}
}

// Delete deletes a run from the DB
func (db JobDB) Delete(ctx context.Context, id string) error {
	q := pggen.NewQuerier(db.Pool)
	result, err := q.DeleteJobByID(ctx, pgtype.Text{String: id, Status: pgtype.Present})
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return otf.ErrResourceNotFound
	}

	return nil
}

func getJobID(ctx context.Context, q *pggen.DBQuerier, opts otf.JobGetOptions) (pgtype.Text, error) {
	if opts.PlanID != nil {
		return q.FindJobIDByPlanID(ctx, pgtype.Text{String: *opts.PlanID, Status: pgtype.Present})
	} else if opts.ApplyID != nil {
		return q.FindJobIDByApplyID(ctx, pgtype.Text{String: *opts.ApplyID, Status: pgtype.Present})
	} else if opts.ID != nil {
		return pgtype.Text{String: *opts.ID, Status: pgtype.Present}, nil
	} else {
		return pgtype.Text{}, fmt.Errorf("no ID specified")
	}
}

func insertJobStatusTimestamp(ctx context.Context, q *pggen.DBQuerier, run *otf.Job) error {
	ts, err := run.StatusTimestamp(run.Status())
	if err != nil {
		return err
	}
	_, err = q.InsertJobStatusTimestamp(ctx, pggen.InsertJobStatusTimestampParams{
		ID:        pgtype.Text{String: run.ID(), Status: pgtype.Present},
		Status:    pgtype.Text{String: string(run.Status()), Status: pgtype.Present},
		Timestamp: ts,
	})
	return err
}

func insertPlanStatusTimestamp(ctx context.Context, q *pggen.DBQuerier, plan *otf.Plan) error {
	ts, err := plan.StatusTimestamp(plan.Status())
	if err != nil {
		return err
	}
	_, err = q.InsertPlanStatusTimestamp(ctx, pggen.InsertPlanStatusTimestampParams{
		ID:        pgtype.Text{String: plan.ID(), Status: pgtype.Present},
		Status:    pgtype.Text{String: string(plan.Status()), Status: pgtype.Present},
		Timestamp: ts,
	})
	return err
}

func insertApplyStatusTimestamp(ctx context.Context, q *pggen.DBQuerier, apply *otf.Apply) error {
	ts, err := apply.StatusTimestamp(apply.Status())
	if err != nil {
		return err
	}
	_, err = q.InsertApplyStatusTimestamp(ctx, pggen.InsertApplyStatusTimestampParams{
		ID:        pgtype.Text{String: apply.ID(), Status: pgtype.Present},
		Status:    pgtype.Text{String: string(apply.Status()), Status: pgtype.Present},
		Timestamp: ts,
	})
	return err
}

func convertStatusSliceToStringSlice(statuses []otf.JobStatus) (s []string) {
	for _, status := range statuses {
		s = append(s, string(status))
	}
	return
}
