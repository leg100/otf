package ots

import (
	"context"
	"io"
)

type Runner struct {
	steps    []Step
	current  int
	canceled bool
}

func NewRunner(steps []Step) *Runner {
	return &Runner{
		steps: steps,
	}
}

func (r *Runner) Cancel(force bool) {
	r.canceled = true

	if len(r.steps) > 0 {
		r.steps[r.current].Cancel(force)
	}
}

func (r *Runner) Run(ctx context.Context, path string, out io.Writer) error {
	for i, s := range r.steps {
		if r.canceled {
			return nil
		}

		r.current = i

		if err := s.Run(ctx, path, out); err != nil {
			return err
		}
	}

	return nil
}
