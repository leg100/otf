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
		case run := <-w.GetRun():
			w.handle(ctx, run)
		case <-ctx.Done():
			return
		}
	}
}

// handle actually executes the Run job
func (w *Worker) handle(ctx context.Context, run *otf.Run) {
	job, js, err := w.GetJob(run)
	if err != nil {
		w.Error(err, "getting job for run", "id", run.GetID())
		return
	}

	log := w.Logger.WithValues("job", job.GetID())

	env, err := NewEnvironment(
		log,
		w.RunService,
		w.ConfigurationVersionService,
		w.StateVersionService,
		js,
		job,
		w.environmentVariables,
	)
	if err != nil {
		log.Error(err, "unable to create execution environment")
		return
	}

	// Claim the job before proceeding in case another agent has claimed it.
	err := js.Claim(context.Background(), job.GetID(), otf.JobClaimOptions{AgentID: DefaultID})
	if err != nil {
		log.Error(err, "unable to start job")
		return
	}

	// Check run in with the supervisor so that it can cancel the run if a
	// cancelation request arrives
	w.CheckIn(job.GetID(), env)
	defer w.CheckOut(job.GetID())

	log.Info("executing job", "status", job.GetStatus())

	var finishOptions otf.JobFinishOptions

	if err := env.Execute(job); err != nil {
		log.Error(err, "executing job")
		finishOptions.Errored = true
	}

	// Regardless of job success, mark job as finished
	_, err = js.Finish(context.Background(), job.GetID(), finishOptions)
	if err != nil {
		log.Error(err, "finishing job")
	}

	log.Info("finished job")
}
