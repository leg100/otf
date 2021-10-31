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
		case exe := <-w.GetExecutable():
			w.handle(ctx, exe)
		case <-ctx.Done():
			return
		}
	}
}

func (w *Worker) handle(ctx context.Context, run *otf.Run) {
	log := w.Logger.WithValues("job", run.Job.GetID())

	env, err := NewEnvironment(
		w.RunService,
		w.ConfigurationVersionService,
		w.StateVersionService,
		log,
		DefaultID,
	)
	if err != nil {
		log.Error(err, "unable to create execution environment")
		return
	}

	// Check executable in with the supervisor for duration of execution.
	w.CheckIn(run.GetID(), env)
	defer w.CheckOut(run.GetID())

	if err := run.Do(env); err != nil {
		log.Error(err, "run execution failed")
	}
}
