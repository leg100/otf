package remoteops

import (
	"context"
	"errors"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/run"
)

// worker sequentially executes runs.
type worker struct {
	*daemon
}

// Start starts the worker which waits for runs to execute.
func (w *worker) Start(ctx context.Context) {
	for {
		select {
		case job := <-w.spooler.getRun():
			w.handle(ctx, job)
		case <-ctx.Done():
			return
		}
	}
}

// handle executes the incoming run
func (w *worker) handle(ctx context.Context, r *run.Run) {
	log := w.Logger.WithValues("run", r.ID, "phase", r.Phase())

	// claim run phase
	r, err := w.StartPhase(ctx, r.ID, r.Phase(), run.PhaseStartOptions{AgentID: DefaultID})
	if errors.Is(err, internal.ErrPhaseAlreadyStarted) {
		// another agent has already claimed it
		return
	} else if err != nil {
		log.Error(err, "starting phase")
		return
	}

	env, err := newEnvironment(
		ctx,
		log,
		w.daemon,
		r,
	)
	if err != nil {
		log.Error(err, "creating execution environment")
		return
	}
	defer env.close()

	// Check run in with the terminator so that it can cancel the run if a
	// cancelation request arrives
	w.checkIn(r.ID, env)
	defer w.checkOut(r.ID)

	var finishOptions run.PhaseFinishOptions

	log.Info("executing phase")

	if err := env.execute(); err != nil {
		log.Error(err, "executing phase")
		finishOptions.Errored = true
	}

	log.Info("finishing phase")

	// Regardless of success, mark phase as finished
	_, err = w.FinishPhase(ctx, r.ID, r.Phase(), finishOptions)
	if err != nil {
		log.Error(err, "finishing phase")
		return
	}
}
