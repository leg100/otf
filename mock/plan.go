package mock

import (
	"github.com/leg100/otf"
)

var _ otf.PlanService = (*PlanService)(nil)

type PlanService struct {
	GetFn         func(id string) (*otf.Plan, error)
	GetPlanJSONFn func(id string) ([]byte, error)
}

func (s PlanService) Get(id string) (*otf.Plan, error) {
	return s.GetFn(id)
}

func (s PlanService) GetPlanJSON(id string) ([]byte, error) {
	return s.GetPlanJSONFn(id)
}
