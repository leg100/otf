package mock

import (
	"context"

	"github.com/leg100/otf"
)

var _ otf.PlanService = (*PlanService)(nil)

type PlanService struct {
	GetFn         func(id string) (*otf.Plan, error)
	GetPlanJSONFn func(id string) ([]byte, error)
	StartFn       func(ctx context.Context, id string, opts otf.JobStartOptions) (*otf.Run, error)
	FinishFn      func(ctx context.Context, id string, opts otf.JobFinishOptions) (*otf.Run, error)
	GetChunkFn    func(ctx context.Context, id string, opts otf.GetChunkOptions) ([]byte, error)
	PutChunkFn    func(ctx context.Context, id string, chunk []byte, opts otf.PutChunkOptions) error
}

func (s PlanService) Get(id string) (*otf.Plan, error)      { return s.GetFn(id) }
func (s PlanService) GetPlanJSON(id string) ([]byte, error) { return s.GetPlanJSONFn(id) }
func (s PlanService) Start(ctx context.Context, id string, opts otf.JobStartOptions) (*otf.Run, error) {
	return s.StartFn(ctx, id, opts)
}
func (s PlanService) Finish(ctx context.Context, id string, opts otf.JobFinishOptions) (*otf.Run, error) {
	return s.FinishFn(ctx, id, opts)
}
func (s PlanService) GetChunk(ctx context.Context, id string, opts otf.GetChunkOptions) ([]byte, error) {
	return s.GetChunkFn(ctx, id, opts)
}

func (s PlanService) PutChunk(ctx context.Context, id string, chunk []byte, opts otf.PutChunkOptions) error {
	return s.PutChunkFn(ctx, id, chunk, opts)
}
