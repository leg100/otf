package agent

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/leg100/go-tfe"
	"github.com/leg100/otf"
)

var _ Spooler = (*SpoolerDaemon)(nil)

// Spooler is a daemon from which jobs can be retrieved
type Spooler interface {
	// Start the daemon
	Start(context.Context)

	// GetJob receives spooled job
	GetJob() <-chan otf.Job

	// GetCancelation receives cancelation request for a job
	GetCancelation() <-chan otf.Job
}

// SpoolerDaemon implements Spooler, receiving runs with either a queued plan or
// apply, and converting them into spooled jobs.
type SpoolerDaemon struct {
	// Queue of queued jobs
	queue chan otf.Job

	// Queue of cancelation requests
	cancelations chan otf.Job

	// EventService allows subscribing to stream of events
	otf.EventService

	// Logger for logging various events
	logr.Logger
}

type RunLister interface {
	List(otf.RunListOptions) (*otf.RunList, error)
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
func NewSpooler(rl RunLister, es otf.EventService, logger logr.Logger) (*SpoolerDaemon, error) {
	// TODO: order runs by created_at date
	runs, err := rl.List(otf.RunListOptions{Statuses: QueuedStatuses})
	if err != nil {
		return nil, err
	}

	// Populate queue
	queue := make(chan otf.Job, SpoolerCapacity)
	for _, r := range runs.Items {
		queue <- r
	}

	return &SpoolerDaemon{
		queue:        queue,
		cancelations: make(chan otf.Job, SpoolerCapacity),
		EventService: es,
		Logger:       logger,
	}, nil
}

// Start starts the spooler
func (s *SpoolerDaemon) Start(ctx context.Context) {
	sub := s.Subscribe(DefaultID)
	defer sub.Close()

	for {
		select {
		case <-ctx.Done():
			return
		case event := <-sub.C():
			s.handleEvent(event)
		}
	}
}

// GetJob returns a channel of queued jobs
func (s *SpoolerDaemon) GetJob() <-chan otf.Job {
	return s.queue
}

// GetCancelation returns a channel of cancelation requests
func (s *SpoolerDaemon) GetCancelation() <-chan otf.Job {
	return s.cancelations
}

func (s *SpoolerDaemon) handleEvent(ev otf.Event) {
	switch obj := ev.Payload.(type) {
	case *otf.Run:
		s.Info("run event received", "run", obj.ID, "type", ev.Type, "status", obj.Status)

		switch ev.Type {
		case otf.PlanQueued, otf.ApplyQueued:
			s.queue <- obj
		case otf.RunCanceled:
			s.cancelations <- obj
		}
	}
}
