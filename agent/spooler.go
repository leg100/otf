package agent

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/leg100/otf"
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

	// Watcher allows subscribing to stream of events
	Watcher

	// Logger for logging various events
	logr.Logger
}

type RunLister interface {
	List(context.Context, otf.RunListOptions) (*otf.RunList, error)
}

type Watcher interface {
	Watch(context.Context, otf.WatchOptions) (<-chan otf.Event, error)
}

type Cancelation struct {
	Run      *otf.Run
	Forceful bool
}

const (
	// SpoolerCapacity is the max number of queued runs the spooler can store
	SpoolerCapacity = 100
)

var (
	// QueuedStatuses are the list of run statuses that indicate it is in a
	// queued state
	QueuedStatuses = []otf.RunStatus{otf.RunPlanQueued, otf.RunApplyQueued}
)

// NewSpooler populates a Spooler with queued runs
func NewSpooler(svc otf.RunService, watcher Watcher, logger logr.Logger) (*SpoolerDaemon, error) {
	// retrieve existing runs, page by page
	var existing []*otf.Run
	for {
		opts := otf.RunListOptions{Statuses: QueuedStatuses}
		page, err := svc.ListRuns(otf.ContextWithAppUser(), opts)
		if err != nil {
			return nil, err
		}
		existing = append(existing, page.Items...)
		if page.NextPage() == nil {
			break
		}
		opts.PageNumber = *page.NextPage()
	}
	// svc returns runs ordered by creation date, newest first, but we want
	// oldest first, so we reverse the order
	var oldest []*otf.Run
	for _, r := range existing {
		oldest = append([]*otf.Run{r}, oldest...)
	}
	// Populate queue
	queue := make(chan *otf.Run, SpoolerCapacity)
	for _, r := range oldest {
		queue <- r
	}
	return &SpoolerDaemon{
		queue:        queue,
		cancelations: make(chan Cancelation, SpoolerCapacity),
		Watcher:      watcher,
		Logger:       logger,
	}, nil
}

// Start starts the spooler
func (s *SpoolerDaemon) Start(ctx context.Context) error {
	sub, err := s.Watch(ctx, otf.WatchOptions{})
	if err != nil {
		return err
	}
	for {
		select {
		case <-ctx.Done():
			return nil
		case event := <-sub:
			s.handleEvent(event)
		}
	}
}

// GetRun returns a channel of queued runs
func (s *SpoolerDaemon) GetRun() <-chan *otf.Run {
	return s.queue
}

// GetCancelation returns a channel of cancelation requests
func (s *SpoolerDaemon) GetCancelation() <-chan Cancelation {
	return s.cancelations
}

func (s *SpoolerDaemon) handleEvent(ev otf.Event) {
	switch obj := ev.Payload.(type) {
	case *otf.Run:
		s.V(2).Info("received run event", "run", obj.ID(), "type", ev.Type, "status", obj.Status())

		if obj.Queued() {
			s.queue <- obj
		} else if ev.Type == otf.EventRunCancel {
			s.cancelations <- Cancelation{Run: obj}
		} else if ev.Type == otf.EventRunForceCancel {
			s.cancelations <- Cancelation{Run: obj, Forceful: true}
		}
	}
}
