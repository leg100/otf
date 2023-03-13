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

var _ Spooler = (*SpoolerDaemon)(nil)

// Spooler is a daemon from which enqueued runs can be retrieved
type Spooler interface {
	// Start the daemon
	Start(context.Context) error
	// GetRun receives spooled runs
	GetRun() <-chan *run.Run
	// GetCancelation receives requests to cancel runs
	GetCancelation() <-chan Cancelation
}

// SpoolerDaemon implements Spooler, receiving runs with either a queued plan or
// apply, and converting them into spooled jobs.
type SpoolerDaemon struct {
	queue         chan *run.Run    // Queue of queued jobs
	cancelations  chan Cancelation // Queue of cancelation requests
	client.Client                  // Application for retrieving queued runs
	logr.Logger
	Config
}

type Cancelation struct {
	Run      *run.Run
	Forceful bool
}

// SpoolerCapacity is the max number of queued runs the spooler can store
const SpoolerCapacity = 100

// NewSpooler populates a Spooler with queued runs
func NewSpooler(app client.Client, logger logr.Logger, cfg Config) *SpoolerDaemon {
	return &SpoolerDaemon{
		queue:        make(chan *run.Run, SpoolerCapacity),
		cancelations: make(chan Cancelation, SpoolerCapacity),
		Client:       app,
		Logger:       logger,
		Config:       cfg,
	}
}

// Start starts the spooler
func (s *SpoolerDaemon) Start(ctx context.Context) error {
	op := func() error {
		return s.reinitialize(ctx)
	}
	policy := backoff.WithContext(backoff.NewExponentialBackOff(), ctx)
	return backoff.RetryNotify(op, policy, func(err error, next time.Duration) {
		s.Error(fmt.Errorf("%w: reconnecting in %s", err, next), "stream update")
	})
}

// GetRun returns a channel of queued runs
func (s *SpoolerDaemon) GetRun() <-chan *run.Run {
	return s.queue
}

// GetCancelation returns a channel of cancelation requests
func (s *SpoolerDaemon) GetCancelation() <-chan Cancelation {
	return s.cancelations
}

func (s *SpoolerDaemon) reinitialize(ctx context.Context) error {
	sub, err := s.Watch(ctx, otf.WatchOptions{
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
	for {
		select {
		case event, ok := <-sub:
			if !ok {
				return fmt.Errorf("watch subscription closed")
			}
			s.handleEvent(event)
		case <-ctx.Done():
			return nil
		}
	}
}

func (s *SpoolerDaemon) handleEvent(ev otf.Event) {
	switch payload := ev.Payload.(type) {
	case *run.Run:
		s.handleRun(ev.Type, payload)
	case string:
		s.Info("stream update", "info", string(payload))
	case error:
		s.Error(payload, "stream update")
	}
}

func (s *SpoolerDaemon) handleRun(event otf.EventType, run *run.Run) {
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
		s.cancelations <- Cancelation{Run: run}
	} else if event == otf.EventRunForceCancel {
		s.cancelations <- Cancelation{Run: run, Forceful: true}
	}
}
