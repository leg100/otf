package agent

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/fatih/color"
	"github.com/go-logr/logr"
	"github.com/hashicorp/go-multierror"
	"github.com/leg100/otf"
)

// Environment is an implementation of an execution environment
var _ otf.Environment = (*Environment)(nil)

// Environment provides an execution environment for a run, providing a working
// directory, services, capturing logs etc.
type Environment struct {
	runService                  otf.RunService
	configurationVersionService otf.ConfigurationVersionService
	stateVersionService         otf.StateVersionService

	logr.Logger

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

	environmentVariables []string
}

func NewEnvironment(
	logger logr.Logger,
	app otf.Application,
	id string,
	phase otf.PhaseType,
	environmentVariables []string) (*Environment, error) {

	path, err := os.MkdirTemp("", "otf-plan")
	if err != nil {
		return nil, err
	}

	out := &otf.JobWriter{
		ID:         id,
		Phase:      phase,
		Logger:     logger,
		LogService: app.RunService(),
	}

	return &Environment{
		Logger:                      logger,
		runService:                  app.RunService(),
		configurationVersionService: app.ConfigurationVersionService(),
		stateVersionService:         app.StateVersionService(),
		out:                         out,
		path:                        path,
		environmentVariables:        environmentVariables,
	}, nil
}

// Execute executes a phase and regardless of whether it fails, it'll close the
// environment logs.
func (e *Environment) Execute(phase Doer) (err error) {
	var errors *multierror.Error

	if err := phase.Do(e); err != nil {
		errors = multierror.Append(errors, fmt.Errorf("executing phase: %w", err))
	}

	// Mark the logs as fully uploaded
	if err := e.out.Close(); err != nil {
		errors = multierror.Append(errors, fmt.Errorf("closing logs: %w", err))
	}

	return errors.ErrorOrNil()
}

func (e *Environment) ConfigurationVersionService() otf.ConfigurationVersionService {
	return e.configurationVersionService
}

func (e *Environment) StateVersionService() otf.StateVersionService {
	return e.stateVersionService
}

func (e *Environment) RunService() otf.RunService {
	return e.runService
}

func (e *Environment) Path() string {
	return e.path
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
	cmd.Dir = e.path
	cmd.Stdout = e.out
	cmd.Env = e.environmentVariables

	stderr := new(bytes.Buffer)
	errWriter := io.MultiWriter(e.out, stderr)
	cmd.Stderr = errWriter

	if err := cmd.Start(); err != nil {
		e.Error(err, "starting command", "stderr", stderr.String(), "path", e.path)
		return err
	}
	// store process so that it can be canceled
	e.proc = cmd.Process

	if err := cmd.Wait(); err != nil {
		e.Error(err, "running command", "stderr", stderr.String(), "path", e.path)
		return err
	}
	e.V(2).Info("ran command", "name", name, "args", args, "path", e.path)

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
	if err := fn(e.ctx, e); err != nil {
		e.printRedErrorMessage(err)
		return err
	}
	return nil
}

func (e *Environment) printRedErrorMessage(err error) {
	fmt.Fprintln(e.out)

	// Print "Error" in bright red, overriding the behaviour to disable
	// colors on a non-tty output
	red := color.New(color.FgHiRed)
	red.EnableColor()
	red.Fprintf(e.out, "Error: ")

	fmt.Fprintf(e.out, err.Error())
	fmt.Fprintln(e.out)
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

type Doer interface {
	// TODO: environment is excessive; can we pass in something that exposes
	// fewer methods like an 'executor'?
	Do(otf.Environment) error
}
