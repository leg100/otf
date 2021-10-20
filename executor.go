package otf

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/go-logr/logr"
)

// Executor executes a job, providing it with services, temp directory etc,
// capturing its stdout
type Executor struct {
	JobService

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

	// agentID is the ID of the agent hosting the execution executor
	agentID string
}

// ExecutorFunc is a func that can be invoked in the executor
type ExecutorFunc func(context.Context, *Executor) error

// NewExecutor constructs an Executor.
func NewExecutor(logger logr.Logger, rs RunService, cvs ConfigurationVersionService, svs StateVersionService, agentID string) (*Executor, error) {
	path, err := os.MkdirTemp("", "otf-plan")
	if err != nil {
		return nil, err
	}

	return &Executor{
		RunService:                  rs,
		JobService:                  rs,
		ConfigurationVersionService: cvs,
		StateVersionService:         svs,
		Path:                        path,
		agentID:                     agentID,
		Logger:                      logger,
	}, nil
}

// Execute performs the full lifecycle of executing a job: marking it as
// started, running the job, and then marking it as finished. Its logs are
// captured and forwarded.
func (e *Executor) Execute(job Job) (err error) {
	job, err = e.Start(job.GetID(), JobStartOptions{AgentID: e.agentID})
	if err != nil {
		return fmt.Errorf("unable to start job: %w", err)
	}

	e.out = &JobWriter{
		ID:              job.GetID(),
		JobLogsUploader: e.JobService,
		Logger:          e.Logger,
		// TODO: pass in proper context
		ctx: context.Background(),
	}

	// Record whether job errored
	var errored bool

	e.Info("executing job", "status", job.GetStatus())

	if err := job.Do(e); err != nil {
		errored = true
		e.Error(err, "unable to execute job")
	}

	// Mark the logs as fully uploaded
	if err := e.out.Close(); err != nil {
		errored = true
		e.Error(err, "unable to finish writing logs")
	}

	// Regardless of job success, mark job as finished
	_, err = e.Finish(job.GetID(), JobFinishOptions{Errored: errored})
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
func (e *Executor) Cancel(force bool) {
	e.canceled = true

	e.cancelCLI(force)
	e.cancelFunc(force)
}

// RunCLI executes a CLI process in the executor.
func (e *Executor) RunCLI(name string, args ...string) error {
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
func (e *Executor) RunFunc(fn ExecutorFunc) error {
	if e.canceled {
		return fmt.Errorf("execution canceled")
	}

	// Create and store cancel func so func's context can be canceled
	ctx, cancel := context.WithCancel(context.Background())
	e.cancel = cancel

	return fn(ctx, e)
}

func (e *Executor) cancelCLI(force bool) {
	if e.proc == nil {
		return
	}

	if force {
		e.proc.Signal(os.Kill)
	} else {
		e.proc.Signal(os.Interrupt)
	}
}

func (e *Executor) cancelFunc(force bool) {
	// Don't cancel a func's context unless forced
	if !force {
		return
	}
	if e.cancel == nil {
		return
	}
	e.cancel()
}
