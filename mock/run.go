package mock

import (
	"github.com/leg100/go-tfe"
	"github.com/leg100/ots"
)

var _ ots.RunService = (*RunService)(nil)

type RunService struct {
	CreateFn            func(opts *tfe.RunCreateOptions) (*ots.Run, error)
	GetFn               func(id string) (*ots.Run, error)
	ListFn              func(workspaceID string, opts tfe.RunListOptions) (*ots.RunList, error)
	ApplyFn             func(id string, opts *tfe.RunApplyOptions) error
	DiscardFn           func(id string, opts *tfe.RunDiscardOptions) error
	CancelFn            func(id string, opts *tfe.RunCancelOptions) error
	ForceCancelFn       func(id string, opts *tfe.RunForceCancelOptions) error
	GetQueuedFn         func(opts tfe.RunListOptions) (*ots.RunList, error)
	GetPlanLogsFn       func(id string, opts ots.PlanLogOptions) ([]byte, error)
	UpdatePlanStatusFn  func(id string, status tfe.PlanStatus) (*ots.Run, error)
	UpdateApplyStatusFn func(id string, status tfe.ApplyStatus) (*ots.Run, error)
	UploadPlanLogsFn    func(id string, logs []byte) error
	FinishPlanFn        func(id string, opts ots.PlanFinishOptions) (*ots.Run, error)
}

func (s RunService) Create(opts *tfe.RunCreateOptions) (*ots.Run, error) {
	return s.CreateFn(opts)
}

func (s RunService) Get(id string) (*ots.Run, error) {
	return s.GetFn(id)
}

func (s RunService) List(workspaceID string, opts tfe.RunListOptions) (*ots.RunList, error) {
	return s.ListFn(workspaceID, opts)
}

func (s RunService) Apply(id string, opts *tfe.RunApplyOptions) error {
	return s.ApplyFn(id, opts)
}

func (s RunService) Discard(id string, opts *tfe.RunDiscardOptions) error {
	return s.DiscardFn(id, opts)
}

func (s RunService) Cancel(id string, opts *tfe.RunCancelOptions) error {
	return s.CancelFn(id, opts)
}

func (s RunService) ForceCancel(id string, opts *tfe.RunForceCancelOptions) error {
	return s.ForceCancelFn(id, opts)
}

func (s RunService) GetQueued(opts tfe.RunListOptions) (*ots.RunList, error) {
	return s.GetQueuedFn(opts)
}

func (s RunService) GetPlanLogs(id string, opts ots.PlanLogOptions) ([]byte, error) {
	return s.GetPlanLogsFn(id, opts)
}

func (s RunService) UpdatePlanStatus(id string, status tfe.PlanStatus) (*ots.Run, error) {
	return s.UpdatePlanStatusFn(id, status)
}

func (s RunService) UpdateApplyStatus(id string, status tfe.ApplyStatus) (*ots.Run, error) {
	return s.UpdateApplyStatusFn(id, status)
}

func (s RunService) UploadPlanLogs(id string, logs []byte) error {
	return s.UploadPlanLogsFn(id, logs)
}

func (s RunService) FinishPlan(id string, opts ots.PlanFinishOptions) (*ots.Run, error) {
	return s.FinishPlanFn(id, opts)
}
