package mock

import (
	"context"

	"github.com/leg100/otf"
)

var _ otf.RunService = (*RunService)(nil)

type RunService struct {
	CreateFn         func(ctx context.Context, opts otf.RunCreateOptions) (*otf.Run, error)
	GetFn            func(id string) (*otf.Run, error)
	ListFn           func(opts otf.RunListOptions) (*otf.RunList, error)
	DeleteFn         func(id string) error
	ApplyFn          func(id string, opts otf.RunApplyOptions) error
	DiscardFn        func(id string, opts otf.RunDiscardOptions) error
	CancelFn         func(id string, opts otf.RunCancelOptions) error
	ForceCancelFn    func(id string, opts otf.RunForceCancelOptions) error
	EnqueuePlanFn    func(id string) error
	GetPlanFileFn    func(ctx context.Context, spec otf.RunGetOptions, format otf.PlanFormat) ([]byte, error)
	UploadPlanFileFn func(ctx context.Context, id string, plan []byte, format otf.PlanFormat) error
}

func (s RunService) Create(ctx context.Context, opts otf.RunCreateOptions) (*otf.Run, error) {
	return s.CreateFn(ctx, opts)
}

func (s RunService) Get(ctx context.Context, id string) (*otf.Run, error) {
	return s.GetFn(id)
}

func (s RunService) List(ctx context.Context, opts otf.RunListOptions) (*otf.RunList, error) {
	return s.ListFn(opts)
}

func (s RunService) Delete(ctx context.Context, id string) error {
	return s.DeleteFn(id)
}

func (s RunService) Apply(ctx context.Context, id string, opts otf.RunApplyOptions) error {
	return s.ApplyFn(id, opts)
}

func (s RunService) Discard(ctx context.Context, id string, opts otf.RunDiscardOptions) error {
	return s.DiscardFn(id, opts)
}

func (s RunService) Cancel(ctx context.Context, id string, opts otf.RunCancelOptions) error {
	return s.CancelFn(id, opts)
}

func (s RunService) ForceCancel(ctx context.Context, id string, opts otf.RunForceCancelOptions) error {
	return s.ForceCancelFn(id, opts)
}

func (s RunService) EnqueuePlan(ctx context.Context, id string) error {
	return s.EnqueuePlanFn(id)
}

func (s RunService) GetPlanFile(ctx context.Context, spec otf.RunGetOptions, format otf.PlanFormat) ([]byte, error) {
	return s.GetPlanFileFn(ctx, spec, format)
}

func (s RunService) UploadPlanFile(ctx context.Context, id string, plan []byte, format otf.PlanFormat) error {
	return s.UploadPlanFileFn(ctx, id, plan, format)
}
