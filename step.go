package ots

import (
	"context"
	"io"
	"os"
	"os/exec"
)

type Step interface {
	Cancel(bool)
	Run(ctx context.Context, path string, out io.Writer) error
}

type CommandStep struct {
	cmd  string
	args []string
	proc *os.Process
}

func NewCommandStep(cmd string, args ...string) *CommandStep {
	return &CommandStep{
		cmd:  cmd,
		args: args,
	}
}

func (s *CommandStep) Cancel(force bool) {
	if s.proc == nil {
		return
	}

	if force {
		s.proc.Signal(os.Kill)
	} else {
		s.proc.Signal(os.Interrupt)
	}
}

func (s *CommandStep) Run(ctx context.Context, path string, out io.Writer) error {
	cmd := exec.Command(s.cmd, s.args...)
	cmd.Dir = path
	cmd.Stdout = out
	cmd.Stderr = out

	s.proc = cmd.Process

	return cmd.Run()
}

type FuncStep struct {
	cancel context.CancelFunc
	fn     func(context.Context, string) error
}

func NewFuncStep(fn func(context.Context, string) error) *FuncStep {
	return &FuncStep{
		fn: fn,
	}
}

func (s *FuncStep) Cancel(force bool) {
	if !force {
		return
	}
	if s.cancel == nil {
		return
	}
	s.cancel()
}

// Run invokes the func, setting the working dir to the given path
func (s *FuncStep) Run(ctx context.Context, path string, out io.Writer) error {
	ctx, s.cancel = context.WithCancel(ctx)
	return s.fn(ctx, path)
}
