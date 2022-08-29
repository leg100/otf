package agent

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/mount"
	dockerclient "github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/fatih/color"
	"github.com/go-logr/logr"
	"github.com/hashicorp/go-multierror"
	"github.com/leg100/otf"
)

var (
	// Environment is an implementation of an execution environment
	_ otf.Environment = (*Environment)(nil)

	ErrNonZeroExitCode = errors.New("non-zero exit code")
)

// Environment provides an execution environment for a run, providing a working
// directory, services, capturing logs etc.
type Environment struct {
	otf.Application

	logr.Logger

	// Current working directory
	path string

	// Whether cancelation has been triggered
	canceled bool

	// Cancel context func for currently running func
	cancel context.CancelFunc

	// Docker client for talking to server API
	client *dockerclient.Client

	// ID of currently running container.
	containerID string

	// CLI process output is written to this
	out io.WriteCloser

	environmentVariables []string

	// Environment context - should contain subject for authenticating to
	// services
	ctx context.Context
}

func NewEnvironment(
	logger logr.Logger,
	app otf.Application,
	runID string,
	phase otf.PhaseType,
	ctx context.Context,
	environmentVariables []string) (*Environment, error) {

	path, err := os.MkdirTemp("", "otf-plan")
	if err != nil {
		return nil, err
	}

	out := &otf.JobWriter{
		ID:         runID,
		Phase:      phase,
		Logger:     logger,
		LogService: app,
	}

	client, err := dockerclient.NewClientWithOpts(
		dockerclient.FromEnv, dockerclient.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}

	// create cancel func so that blocking tasks can be canceled
	ctx, cancel := context.WithCancel(ctx)

	return &Environment{
		Logger:               logger,
		Application:          app,
		out:                  out,
		path:                 path,
		environmentVariables: environmentVariables,
		client:               client,
		cancel:               cancel,
		ctx:                  ctx,
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

func (e *Environment) Path() string {
	return e.path
}

// Cancel terminates execution. Force controls whether termination is graceful
// or not. Performed on a best-effort basis - the func or process may have
// finished before they are cancelled, in which case only the next func or
// process will be stopped from executing.
func (e *Environment) Cancel(force bool) error {
	e.canceled = true

	e.cancelFunc(force)
	return e.cancelCLI(force)
}

// RunCLI executes a CLI process in the executor.
func (e *Environment) RunCLI(name string, args ...string) error {
	if e.canceled {
		return fmt.Errorf("execution canceled")
	}

	images, err := e.client.ImageList(e.ctx, types.ImageListOptions{
		Filters: filters.NewArgs(filters.Arg("reference", otf.DefaultTerraformImage)),
	})
	if err != nil {
		return err
	}
	if len(images) == 0 {
		_, err = e.client.ImagePull(e.ctx, otf.DefaultTerraformImage, types.ImagePullOptions{})
		if err != nil {
			return err
		}
	}

	resp, err := e.client.ContainerCreate(
		e.ctx,
		&container.Config{
			Entrypoint: []string{name},
			Image:      otf.DefaultTerraformImage,
			Cmd:        args,
			Env:        e.environmentVariables,
			Tty:        false,
			WorkingDir: "/workspace",
		},
		&container.HostConfig{
			Mounts: []mount.Mount{
				{
					Type:   mount.TypeBind,
					Source: e.path,
					Target: "/workspace",
				},
				{
					Type:   mount.TypeBind,
					Source: PluginCacheDir,
					Target: PluginCacheDir,
				},
			},
		},
		nil,
		nil,
		"",
	)
	if err != nil {
		return err
	}
	e.containerID = resp.ID

	if err := e.client.ContainerStart(e.ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		return err
	}

	out, err := e.client.ContainerLogs(e.ctx, resp.ID, types.ContainerLogsOptions{
		ShowStdout: true,
		Follow:     true,
	})
	if err != nil {
		return err
	}

	// send both stdout and stderr to our environment output, and retain a copy
	// of stderr for logging alongside any error
	stderr := new(bytes.Buffer)
	errWriter := io.MultiWriter(e.out, stderr)
	_, err = stdcopy.StdCopy(e.out, errWriter, out)
	if err != nil {
		return err
	}

	exit, errC := e.client.ContainerWait(e.ctx, resp.ID, container.WaitConditionNotRunning)
	select {
	case err := <-errC:
		if err != nil {
			return fmt.Errorf("waiting for container: %w", err)
		}
	case code := <-exit:
		if code.StatusCode != 0 {
			err := fmt.Errorf("%w: %d", ErrNonZeroExitCode, code.StatusCode)
			e.Error(err, "", "code", code.StatusCode, "stderr", stderr.String())
			return err
		}
	}

	return nil
}

// RunFunc invokes a func in the executor.
func (e *Environment) RunFunc(fn otf.EnvironmentFunc) error {
	if e.canceled {
		return fmt.Errorf("execution canceled")
	}

	if err := fn(e.ctx, e); err != nil {
		e.printRedErrorMessage(err)
		return err
	}
	return nil
}

func (e *Environment) printRedErrorMessage(err error) {
	fmt.Fprintln(e.out)

	// Print "Error:" in bright red, overriding the behaviour to disable colors
	// on a non-tty output
	red := color.New(color.FgHiRed)
	red.EnableColor()
	red.Fprint(e.out, "Error: ")

	fmt.Fprint(e.out, err.Error())
	fmt.Fprintln(e.out)
}

func (e *Environment) cancelCLI(force bool) error {
	if force {
		return e.client.ContainerKill(context.Background(), e.containerID, "KILL")
	} else {
		return e.client.ContainerKill(context.Background(), e.containerID, "INT")
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
