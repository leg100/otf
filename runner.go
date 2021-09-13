package ots

import (
	"bytes"
	"context"
	"os"
	"time"

	"github.com/go-logr/logr"
)

type Runner struct {
	steps    []Step
	current  int
	canceled bool

	runID     string
	runLogger RunLogger
	logr.Logger
	out *bytes.Buffer

	// logStarted indicates whether first chunk of logs have been sent.
	logStarted bool
}

type RunLogger func(id string, out []byte, opts AppendLogOptions) error

func NewRunner(steps []Step, rl RunLogger, logger logr.Logger, runID string) *Runner {
	return &Runner{
		steps:     steps,
		runID:     runID,
		runLogger: rl,
		Logger:    logger,
		out:       new(bytes.Buffer),
	}
}

func (r *Runner) Cancel(force bool) {
	r.canceled = true

	if len(r.steps) > 0 {
		r.steps[r.current].Cancel(force)
	}
}

func (r *Runner) Run(ctx context.Context) error {
	done := make(chan struct{})
	go func() {
		for {
			select {
			case <-done:
				r.writeLogs(ctx, true)
				return
			case <-time.After(time.Millisecond * 500):
				r.writeLogs(ctx, false)
			}
		}
	}()
	defer func() {
		done <- struct{}{}
	}()

	path, err := os.MkdirTemp("", "ots-plan")
	if err != nil {
		// TODO: update run status with error
		r.Error(err, "unable to create temp path")
		return err
	}

	for i, s := range r.steps {
		if r.canceled {
			return nil
		}

		r.current = i

		if err := s.Run(ctx, path, r.out); err != nil {
			return err
		}
	}

	return nil
}

func (r *Runner) writeLogs(ctx context.Context, end bool) {
	opts := AppendLogOptions{End: end}
	if !r.logStarted {
		opts.Start = true
	}

	if err := r.runLogger(r.runID, r.out.Bytes(), opts); err != nil {
		r.Error(err, "unable to write logs")
	} else {
		// Only upon success mark start chunk as having been sent.
		if !r.logStarted {
			r.logStarted = true
		}
	}
}
