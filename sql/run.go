package sql

import (
	"context"
	"fmt"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/leg100/otf"
	"github.com/leg100/otf/sql/pggen"
)

// CreateRun persists a Run to the DB.
func (db *DB) CreateRun(ctx context.Context, run *otf.Run) error {
	return db.Tx(ctx, func(tx otf.DB) error {
		_, err := db.InsertRun(ctx, pggen.InsertRunParams{
			ID:                     String(run.ID()),
			CreatedAt:              run.CreatedAt(),
			IsDestroy:              run.IsDestroy(),
			Refresh:                run.Refresh(),
			RefreshOnly:            run.RefreshOnly(),
			Status:                 String(string(run.Status())),
			ReplaceAddrs:           run.ReplaceAddrs(),
			TargetAddrs:            run.TargetAddrs(),
			ConfigurationVersionID: String(run.ConfigurationVersionID()),
			WorkspaceID:            String(run.WorkspaceID()),
		})
		if err != nil {
			return err
		}
		_, err = db.InsertJob(ctx, pggen.InsertJobParams{
			JobID:  String(run.Plan.JobID()),
			RunID:  String(run.ID()),
			Status: String(string(run.Plan.Status())),
		})
		if err != nil {
			return err
		}
		_, err = db.InsertPlan(ctx, String(run.Plan.ID()), String(run.Plan.JobID()))
		if err != nil {
			return err
		}
		_, err = db.InsertJob(ctx, pggen.InsertJobParams{
			JobID:  String(run.Apply.JobID()),
			RunID:  String(run.ID()),
			Status: String(string(run.Apply.Status())),
		})
		if err != nil {
			return err
		}
		_, err = db.InsertApply(ctx, String(run.Apply.ID()), String(run.Apply.JobID()))
		if err != nil {
			return err
		}
		if err := db.insertRunStatusTimestamp(ctx, run); err != nil {
			return fmt.Errorf("inserting run status timestamp: %w", err)
		}
		if err := db.insertJobStatusTimestamp(ctx, run.Plan); err != nil {
			return fmt.Errorf("inserting plan status timestamp: %w", err)
		}
		if err := db.insertJobStatusTimestamp(ctx, run.Apply); err != nil {
			return fmt.Errorf("inserting apply status timestamp: %w", err)
		}
		return nil
	})
}

// UpdateStatus updates the run status as well as its plan and/or apply.
func (db *DB) UpdateStatus(ctx context.Context, opts otf.RunGetOptions, fn func(*otf.Run) error) (*otf.Run, error) {
	var run *otf.Run
	err := db.Tx(ctx, func(tx otf.DB) error {
		// Get run ID first
		runID, err := db.getRunID(ctx, opts)
		if err != nil {
			return databaseError(err)
		}
		// select ...for update
		result, err := db.FindRunByIDForUpdate(ctx, runID)
		if err != nil {
			return databaseError(err)
		}
		run, err = otf.UnmarshalRunDBResult(otf.RunDBResult(result), nil)
		if err != nil {
			return err
		}

		// Make copies of statuses before update
		runStatus := run.Status()
		planStatus := run.Plan.Status()
		applyStatus := run.Apply.Status()

		if err := fn(run); err != nil {
			return err
		}

		if run.Status() != runStatus {
			var err error
			_, err = db.UpdateRunStatus(ctx, String(string(run.Status())), String(run.ID()))
			if err != nil {
				return err
			}

			if err := db.insertRunStatusTimestamp(ctx, run); err != nil {
				return err
			}
		}

		if run.Plan.Status() != planStatus {
			var err error
			_, err = db.UpdateJobStatus(ctx, String(string(run.Plan.Status())), String(run.Plan.JobID()))
			if err != nil {
				return err
			}

			if err := db.insertJobStatusTimestamp(ctx, run.Plan); err != nil {
				return err
			}
		}

		if run.Apply.Status() != applyStatus {
			var err error
			_, err = db.UpdateJobStatus(ctx, String(string(run.Apply.Status())), String(run.Apply.JobID()))
			if err != nil {
				return err
			}

			if err := db.insertJobStatusTimestamp(ctx, run.Apply); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return run, nil
}

func (db *DB) CreatePlanReport(ctx context.Context, planID string, report otf.ResourceReport) error {
	_, err := db.UpdatePlannedChangesByID(ctx, pggen.UpdatePlannedChangesByIDParams{
		PlanID:       String(planID),
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
		ApplyID:      String(applyID),
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
		OrganizationNames: []string{organizationName},
		WorkspaceNames:    []string{workspaceName},
		WorkspaceIds:      []string{workspaceID},
		Statuses:          statuses,
		Limit:             opts.GetLimit(),
		Offset:            opts.GetOffset(),
	})
	db.CountRunsBatch(batch, pggen.CountRunsParams{
		OrganizationNames: []string{organizationName},
		WorkspaceNames:    []string{workspaceName},
		WorkspaceIds:      []string{workspaceID},
		Statuses:          statuses,
	})
	if includeWorkspace(opts.Include) {
		if opts.WorkspaceID != nil {
			db.FindWorkspaceByIDBatch(batch, false,
				String(*opts.WorkspaceID))
		} else if opts.OrganizationName != nil && opts.WorkspaceName != nil {
			db.FindWorkspaceByNameBatch(batch, pggen.FindWorkspaceByNameParams{
				Name:             String(*opts.WorkspaceName),
				OrganizationName: String(*opts.OrganizationName),
			})
		} else {
			return nil, fmt.Errorf("cannot include workspace without specifying workspace")
		}
	}

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

	var ws *otf.Workspace
	if includeWorkspace(opts.Include) {
		if opts.WorkspaceID != nil {
			result, err := db.FindWorkspaceByIDScan(results)
			if err != nil {
				return nil, err
			}
			ws, err = otf.UnmarshalWorkspaceDBResult(otf.WorkspaceDBResult(result))
			if err != nil {
				return nil, err
			}
		} else if opts.OrganizationName != nil && opts.WorkspaceName != nil {
			result, err := db.FindWorkspaceByNameScan(results)
			if err != nil {
				return nil, err
			}
			ws, err = otf.UnmarshalWorkspaceDBResult(otf.WorkspaceDBResult(result))
			if err != nil {
				return nil, err
			}
		}
	}

	var items []*otf.Run
	for _, r := range rows {
		run, err := otf.UnmarshalRunDBResult(otf.RunDBResult(r), ws)
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
	result, err := db.FindRunByID(ctx, runID)
	if err != nil {
		return nil, databaseError(err)
	}
	var ws *otf.Workspace
	if includeWorkspace(opts.Include) {
		result, err := db.FindWorkspaceByID(ctx, false, result.WorkspaceID)
		if err != nil {
			return nil, databaseError(err)
		}
		ws, err = otf.UnmarshalWorkspaceDBResult(otf.WorkspaceDBResult(result))
		if err != nil {
			return nil, err
		}
	}
	return otf.UnmarshalRunDBResult(otf.RunDBResult(result), ws)
}

// SetPlanFile writes a plan file to the db
func (db *DB) SetPlanFile(ctx context.Context, planID string, file []byte, format otf.PlanFormat) error {
	switch format {
	case otf.PlanFormatBinary:
		_, err := db.UpdatePlanBinByID(ctx, file, String(planID))
		return err
	case otf.PlanFormatJSON:
		_, err := db.UpdatePlanJSONByID(ctx, file, String(planID))
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

// DeleteRun deletes a run from the DB
func (db *DB) DeleteRun(ctx context.Context, id string) error {
	_, err := db.DeleteRunByID(ctx, String(id))
	return err
}

func (db *DB) getRunID(ctx context.Context, opts otf.RunGetOptions) (pgtype.Text, error) {
	if opts.PlanID != nil {
		return db.FindRunIDByPlanID(ctx, String(*opts.PlanID))
	} else if opts.ApplyID != nil {
		return db.FindRunIDByApplyID(ctx, String(*opts.ApplyID))
	} else if opts.JobID != nil {
		return db.FindRunIDByJobID(ctx, String(*opts.JobID))
	} else if opts.ID != nil {
		return String(*opts.ID), nil
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
		ID:        String(run.ID()),
		Status:    String(string(run.Status())),
		Timestamp: ts,
	})
	return err
}

func (db *DB) insertJobStatusTimestamp(ctx context.Context, job otf.Job) error {
	ts, err := job.JobStatusTimestamp(job.JobStatus())
	if err != nil {
		return err
	}
	_, err = db.InsertJobStatusTimestamp(ctx, pggen.InsertJobStatusTimestampParams{
		ID:        String(job.JobID()),
		Status:    String(string(job.JobStatus())),
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
