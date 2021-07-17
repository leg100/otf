package agent

import (
	"context"

	"github.com/leg100/ots"
)

type mockProcessor struct {
	Processor

	ProcessFn func(context.Context, *ots.Run, string) error
}

func (p *mockProcessor) Process(ctx context.Context, run *ots.Run, path string) error {
	return p.ProcessFn(ctx, run, path)
}
