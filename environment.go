package ots

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/go-logr/logr"
)

// Environment provides an execution environment for a Run.
type Environment struct {
	Run *Run

	RunService                  RunService
	ConfigurationVersionService ConfigurationVersionService
	StateVersionService         StateVersionService

	// Current working directory
	path string

	// Whether cancelation has been triggered
	canceled bool

	// Cancel context func for currently running func
	cancel context.CancelFunc

	// Current process
	proc *os.Process

	// CLI process output is written to this
	out io.WriteCloser
}

// EnvironmentFunc is a func that can be invoked in the environment
type EnvironmentFunc func(context.Context, *Environment) error

// NewEnvironment is a constructor for Environment.
func NewEnvironment(logger logr.Logger, run *Run, rs RunService, cvs ConfigurationVersionService, svs StateVersionService) (*Environment, error) {
	path, err := os.MkdirTemp("", "ots-plan")
	if err != nil {
		return nil, err
	}

	return &Environment{
		Run:                         run,
		RunService:                  rs,
		ConfigurationVersionService: cvs,
		StateVersionService:         svs,
		path:                        path,
		out: &LogsWriter{
			runID:     run.ID,
			runLogger: rs.UploadLogs,
			Logger:    logger,
		},
	}, nil
}

// Cancel terminates execution. Force controls whether termination is graceful
// or not.
func (e *Environment) Cancel(force bool) {
	e.canceled = true

	if e.proc == nil {
		return
	}

	if force {
		e.proc.Signal(os.Kill)
	} else {
		e.proc.Signal(os.Interrupt)
	}
}

// RunCLI executes a CLI process in the environment.
func (e *Environment) RunCLI(name string, args ...string) error {
	if e.canceled {
		return fmt.Errorf("execution canceled")
	}

	cmd := exec.Command(name, args...)
	cmd.Dir = e.path
	cmd.Stdout = e.out
	cmd.Stderr = e.out

	e.proc = cmd.Process

	return cmd.Run()
}

// RunFunc invokes a func in the environment.
func (e *Environment) RunFunc(fn EnvironmentFunc) error {
	if e.canceled {
		return fmt.Errorf("execution canceled")
	}

	// Create and store cancel func so func's context can be canceled
	ctx, cancel := context.WithCancel(context.Background())
	e.cancel = cancel

	return fn(ctx, e)
}
