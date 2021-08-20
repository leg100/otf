package agent

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/leg100/go-tfe"
	"github.com/leg100/ots"
)

type RunLister interface {
	List(ots.RunListOptions) (*ots.RunList, error)
}

// Spooler queues jobs.
type Spooler struct {
	// Queue of queued runs
	queue chan *ots.Run
	// EventService allows subscribing to stream of events
	ots.EventService
	// Logger for logging various events
	logr.Logger
}

const (
	// SpoolerCapacity is the max number of queued runs the spooler can store
	SpoolerCapacity = 100
)

var (
	// QueuedStatuses are the list of run statuses that indicate it is in a
	// queued state
	QueuedStatuses = []tfe.RunStatus{tfe.RunPlanQueued, tfe.RunApplyQueued}
)

// NewSpooler is a constructor for a Spooler pre-populated with queued runs
func NewSpooler(rl RunLister, es ots.EventService, logger logr.Logger) (*Spooler, error) {
	// TODO: order runs by created_at date
	runs, err := rl.List(ots.RunListOptions{Statuses: QueuedStatuses})
	if err != nil {
		return nil, err
	}

	// Populate queue
	queue := make(chan *ots.Run, SpoolerCapacity)
	for _, r := range runs.Items {
		queue <- r
	}

	return &Spooler{
		queue:        queue,
		EventService: es,
		Logger:       logger.WithValues("component", "spooler"),
	}, nil
}

// Start starts the spooler
func (s *Spooler) Start(ctx context.Context) error {
	sub := s.Subscribe(DefaultID)
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

// GetJob retrieves receive-only job queue
func (s *Spooler) GetJob() <-chan *ots.Run {
	return s.queue
}

func (s *Spooler) handleEvent(ev ots.Event) {
	switch obj := ev.Payload.(type) {
	case *ots.Run:
		switch ev.Type {
		case ots.PlanQueued, ots.ApplyQueued:
			s.queue <- obj
		case ots.RunCanceled:
			// TODO: forward event immediately to job supervisor
		}
	}
}
