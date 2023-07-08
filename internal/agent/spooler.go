package agent

import (
	"context"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/client"
	"github.com/leg100/otf/internal/pubsub"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/run"
	"github.com/leg100/otf/internal/workspace"
	"gopkg.in/cenkalti/backoff.v1"
)

// spoolerCapacity is the max number of queued runs the spooler can store
const spoolerCapacity = 100

var _ spooler = (*spoolerDaemon)(nil)

type (
	// spooler is a daemon from which enqueued runs can be retrieved
	spooler interface {
		// start the daemon
		start(context.Context) error
		// getRun receives spooled runs
		getRun() <-chan *run.Run
		// getCancelation receives requests to cancel runs
		getCancelation() <-chan cancelation
	}

	// spoolerDaemon implements Spooler, receiving runs with either a queued plan or
	// apply, and converting them into spooled jobs.
	spoolerDaemon struct {
		queue         chan *run.Run    // Queue of queued jobs
		cancelations  chan cancelation // Queue of cancelation requests
		client.Client                  // Application for retrieving queued runs
		logr.Logger
		Config
	}

	cancelation struct {
		Run      *run.Run
		Forceful bool
	}
)

// newSpooler populates a Spooler with queued runs
func newSpooler(app client.Client, logger logr.Logger, cfg Config) *spoolerDaemon {
	return &spoolerDaemon{
		queue:        make(chan *run.Run, spoolerCapacity),
		cancelations: make(chan cancelation, spoolerCapacity),
		Client:       app,
		Logger:       logger,
		Config:       cfg,
	}
}

// start starts the spooler
func (s *spoolerDaemon) start(ctx context.Context) error {
	op := func() error {
		return s.reinitialize(ctx)
	}
	policy := backoff.WithContext(backoff.NewExponentialBackOff(), ctx)
	return backoff.RetryNotify(op, policy, func(err error, next time.Duration) {
		s.Error(fmt.Errorf("%w: reconnecting in %s", err, next), "stream update")
	})
}

// getRun returns a channel of queued runs
func (s *spoolerDaemon) getRun() <-chan *run.Run {
	return s.queue
}

// getCancelation returns a channel of cancelation requests
func (s *spoolerDaemon) getCancelation() <-chan cancelation {
	return s.cancelations
}

func (s *spoolerDaemon) reinitialize(ctx context.Context) error {
	sub, err := s.Watch(ctx, run.WatchOptions{
		Organization: s.Organization,
	})
	if err != nil {
		return err
	}

	// retrieve all existing runs
	existing, err := resource.ListAll(func(opts resource.PageOptions) (*resource.Page[*run.Run], error) {
		return s.ListRuns(ctx, run.RunListOptions{
			PageOptions:  opts,
			Statuses:     []internal.RunStatus{internal.RunPlanQueued, internal.RunApplyQueued},
			Organization: s.Organization,
		})
	})
	if err != nil {
		return fmt.Errorf("retrieving queued runs: %w", err)
	}

	s.V(2).Info("retrieved queued runs", "total", len(existing))

	// spool existing runs in reverse order; ListRuns returns runs newest first,
	// whereas we want oldest first.
	for i := len(existing) - 1; i >= 0; i-- {
		s.handleEvent(pubsub.Event{
			Payload: existing[i],
		})
	}
	// then spool events as they come in
	for event := range sub {
		s.handleEvent(event)
	}
	return nil
}

func (s *spoolerDaemon) handleEvent(ev pubsub.Event) {
	switch payload := ev.Payload.(type) {
	case *run.Run:
		s.handleRun(ev.Type, payload)
	case string:
		s.Info("stream update", "info", string(payload))
	case error:
		s.Error(payload, "stream update")
	}
}

func (s *spoolerDaemon) handleRun(event pubsub.EventType, run *run.Run) {
	// (a) external agents only handle runs with agent execution mode
	// (b) internal agents only handle runs with remote execution mode
	// (c) if neither (a) nor (b) then skip run
	if s.External && run.ExecutionMode != workspace.AgentExecutionMode {
		return
	} else if !s.External && run.ExecutionMode != workspace.RemoteExecutionMode {
		return
	}

	if run.Queued() {
		s.queue <- run
	} else if run.Status == internal.RunCanceled {
		s.cancelations <- cancelation{Run: run}
	} else if run.Status == internal.RunForceCanceled {
		s.cancelations <- cancelation{Run: run, Forceful: true}
	}
}
