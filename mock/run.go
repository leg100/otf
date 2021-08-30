package mock

import (
	"github.com/leg100/go-tfe"
	"github.com/leg100/ots"
)

var _ ots.RunService = (*RunService)(nil)

type RunService struct {
	CreateFn          func(opts *tfe.RunCreateOptions) (*ots.Run, error)
	GetFn             func(id string) (*ots.Run, error)
	ListFn            func(opts ots.RunListOptions) (*ots.RunList, error)
	ApplyFn           func(id string, opts *tfe.RunApplyOptions) error
	DiscardFn         func(id string, opts *tfe.RunDiscardOptions) error
	CancelFn          func(id string, opts *tfe.RunCancelOptions) error
	ForceCancelFn     func(id string, opts *tfe.RunForceCancelOptions) error
	GetPlanLogsFn     func(id string, opts ots.PlanLogOptions) ([]byte, error)
	GetApplyLogsFn    func(id string, opts ots.ApplyLogOptions) ([]byte, error)
	EnqueuePlanFn     func(id string) error
	UpdateStatusFn    func(id string, status tfe.RunStatus) (*ots.Run, error)
	UploadPlanLogsFn  func(id string, logs []byte) error
	UploadApplyLogsFn func(id string, logs []byte) error
	FinishPlanFn      func(id string, opts ots.PlanFinishOptions) (*ots.Run, error)
	FinishApplyFn     func(id string, opts ots.ApplyFinishOptions) (*ots.Run, error)
	GetPlanJSONFn     func(id string) ([]byte, error)
	GetPlanFileFn     func(id string) ([]byte, error)
	UploadPlanFn      func(id string, plan []byte, json bool) error
}

func (s RunService) Create(opts *tfe.RunCreateOptions) (*ots.Run, error) {
	return s.CreateFn(opts)
}

func (s RunService) Get(id string) (*ots.Run, error) {
	return s.GetFn(id)
}

func (s RunService) List(opts ots.RunListOptions) (*ots.RunList, error) {
	return s.ListFn(opts)
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

func (s RunService) GetPlanLogs(id string, opts ots.PlanLogOptions) ([]byte, error) {
	return s.GetPlanLogsFn(id, opts)
}

func (s RunService) GetApplyLogs(id string, opts ots.ApplyLogOptions) ([]byte, error) {
	return s.GetApplyLogsFn(id, opts)
}

func (s RunService) EnqueuePlan(id string) error {
	return s.EnqueuePlanFn(id)
}

func (s RunService) UpdateStatus(id string, status tfe.RunStatus) (*ots.Run, error) {
	return s.UpdateStatusFn(id, status)
}

func (s RunService) UploadPlanLogs(id string, logs []byte) error {
	return s.UploadPlanLogsFn(id, logs)
}

func (s RunService) UploadApplyLogs(id string, logs []byte) error {
	return s.UploadApplyLogsFn(id, logs)
}

func (s RunService) FinishPlan(id string, opts ots.PlanFinishOptions) (*ots.Run, error) {
	return s.FinishPlanFn(id, opts)
}

func (s RunService) FinishApply(id string, opts ots.ApplyFinishOptions) (*ots.Run, error) {
	return s.FinishApplyFn(id, opts)
}

func (s RunService) GetPlanJSON(id string) ([]byte, error) {
	return s.GetPlanJSONFn(id)
}

func (s RunService) GetPlanFile(id string) ([]byte, error) {
	return s.GetPlanFileFn(id)
}

func (s RunService) UploadPlan(id string, plan []byte, json bool) error {
	return s.UploadPlanFn(id, plan, json)
}
