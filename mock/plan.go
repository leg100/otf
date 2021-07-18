package mock

import (
	"github.com/leg100/ots"
)

var _ ots.PlanService = (*PlanService)(nil)

type PlanService struct {
	GetFn         func(id string) (*ots.Plan, error)
	GetPlanJSONFn func(id string) ([]byte, error)
}

func (s PlanService) Get(id string) (*ots.Plan, error) {
	return s.GetFn(id)
}

func (s PlanService) GetPlanJSON(id string) ([]byte, error) {
	return s.GetPlanJSONFn(id)
}
