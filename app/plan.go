package app

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/leg100/otf"
	"github.com/leg100/otf/sql"
)

var _ otf.PlanService = (*PlanService)(nil)

type PlanService struct {
	db *sql.DB
	rs otf.RunService
	logr.Logger
}

func NewPlanService(db *sql.DB, logger logr.Logger, rs otf.RunService) *PlanService {
	return &PlanService{
		db:     db,
		rs:     rs,
		Logger: logger,
	}
}

func (s PlanService) Get(ctx context.Context, planID string) (*otf.Plan, error) {
	run, err := s.db.GetRun(ctx, otf.RunGetOptions{PlanID: &planID})
	if err != nil {
		s.Error(err, "retrieving plan", "id", planID)
		return nil, err
	}
	s.V(2).Info("retrieved plan", "id", planID)
	return run.Plan, nil
}
