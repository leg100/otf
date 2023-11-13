package remoteops

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/go-logr/logr"
	"github.com/hashicorp/go-multierror"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/logs"
	"github.com/leg100/otf/internal/releases"
	otfrun "github.com/leg100/otf/internal/run"
	"github.com/leg100/otf/internal/variable"
	"github.com/pkg/errors"
)

// operation is a piece of work corresponding to a particular run phase, e.g. a
// plan, apply etc.
type operation struct {
	client
	logr.Logger
	releases.Downloader

	steps []step // sequence of steps to execute

	ctx       context.Context      // contains subject for authenticating to services
	out       io.WriteCloser       // captures CLI process output
	variables []*variable.Variable // terraform workspace variables

	*executor // executes processes
	*runner   // execute sequence of steps
	*workdir  // working directory fs for workspace
}

func newOperation(
	ctx context.Context,
	logger logr.Logger,
	dmon *daemon,
	run *otfrun.Run,
	envs []string,
) (*operation, error) {
	wd, err := newWorkdir("")
	if err != nil {
		return nil, err
	}

	// retrieve variables that are applicable to the operation's run
	// Create token for terraform for it to authenticate with the otf registry
	// when retrieving modules and providers, and make it available to terraform
	// via an environment variable.
	//
	// NOTE: environment variable support is only available in terraform >= 1.2.0
	token, err := dmon.CreateRunToken(ctx, otfrun.CreateRunTokenOptions{
		Organization: &run.Organization,
	})
	if err != nil {
		return nil, errors.Wrap(err, "creating run token")
	}
	envs := internal.SafeAppend(dmon.envs, internal.CredentialEnv(dmon.Hostname(), token))

	// retrieve variables and add them to the environment
	variables, err := dmon.ListEffectiveVariables(ctx, run.ID)
	if err != nil {
		return nil, fmt.Errorf("retrieving workspace variables: %w", err)
	}
	// append variables that are environment variables to the list of
	// environment variables
	for _, v := range variables {
		if v.Category == variable.CategoryEnv {
			ev := fmt.Sprintf("%s=%s", v.Key, v.Value)
			envs = append(envs, ev)
		}
	}

	writer := logs.NewPhaseWriter(ctx, logs.PhaseWriterOptions{
		RunID:  run.ID,
		Phase:  run.Phase(),
		Writer: dmon,
	})

	op := &operation{
		Logger:     logger,
		Downloader: dmon.Downloader,
		client:     dmon,
		out:        writer,
		workdir:    wd,
		variables:  variables,
		ctx:        ctx,
		runner:     &runner{out: writer},
		executor: &executor{
			Config:  dmon.Config,
			version: run.TerraformVersion,
			out:     writer,
			envs:    envs,
			workdir: wd,
		},
	}

	op.steps = buildSteps(op, run)

	return op, nil
}

// execute executes a phase and regardless of whether it fails, it'll close its
// logs.
func (e *operation) execute() (err error) {
	var errors *multierror.Error

	// Dump info if in debug mode
	if e.Debug {
		hostname, err := os.Hostname()
		if err != nil {
			return err
		}
		fmt.Fprintln(e.out)
		fmt.Fprintln(e.out, "Debug mode enabled")
		fmt.Fprintln(e.out, "------------------")
		fmt.Fprintf(e.out, "Hostname: %s\n", hostname)
		fmt.Fprintf(e.out, "External agent: %t\n", e.isAgent)
		fmt.Fprintf(e.out, "Sandbox mode: %t\n", e.Sandbox)
		fmt.Fprintln(e.out, "------------------")
		fmt.Fprintln(e.out)
	}

	if err := e.processSteps(e.ctx, e.steps); err != nil {
		errors = multierror.Append(errors, err)
	}

	// Mark the logs as fully uploaded
	if err := e.out.Close(); err != nil {
		errors = multierror.Append(errors, fmt.Errorf("closing logs: %w", err))
	}

	return errors.ErrorOrNil()
}

// cancel terminates execution. Force controls whether termination is graceful
// or not. Performed on a best-effort basis - the func or process may have
// finished before they are cancelled, in which case only the next func or
// process will be stopped from executing.
func (e *operation) cancel(force bool) {
	e.runner.cancel(force)
	e.executor.cancel(force)
}
