package agent

import (
	"context"

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
		case job := <-w.GetRun():
			w.handle(ctx, job)
		case <-ctx.Done():
			return
		}
	}
}

// handle executes the incoming job
func (w *Worker) handle(ctx context.Context, run *otf.Run) {
	log := w.Logger.WithValues("run", run.ID(), "phase", run.Phase())

	env, err := NewEnvironment(
		log,
		w.App,
		run.ID(),
		run.Phase(),
		w.environmentVariables,
	)
	if err != nil {
		log.Error(err, "creating execution environment")
		return
	}

	// Start the job before proceeding in case another agent has started it.
	run, err = w.App.RunService().Start(ctx, run.ID(), run.Phase(), otf.PhaseStartOptions{AgentID: DefaultID})
	if err != nil {
		log.Error(err, "starting phase")
		return
	}

	// Check run in with the supervisor so that it can cancel the run if a
	// cancelation request arrives
	w.CheckIn(run.ID(), env)
	defer w.CheckOut(run.ID())

	log.Info("executing phase")

	var finishOptions otf.PhaseFinishOptions

	if err := env.Execute(run); err != nil {
		log.Error(err, "executing phase")
		finishOptions.Errored = true
	}

	// Regardless of job success, mark job as finished
	_, err = w.App.RunService().Finish(ctx, run.ID(), run.Phase(), finishOptions)
	if err != nil {
		log.Error(err, "finishing phase")
	}

	log.Info("finished phase")
}
