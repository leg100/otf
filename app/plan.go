package app

import (
	"github.com/leg100/otf"
)

var _ otf.PlanService = (*PlanService)(nil)

type PlanService struct {
	bs otf.BlobStore
	db otf.RunStore
}

func NewPlanService(db otf.RunStore, bs otf.BlobStore) *PlanService {
	return &PlanService{
		bs: bs,
		db: db,
	}
}

func (s PlanService) Get(id string) (*otf.Plan, error) {
	run, err := s.db.Get(otf.RunGetOptions{PlanID: &id})
	if err != nil {
		return nil, err
	}
	return run.Plan, nil
}

// GetPlanJSON returns the JSON formatted plan file for the plan.
func (s PlanService) GetPlanJSON(id string) ([]byte, error) {
	run, err := s.db.Get(otf.RunGetOptions{PlanID: &id})
	if err != nil {
		return nil, err
	}
	return s.bs.Get(run.Plan.PlanJSONBlobID)
}
