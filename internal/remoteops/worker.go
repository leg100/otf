package remoteops

import (
	"context"

	"github.com/leg100/otf/internal"
	otfrun "github.com/leg100/otf/internal/run"
	"github.com/leg100/otf/internal/tokens"
)

// worker handles incoming runs and spawns operations from them.
type worker struct {
	*daemon
}

// Start starts the worker which waits for incoming runs and spawns operations
// from them.
func (w *worker) Start(ctx context.Context) {
	for {
		select {
		case run := <-w.spooler.getRun():
			w.handle(ctx, run)
		case <-ctx.Done():
			return
		}
	}
}

// handle handles the incoming run and spawns an operation.
func (w *worker) handle(ctx context.Context, run *otfrun.Run) {
	logger := w.Logger.WithValues("run", run.ID, "phase", run.Phase())

	// Create token for terraform for it to authenticate with the OTF registry
	// when retrieving modules and providers, and make it available to terraform
	// via an environment variable.
	//
	// NOTE: environment variable support is only available in terraform >= 1.2.0
	token, err := w.CreateRunToken(ctx, tokens.CreateRunTokenOptions{
		Organization: &run.Organization,
		RunID:        &run.ID,
	})
	if err != nil {
		logger.Error(err, "creating run token")
	}

	op, err := newOperation(
		ctx,
		logger,
		w.daemon,
		run,
		internal.SafeAppend(w.envs, internal.CredentialEnv(w.Hostname(), token)),
	)
	if err != nil {
		logger.Error(err, "creating operation for run phase")
		return
	}
	defer op.close()

	// Check operation in with the terminator so that it can cancel the op if a
	// cancelation request arrives
	w.checkIn(run.ID, op)
	defer w.checkOut(run.ID)

	var finishOptions otfrun.PhaseFinishOptions

	logger.Info("executing operation")

	if err := op.execute(); err != nil {
		logger.Error(err, "executing operation")
		finishOptions.Errored = true
	}

	logger.Info("finishing operation")
}
