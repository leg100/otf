package agent

import (
	"context"

	"github.com/leg100/ots"
)

type mockProcessor struct {
	PlanFn  func(context.Context, *ots.Run, string) error
	ApplyFn func(context.Context, *ots.Run, string) error
}

func (p *mockProcessor) Plan(ctx context.Context, run *ots.Run, path string) error {
	return p.PlanFn(ctx, run, path)
}

func (p *mockProcessor) Apply(ctx context.Context, run *ots.Run, path string) error {
	return p.ApplyFn(ctx, run, path)
}
