package ots

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/go-logr/logr"
)

// Environment provides an execution environment for a Job.
type Environment struct {
	RunService                  RunService
	ConfigurationVersionService ConfigurationVersionService
	StateVersionService         StateVersionService

	// Current working directory
	Path string

	// Whether cancelation has been triggered
	canceled bool

	// Cancel context func for currently running func
	cancel context.CancelFunc

	// Current process
	proc *os.Process

	// CLI process output is written to this
	out io.WriteCloser

	// logger
	logr.Logger

	// agentID is the ID of the agent hosting the execution environment
	agentID string
}

// EnvironmentFunc is a func that can be invoked in the environment
type EnvironmentFunc func(context.Context, *Environment) error

// NewEnvironment constructs an Environment.
func NewEnvironment(logger logr.Logger, rs RunService, cvs ConfigurationVersionService, svs StateVersionService, agentID string) (*Environment, error) {
	path, err := os.MkdirTemp("", "ots-plan")
	if err != nil {
		return nil, err
	}

	return &Environment{
		RunService:                  rs,
		ConfigurationVersionService: cvs,
		StateVersionService:         svs,
		Path:                        path,
		agentID:                     agentID,
		Logger:                      logger,
	}, nil
}

// Execute executes the job in the environment.
func (e *Environment) Execute(job Job) (err error) {
	job, err = e.RunService.Start(job.GetID(), JobStartOptions{AgentID: e.agentID})
	if err != nil {
		return fmt.Errorf("unable to start job: %w", err)
	}

	e.out = &LogsWriter{
		runID:     job.GetID(),
		runLogger: e.RunService.UploadLogs,
		Logger:    e.Logger,
	}

	// Record whether job errored
	var errored bool

	e.Info("executing job", "status", job.GetStatus())

	if err := job.Do(e); err != nil {
		errored = true
		e.Error(err, "unable to execute job")
	}

	if err := e.out.Close(); err != nil {
		errored = true
		e.Error(err, "unable to finish writing logs")
	}

	// Regardless of job success, mark job as finished
	_, err = e.RunService.Finish(job.GetID(), JobFinishOptions{Errored: errored})
	if err != nil {
		e.Error(err, "finishing job")
		return err
	}

	e.Info("finished job")

	return nil
}

// Cancel terminates execution. Force controls whether termination is graceful
// or not. Performed on a best-effort basis - the func or process may have
// finished before they are cancelled, in which case only the next func or
// process will be stopped from executing.
func (e *Environment) Cancel(force bool) {
	e.canceled = true

	e.cancelCLI(force)
	e.cancelFunc(force)
}

func (e *Environment) cancelCLI(force bool) {
	if e.proc == nil {
		return
	}

	if force {
		e.proc.Signal(os.Kill)
	} else {
		e.proc.Signal(os.Interrupt)
	}
}

func (e *Environment) cancelFunc(force bool) {
	// Don't cancel a func's context unless forced
	if !force {
		return
	}
	if e.cancel == nil {
		return
	}
	e.cancel()
}

// RunCLI executes a CLI process in the environment.
func (e *Environment) RunCLI(name string, args ...string) error {
	if e.canceled {
		return fmt.Errorf("execution canceled")
	}

	cmd := exec.Command(name, args...)
	cmd.Dir = e.Path
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
