package agent

import (
	"context"

	"github.com/leg100/ots"
)

// Worker sequentially executes jobs on behalf of a supervisor.
type Worker struct {
	*Supervisor
}

// Start starts the worker which waits for jobs to execute.
func (w *Worker) Start(ctx context.Context) {
	for {
		select {
		case job := <-w.GetJob():
			w.handleJob(ctx, job)
		case <-ctx.Done():
			return
		}
	}
}

func (w *Worker) handleJob(ctx context.Context, job ots.Job) {
	log := w.Logger.WithValues("job", job.GetID())

	env, err := ots.NewExecutor(
		log,
		w.RunService,
		w.ConfigurationVersionService,
		w.StateVersionService,
		DefaultID,
	)
	if err != nil {
		log.Error(err, "unable to create execution environment")
		return
	}

	// Check job in with the supervisor for duration of job.
	w.CheckIn(job.GetID(), env)
	defer w.CheckOut(job.GetID())

	if err := env.Execute(job); err != nil {
		log.Error(err, "job execution failed")
	}
}
