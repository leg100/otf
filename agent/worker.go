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

func (w *Worker) handleJob(ctx context.Context, job Job) {
	log := w.Logger.WithValues("job", job.GetID())

	if err := w.RunService.Start(job.GetID(), ots.RunStartOptions{AgentID: DefaultID}); err != nil {
		log.Error(err, "starting job")
		return
	}

	env, err := ots.NewEnvironment(
		log,
		job.GetID(),
		w.RunService,
		w.ConfigurationVersionService,
		w.StateVersionService,
	)
	if err != nil {
		log.Error(err, "setting up execution environment")
		return
	}

	// Record whether job errored
	var errored bool

	log.Info("executing job", "status", job.GetStatus())

	if err := job.Do(env); err != nil {
		log.Error(err, "unable to execute job")
		return
	}

	if err := w.RunService.Finish(job.GetID(), ots.RunFinishOptions{Errored: errored}); err != nil {
		log.Error(err, "finishing job")
		return
	}

	log.Info("finished job")
}
