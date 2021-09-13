package ots

import (
	"context"
	"os"

	"github.com/go-logr/logr"
)

type Runner struct {
	steps    []Step
	current  int
	canceled bool

	lw *LogsWriter
}

type RunLogger func(id string, out []byte, opts AppendLogOptions) error

func NewRunner(steps []Step, rl RunLogger, logger logr.Logger, runID string) *Runner {
	return &Runner{
		steps: steps,
		lw: &LogsWriter{
			runLogger: rl,
			runID:     runID,
			Logger:    logger,
		},
	}
}

func (r *Runner) Cancel(force bool) {
	r.canceled = true

	if len(r.steps) > 0 {
		r.steps[r.current].Cancel(force)
	}
}

func (r *Runner) Run(ctx context.Context) error {
	defer r.lw.Close()

	path, err := os.MkdirTemp("", "ots-plan")
	if err != nil {
		return err
	}

	for i, s := range r.steps {
		if r.canceled {
			return nil
		}

		r.current = i

		if err := s.Run(ctx, path, r.lw); err != nil {
			return err
		}
	}

	return nil
}
