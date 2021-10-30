package agent

import (
	"context"
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

func (w *Worker) handle(ctx context.Context, exe Executable) {
	log := w.Logger.WithValues("job", exe.GetID())

	// Check executable in with the supervisor for duration of execution.
	w.CheckIn(run.GetID(), exe)
	defer w.CheckOut(run.GetID())

	if err := exe.Execute(); err != nil {
		log.Error(err, "run execution failed")
	}
}
