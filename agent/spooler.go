package agent

import (
	"context"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	"github.com/leg100/otf"
	"github.com/leg100/otf/client"
	"github.com/leg100/otf/run"
	"github.com/leg100/otf/workspace"
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

	listOpts := run.RunListOptions{
		Statuses:     []otf.RunStatus{otf.RunPlanQueued, otf.RunApplyQueued},
		Organization: s.Organization,
	}

	// retrieve existing runs, page by page
	var existing []*run.Run
	for {
		page, err := s.ListRuns(ctx, listOpts)
		if err != nil {
			return fmt.Errorf("retrieving queued runs: %w", err)
		}
		existing = append(existing, page.Items...)
		if page.NextPage() == nil {
			break
		}
		listOpts.PageNumber = *page.NextPage()
	}

	s.V(2).Info("retrieved queued runs", "total", len(existing))

	// spool existing runs in reverse order; ListRuns returns runs newest first,
	// whereas we want oldest first.
	for i := len(existing) - 1; i >= 0; i-- {
		s.handleEvent(otf.Event{
			Type:    otf.EventRunStatusUpdate,
			Payload: existing[i],
		})
	}
	// then spool events as they come in
	for event := range sub {
		s.handleEvent(event)
	}
	return nil
}

func (s *spoolerDaemon) handleEvent(ev otf.Event) {
	switch payload := ev.Payload.(type) {
	case *run.Run:
		s.handleRun(ev.Type, payload)
	case string:
		s.Info("stream update", "info", string(payload))
	case error:
		s.Error(payload, "stream update")
	}
}

func (s *spoolerDaemon) handleRun(event otf.EventType, run *run.Run) {
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
	} else if event == otf.EventRunCancel {
		s.cancelations <- cancelation{Run: run}
	} else if event == otf.EventRunForceCancel {
		s.cancelations <- cancelation{Run: run, Forceful: true}
	}
}
