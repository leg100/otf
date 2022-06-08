package sql

import (
	"context"
	"fmt"

	"github.com/jackc/pgtype"
	"github.com/leg100/otf"
	"github.com/leg100/otf/sql/pggen"
)

func (db *DB) CreatePlan(ctx context.Context, run *otf.Plan) error {
	_, err := db.InsertPlan(ctx, pggen.InsertPlanParams{
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
		PlanID:       pgtype.Text{String: run.ID(), Status: pgtype.Present},
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
		PlanID:       pgtype.Text{String: run.ID(), Status: pgtype.Present},
		Status:       pgtype.Text{String: string(run.Apply.Status()), Status: pgtype.Present},
		Additions:    0,
		Changes:      0,
		Destructions: 0,
	})
	if err != nil {
		return err
	}
	if err := db.insertPlanStatusTimestamp(ctx, run); err != nil {
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

// GetPlan retrieves a plan from the DB
func (db *DB) GetPlan(ctx context.Context, planID string) (*otf.Plan, error) {
	pgPlanID := pgtype.Text{String: planID, Status: pgtype.Present}
	result, err := db.FindPlanByID(ctx, pgPlanID)
	if err != nil {
		return nil, databaseError(err)
	}
	return otf.UnmarshalPlanDBResult(otf.PlanDBResult(result))
}
