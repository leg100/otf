package agent

import (
	"context"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	"github.com/leg100/otf"
	"gopkg.in/cenkalti/backoff.v1"
)

var _ Spooler = (*SpoolerDaemon)(nil)

// Spooler is a daemon from which enqueued runs can be retrieved
type Spooler interface {
	// Start the daemon
	Start(context.Context) error

	// GetRun receives spooled runs
	GetRun() <-chan *otf.Run

	// GetCancelation receives requests to cancel runs
	GetCancelation() <-chan Cancelation
}

// SpoolerDaemon implements Spooler, receiving runs with either a queued plan or
// apply, and converting them into spooled jobs.
type SpoolerDaemon struct {
	// Queue of queued jobs
	queue chan *otf.Run
	// Queue of cancelation requests
	cancelations chan Cancelation
	// Application for retrieving queued runs
	otf.Application
	// Logger for logging various events
	logr.Logger
	Config
}

type Cancelation struct {
	Run      *otf.Run
	Forceful bool
}

// SpoolerCapacity is the max number of queued runs the spooler can store
const SpoolerCapacity = 100

// NewSpooler populates a Spooler with queued runs
func NewSpooler(app otf.Application, logger logr.Logger, cfg Config) *SpoolerDaemon {
	return &SpoolerDaemon{
		queue:        make(chan *otf.Run, SpoolerCapacity),
		cancelations: make(chan Cancelation, SpoolerCapacity),
		Application:  app,
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
func (s *SpoolerDaemon) GetRun() <-chan *otf.Run {
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

	listOpts := otf.RunListOptions{
		Statuses:         []otf.RunStatus{otf.RunPlanQueued, otf.RunApplyQueued},
		OrganizationName: s.Organization,
	}

	// retrieve existing runs, page by page
	var existing []*otf.Run
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
	switch obj := ev.Payload.(type) {
	case *otf.Run:
		switch obj.ExecutionMode() {
		case otf.LocalExecutionMode:
			// agents never handle runs belonging to workspaces configured to
			// use local execution (it shouldn't be possible for such a run to
			// be created in the first place...)
			s.V(2).Info("ignoring run event", "run", obj.ID(), "execution-mode", obj.ExecutionMode())
			return
		case otf.RemoteExecutionMode:
			if s.External {
				// external agents only handle runs belonging to workspaces
				// configured to use agent mode
				s.V(2).Info("ignoring run event", "run", obj.ID(), "execution-mode", obj.ExecutionMode())
				return
			}
		case otf.AgentExecutionMode:
			if !s.External {
				// internal agents only handle runs belonging to workspaces
				// configured to use remote mode.
				s.V(2).Info("ignoring run event", "run", obj.ID(), "execution-mode", obj.ExecutionMode())
				return
			}
		default:
			// unknown execution mode
			return
		}

		if obj.Queued() {
			s.queue <- obj
		} else if ev.Type == otf.EventRunCancel {
			s.cancelations <- Cancelation{Run: obj}
		} else if ev.Type == otf.EventRunForceCancel {
			s.cancelations <- Cancelation{Run: obj, Forceful: true}
		}
	case string:
		s.Info("stream update", "info", string(obj))
	case error:
		s.Error(obj, "stream update")
	}
}
