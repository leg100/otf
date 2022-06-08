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
	otf.EventService
	logr.Logger
}

func NewPlanService(db *sql.DB, logger logr.Logger, es otf.EventService) (*PlanService, error) {
	return &PlanService{
		db:           db,
		EventService: es,
		Logger:       logger,
	}, nil
}

func (s PlanService) Get(ctx context.Context, id string) (*otf.Plan, error) {
	run, err := s.db.GetPlan(ctx, otf.RunGetOptions{PlanID: &id})
	if err != nil {
		return nil, err
	}
	return run.Plan, nil
}

func (s PlanService) Claim(ctx context.Context, planID string, opts otf.JobClaimOptions) (otf.Job, error) {
	run, err := s.db.UpdateStatus(ctx, otf.RunGetOptions{PlanID: &planID}, func(run *otf.Run) error {
		return run.Plan.Start(run)
	})
	if err != nil {
		s.Error(err, "starting plan", "plan_id", planID)
		return nil, err
	}

	s.V(0).Info("started plan", "id", run.ID())

	return run, nil
}

// Finish marks a plan as having finished.  An event is emitted to notify any
// subscribers of the new state.
func (s PlanService) Finish(ctx context.Context, planID string, opts otf.JobFinishOptions) (otf.Job, error) {
	var event *otf.Event

	run, err := s.db.UpdateStatus(ctx, otf.RunGetOptions{PlanID: &planID}, func(run *otf.Run) (err error) {
		event, err = run.Plan.Finish(opts)
		return err
	})
	if err != nil {
		s.Error(err, "finishing plan", "id", planID)
		return nil, err
	}

	s.V(0).Info("finished plan", "id", planID)

	s.Publish(*event)

	return run, nil
}
