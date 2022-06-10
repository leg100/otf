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
func (w *Worker) handle(ctx context.Context, job otf.Job) {
	log := w.Logger.WithValues("job", job.JobID())

	env, err := NewEnvironment(
		log,
		w.App,
		job,
		w.environmentVariables,
	)
	if err != nil {
		log.Error(err, "unable to create execution environment")
		return
	}

	// Claim the job before proceeding in case another agent has claimed it.
	job, err = w.App.JobService().Claim(ctx, job.JobID(), otf.JobClaimOptions{AgentID: DefaultID})
	if err != nil {
		log.Error(err, "unable to start job")
		return
	}

	// Check run in with the supervisor so that it can cancel the run if a
	// cancelation request arrives
	w.CheckIn(job.JobID(), env)
	defer w.CheckOut(job.JobID())

	log.Info("executing job")

	var finishOptions otf.JobFinishOptions

	if err := env.Execute(job); err != nil {
		log.Error(err, "executing job")
		finishOptions.Errored = true
	}

	// Regardless of job success, mark job as finished
	_, err = w.App.JobService().Finish(ctx, job.JobID(), finishOptions)
	if err != nil {
		log.Error(err, "finishing job")
	}

	log.Info("finished job")
}
