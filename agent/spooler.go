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
	GetCancelation() <-chan *otf.Run
}

// SpoolerDaemon implements Spooler, receiving runs with either a queued plan or
// apply, and converting them into spooled jobs.
type SpoolerDaemon struct {
	// Queue of queued jobs
	queue chan *otf.Run

	// Queue of cancelation requests
	cancelations chan *otf.Run

	// Subscriber allows subscribing to stream of events
	Subscriber

	// Logger for logging various events
	logr.Logger
}

type RunLister interface {
	List(context.Context, otf.RunListOptions) (*otf.RunList, error)
}

type Subscriber interface {
	Subscribe(id string) (otf.Subscription, error)
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

// NewSpooler is a constructor for a Spooler pre-populated with queued runs
func NewSpooler(rl RunLister, sub Subscriber, logger logr.Logger) (*SpoolerDaemon, error) {
	// TODO: order runs by created_at date
	runs, err := rl.List(context.Background(), otf.RunListOptions{Statuses: QueuedStatuses})
	if err != nil {
		return nil, err
	}

	// Populate queue
	queue := make(chan *otf.Run, SpoolerCapacity)
	for _, r := range runs.Items {
		queue <- r
	}

	return &SpoolerDaemon{
		queue:        queue,
		cancelations: make(chan *otf.Run, SpoolerCapacity),
		Subscriber:   sub,
		Logger:       logger,
	}, nil
}

// Start starts the spooler
func (s *SpoolerDaemon) Start(ctx context.Context) error {
	sub, err := s.Subscribe(DefaultID)
	if err != nil {
		return err
	}

	defer sub.Close()

	for {
		select {
		case <-ctx.Done():
			return nil
		case event := <-sub.C():
			s.handleEvent(event)
		}
	}
}

// GetRun returns a channel of queued runs
func (s *SpoolerDaemon) GetRun() <-chan *otf.Run {
	return s.queue
}

// GetCancelation returns a channel of cancelation requests
func (s *SpoolerDaemon) GetCancelation() <-chan *otf.Run {
	return s.cancelations
}

func (s *SpoolerDaemon) handleEvent(ev otf.Event) {
	switch obj := ev.Payload.(type) {
	case *otf.Run:
		s.V(2).Info("received run event", "run", obj.ID(), "type", ev.Type, "status", obj.Status())

		switch ev.Type {
		case otf.EventPlanQueued, otf.EventApplyQueued:
			s.queue <- obj
		case otf.EventRunCanceled:
			s.cancelations <- obj
		}
	}
}
