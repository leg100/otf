package mock

import (
	"github.com/leg100/go-tfe"
	"github.com/leg100/ots"
)

var _ ots.PlanService = (*PlanService)(nil)

type PlanService struct {
	GetFn          func(id string) (*ots.Plan, error)
	GetLogsFn      func(id string, opts ots.PlanLogOptions) ([]byte, error)
	UpdateStatusFn func(id string, status tfe.PlanStatus) (*ots.Plan, error)
	UploadLogsFn   func(id string, logs []byte) error
	CurrentFn      func(workspaceID string) (*ots.Plan, error)
	FinishFn       func(id string, opts ots.PlanFinishOptions) (*ots.Plan, error)
}

func (s PlanService) Get(id string) (*ots.Plan, error) {
	return s.GetFn(id)
}

func (s PlanService) GetLogs(id string, opts ots.PlanLogOptions) ([]byte, error) {
	return s.GetLogsFn(id, opts)
}

func (s PlanService) UpdateStatus(id string, status tfe.PlanStatus) (*ots.Plan, error) {
	return s.UpdateStatusFn(id, status)
}

func (s PlanService) UploadLogs(id string, logs []byte) error {
	return s.UploadLogsFn(id, logs)
}

func (s PlanService) Current(workspaceID string) (*ots.Plan, error) {
	return s.CurrentFn(workspaceID)
}

func (s PlanService) Finish(id string, opts ots.PlanFinishOptions) (*ots.Plan, error) {
	return s.FinishFn(id, opts)
}
