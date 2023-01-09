package agent

import (
	"context"
	"fmt"

	"github.com/leg100/otf"
)

// Worker sequentially executes runs on behalf of a supervisor.
type Worker struct {
	*Supervisor
}

// Start starts the worker which waits for runs to execute.
func (w *Worker) Start(ctx context.Context) {
	for {
		select {
		case job := <-w.Spooler.GetRun():
			w.handle(ctx, job)
		case <-ctx.Done():
			return
		}
	}
}

// handle executes the incoming job
func (w *Worker) handle(ctx context.Context, run *otf.Run) {
	log := w.Logger.WithValues("run", run.ID(), "phase", run.Phase())

	log.Info("starting phase")

	// create token for terraform for it to authenticate with the otf registry
	// when retrieving modules and providers
	//
	// TODO: TF_TOKEN_* environment variables are only supported from terraform
	// v1.20 onwards. We should set that as the min version for use with otf.
	session, err := w.CreateRegistrySession(ctx, run.OrganizationName())
	if err != nil {
		log.Error(err, "creating registry session")
		return
	}
	tokenEnvVar := fmt.Sprintf("%s=%s", otf.HostnameCredentialEnv(w.Hostname()), session.Token())
	environmentVariables := append(w.environmentVariables, tokenEnvVar)

	env, err := NewEnvironment(
		log,
		w.Application,
		run,
		ctx,
		environmentVariables,
		w.Downloader,
		w.Config,
	)
	if err != nil {
		log.Error(err, "creating execution environment")
		return
	}
	defer env.Close()

	// Start the job before proceeding in case another agent has started it.
	run, err = w.StartPhase(ctx, run.ID(), run.Phase(), otf.PhaseStartOptions{AgentID: DefaultID})
	if err != nil {
		log.Error(err, "starting phase")
		return
	}

	// Check run in with the supervisor so that it can cancel the run if a
	// cancelation request arrives
	w.CheckIn(run.ID(), env)
	defer w.CheckOut(run.ID())

	var finishOptions otf.PhaseFinishOptions

	log.Info("executing phase")

	if err := env.Execute(run); err != nil {
		log.Error(err, "executing phase")
		finishOptions.Errored = true
	}

	log.Info("finishing phase")

	// Regardless of job success, mark job as finished
	_, err = w.FinishPhase(ctx, run.ID(), run.Phase(), finishOptions)
	if err != nil {
		log.Error(err, "finishing phase")
		return
	}
}
