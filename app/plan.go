package app

import (
	"github.com/leg100/ots"
)

var _ ots.PlanService = (*PlanService)(nil)

type PlanService struct {
	db ots.RunStore
}

func NewPlanService(db ots.RunStore) *PlanService {
	return &PlanService{
		db: db,
	}
}

func (s PlanService) Get(id string) (*ots.Plan, error) {
	run, err := s.db.Get(ots.RunGetOptions{PlanID: &id})
	if err != nil {
		return nil, err
	}
	return run.Plan, nil
}

// GetPlanJSON returns the JSON formatted plan file for the plan.
func (s PlanService) GetPlanJSON(id string) ([]byte, error) {
	run, err := s.db.Get(ots.RunGetOptions{PlanID: &id})
	if err != nil {
		return nil, err
	}
	return run.Plan.PlanJSON, nil
}
