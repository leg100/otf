package agent

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/go-logr/logr"
	"github.com/hashicorp/go-multierror"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/client"
	"github.com/leg100/otf/internal/logs"
	"github.com/leg100/otf/internal/run"
	"github.com/leg100/otf/internal/tokens"
	"github.com/leg100/otf/internal/variable"
	"github.com/pkg/errors"
)

// environment provides an execution environment for a run, providing a working
// directory, services, capturing logs etc.
//
// TODO: this is a pointless abstraction; refactor it out into worker's
// handle() func.
type environment struct {
	client.Client
	logr.Logger
	Downloader // Downloader for workers to download terraform cli on demand

	steps []step // sequence of steps to execute

	ctx       context.Context      // contains subject for authenticating to services
	out       io.WriteCloser       // captures CLI process output
	variables []*variable.Variable // terraform workspace variables

	*executor // executes processes
	*runner   // execute sequence of steps
	*workdir  // working directory fs for workspace
}

func newEnvironment(
	ctx context.Context,
	logger logr.Logger,
	agent *agent,
	run *run.Run,
) (*environment, error) {
	ws, err := agent.GetWorkspace(ctx, run.WorkspaceID)
	if err != nil {
		return nil, errors.Wrap(err, "retrieving workspace")
	}

	wd, err := newWorkdir(ws.WorkingDirectory)
	if err != nil {
		return nil, err
	}

	// Create token for terraform for it to authenticate with the otf registry
	// when retrieving modules and providers, and make it available to terraform
	// via an environment variable.
	//
	// NOTE: environment variable support is only available in terraform >= 1.2.0
	token, err := agent.CreateRunToken(ctx, tokens.CreateRunTokenOptions{
		Organization: &ws.Organization,
		RunID:        &run.ID,
	})
	if err != nil {
		return nil, errors.Wrap(err, "creating registry session")
	}
	envs := internal.SafeAppend(agent.envs, internal.CredentialEnv(agent.Hostname(), token))

	// retrieve workspace variables and add them to the environment
	variables, err := agent.ListVariables(ctx, run.WorkspaceID)
	if err != nil {
		return nil, errors.Wrap(err, "retrieving workspace variables")
	}
	for _, v := range variables {
		if v.Category == variable.CategoryEnv {
			ev := fmt.Sprintf("%s=%s", v.Key, v.Value)
			envs = append(envs, ev)
		}
	}

	writer := logs.NewPhaseWriter(ctx, logs.PhaseWriterOptions{
		RunID:  run.ID,
		Phase:  run.Phase(),
		Writer: agent,
	})

	env := &environment{
		Logger:     logger,
		Client:     agent,
		Downloader: agent,
		out:        writer,
		workdir:    wd,
		variables:  variables,
		ctx:        ctx,
		runner:     &runner{out: writer},
		executor: &executor{
			Config:              agent.Config,
			terraformPathFinder: agent.terraformPathFinder,
			version:             ws.TerraformVersion,
			out:                 writer,
			envs:                envs,
			workdir:             wd,
		},
	}

	env.steps = buildSteps(env, run)

	return env, nil
}

// execute executes a phase and regardless of whether it fails, it'll close its
// logs.
func (e *environment) execute() (err error) {
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
		fmt.Fprintf(e.out, "External agent: %t\n", e.External)
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
func (e *environment) cancel(force bool) {
	e.runner.cancel(force)
	e.executor.cancel(force)
}
