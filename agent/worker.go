package agent

import (
	"context"

	"github.com/leg100/otf/run"
)

// Worker sequentially executes runs.
type Worker struct {
	*Agent
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
func (w *Worker) handle(ctx context.Context, job *run.Run) {
	log := w.Logger.WithValues("run", job.ID, "phase", job.Phase)

	// Claim run job
	job, err := w.StartPhase(ctx, job.ID, job.Phase(), run.PhaseStartOptions{AgentID: DefaultID})
	if err != nil {
		log.Error(err, "starting phase")
		return
	}

	env, err := NewEnvironment(
		ctx,
		log,
		w.Client,
		job,
		w.envs,
		w.Downloader,
		w.Config,
	)
	if err != nil {
		log.Error(err, "creating execution environment")
		return
	}
	defer env.Close()

	// Check run in with the terminator so that it can cancel the run if a
	// cancelation request arrives
	w.CheckIn(job.ID, env)
	defer w.CheckOut(job.ID)

	var finishOptions run.PhaseFinishOptions

	log.Info("executing phase")

	if err := env.Execute(); err != nil {
		log.Error(err, "executing phase")
		finishOptions.Errored = true
	}

	log.Info("finishing phase")

	// Regardless of job success, mark job as finished
	_, err = w.FinishPhase(ctx, job.ID, job.Phase(), finishOptions)
	if err != nil {
		log.Error(err, "finishing phase")
		return
	}
}
