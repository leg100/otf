package mock

import (
	"github.com/leg100/go-tfe"
	"github.com/leg100/otf"
)

var _ otf.RunService = (*RunService)(nil)

type RunService struct {
	CreateFn       func(opts *tfe.RunCreateOptions) (*otf.Run, error)
	GetFn          func(id string) (*otf.Run, error)
	ListFn         func(opts otf.RunListOptions) (*otf.RunList, error)
	ApplyFn        func(id string, opts *tfe.RunApplyOptions) error
	DiscardFn      func(id string, opts *tfe.RunDiscardOptions) error
	CancelFn       func(id string, opts *tfe.RunCancelOptions) error
	ForceCancelFn  func(id string, opts *tfe.RunForceCancelOptions) error
	GetPlanLogsFn  func(id string, opts otf.GetChunkOptions) ([]byte, error)
	GetApplyLogsFn func(id string, opts otf.GetChunkOptions) ([]byte, error)
	EnqueuePlanFn  func(id string) error
	UpdateStatusFn func(id string, status tfe.RunStatus) (*otf.Run, error)
	UploadLogsFn   func(id string, logs []byte, opts otf.PutChunkOptions) error
	StartFn        func(id string, opts otf.JobStartOptions) (otf.Job, error)
	FinishFn       func(id string, opts otf.JobFinishOptions) (otf.Job, error)
	GetPlanJSONFn  func(id string) ([]byte, error)
	GetPlanFileFn  func(id string) ([]byte, error)
	UploadPlanFn   func(id string, plan []byte, json bool) error
}

func (s RunService) Create(opts *tfe.RunCreateOptions) (*otf.Run, error) {
	return s.CreateFn(opts)
}

func (s RunService) Get(id string) (*otf.Run, error) {
	return s.GetFn(id)
}

func (s RunService) List(opts otf.RunListOptions) (*otf.RunList, error) {
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

func (s RunService) GetPlanLogs(id string, opts otf.GetChunkOptions) ([]byte, error) {
	return s.GetPlanLogsFn(id, opts)
}

func (s RunService) GetApplyLogs(id string, opts otf.GetChunkOptions) ([]byte, error) {
	return s.GetApplyLogsFn(id, opts)
}

func (s RunService) EnqueuePlan(id string) error {
	return s.EnqueuePlanFn(id)
}

func (s RunService) UpdateStatus(id string, status tfe.RunStatus) (*otf.Run, error) {
	return s.UpdateStatusFn(id, status)
}

func (s RunService) UploadLogs(id string, logs []byte, opts otf.PutChunkOptions) error {
	return s.UploadLogsFn(id, logs, opts)
}

func (s RunService) Start(id string, opts otf.JobStartOptions) (otf.Job, error) {
	return s.StartFn(id, opts)
}

func (s RunService) Finish(id string, opts otf.JobFinishOptions) (otf.Job, error) {
	return s.FinishFn(id, opts)
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
