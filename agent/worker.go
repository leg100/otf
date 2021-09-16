package agent

import (
	"context"

	"github.com/leg100/ots"
)

type Worker struct {
	*Supervisor
}

func (w *Worker) Start(ctx context.Context) {
	for job := range w.GetJob() {
		w.handleJob(ctx, job)
	}
}

func (w *Worker) handleJob(ctx context.Context, job ots.Job) {
	log := w.Logger.WithValues("job", job.GetID())

	env, err := ots.NewEnvironment(
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

	if err := env.Execute(job); err != nil {
		log.Error(err, "job execution failed")
	}
}
