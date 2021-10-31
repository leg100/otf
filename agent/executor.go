package agent

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/go-logr/logr"
	"github.com/leg100/otf"
)

var _ otf.Environment = (*Environment)(nil)

// Execution is made available to the Run Job so that it can interact with OTF
// services and write to the local filesystem, use the logger, etc.

// Environment provides an execution environment for agent jobs.
type Environment struct {
	otf.JobService

	RunService                  otf.RunService
	ConfigurationVersionService otf.ConfigurationVersionService
	StateVersionService         otf.StateVersionService

	logr.Logger

	// AgentID is the ID of the agent hosting the executor
	AgentID string

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
}

// Execute performs the full lifecycle of executing a job: marking it as
// started, running the job, and then marking it as finished. Its logs are
// captured and forwarded.
func NewEnvironment(
	rs otf.RunService,
	cvs otf.ConfigurationVersionService,
	svs otf.StateVersionService,
	logger logr.Logger) *Environment {

	out := &JobWriter{
		Job:    run.Job,
		Logger: e.Logger,
		// TODO: pass in proper context
		ctx: context.Background(),
	path, err := os.MkdirTemp("", "otf-plan")
	if err != nil {
		return nil, err
	}

	e.Path, err = os.MkdirTemp("", "otf-plan")
	job, err = e.Start(job.GetID(), otf.JobStartOptions{AgentID: e.agentID})
	if err != nil {
		return err
	}

	run, err := e.Run.Job.Start(context.Background(), e.Run.Job.GetID(), JobStartOptions{AgentID: e.AgentID})
	if err != nil {
		return fmt.Errorf("unable to start job: %w", err)
	}

	// Record whether job errored
	var errored bool

	e.Info("executing job", "status", run.GetStatus())

	if err := run.Job.Do(e); err != nil {
		errored = true
		e.Error(err, "unable to execute job")
	}

	// Mark the logs as fully uploaded
	if err := e.out.Close(); err != nil {
		errored = true
		e.Error(err, "unable to finish writing logs")
	}

	// Regardless of job success, mark job as finished
	_, err = run.Job.Finish(context.Background(), run.Job.GetID(), JobFinishOptions{Errored: errored})
	if err != nil {
		e.Error(err, "finishing job")
		return err
	}

	e.Info("finished job")

	return &Execution{
		Executor: e,
		Run:      run,
		out:      out,
	}
}
}

func (e Environment) GetConfigurationVersionService() otf.ConfigurationVersionService {
	return e.ConfigurationVersionService
}

func (e Environment) GetStateVersionService() otf.StateVersionService {
	return e.StateVersionService
}

func (e Environment) GetRunService() otf.RunService {
	return e.RunService
}

func (e Environment) GetPath() string {
	return e.Path
}


func (e *Environment) GetID() string {
	return e.Run.Job.GetID()
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

// RunCLI executes a CLI process in the executor.
func (e *Environment) RunCLI(name string, args ...string) error {
	if e.canceled {
		return fmt.Errorf("execution canceled")
	}

	cmd := exec.Command(name, args...)
	cmd.Dir = e.Path
	cmd.Stdout = e.out

	stderr := new(bytes.Buffer)
	errWriter := io.MultiWriter(e.out, stderr)
	cmd.Stderr = errWriter

	e.proc = cmd.Process

	if err := cmd.Run(); err != nil {
		e.Error(err, "running CLI step", "stderr", stderr.String())
		return err
	}

	return nil
}

// RunFunc invokes a func in the executor.
func (e *Environment) RunFunc(fn otf.EnvironmentFunc) error {
	if e.canceled {
		return fmt.Errorf("execution canceled")
	}

	// Create and store cancel func so func's context can be canceled
	ctx, cancel := context.WithCancel(context.Background())
	e.cancel = cancel

	return fn(ctx, e)
}

func (e *Execution) cancelCLI(force bool) {
	if e.proc == nil {
		return
	}

	if force {
		e.proc.Signal(os.Kill)
	} else {
		e.proc.Signal(os.Interrupt)
	}
}

func (e *Execution) cancelFunc(force bool) {
	// Don't cancel a func's context unless forced
	if !force {
		return
	}
	if e.cancel == nil {
		return
	}
	e.cancel()
}
