package agent

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/leg100/otf"
)

var _ Spooler = (*SpoolerDaemon)(nil)

// Spooler is a daemon from which enqueued jobs can be retrieved
type Spooler interface {
	// Start the daemon
	Start(context.Context) error

	// GetJob receives spooled runs
	GetJob() <-chan otf.Job

	// GetCancelation receives requests to cancel runs
	GetCancelation() <-chan otf.Job
}

// SpoolerDaemon implements Spooler, receiving runs with either a queued plan or
// apply, and converting them into spooled jobs.
type SpoolerDaemon struct {
	// Queue of queued jobs
	queue chan otf.Job

	// Queue of cancelation requests
	cancelations chan otf.Job

	// Subscriber allows subscribing to stream of events
	Subscriber

	// Logger for logging various events
	logr.Logger
}

type JobQueueGetter interface {
	Queued(context.Context) ([]otf.Job, error)
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

// NewSpooler is a constructor for a Spooler pre-populated with queued jobs
func NewSpooler(rl JobQueueGetter, sub Subscriber, logger logr.Logger) (*SpoolerDaemon, error) {
	jobs, err := rl.Queued(context.Background())
	if err != nil {
		return nil, err
	}

	// Populate queue
	queue := make(chan otf.Job, SpoolerCapacity)
	for _, j := range jobs {
		queue <- j
	}

	return &SpoolerDaemon{
		queue:        queue,
		cancelations: make(chan otf.Job, SpoolerCapacity),
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
func (s *SpoolerDaemon) GetJob() <-chan otf.Job {
	return s.queue
}

// GetCancelation returns a channel of cancelation requests
func (s *SpoolerDaemon) GetCancelation() <-chan otf.Job {
	return s.cancelations
}

func (s *SpoolerDaemon) handleEvent(ev otf.Event) {
	job, ok := ev.Payload.(otf.Job)
	if !ok {
		// skip non-job events
		return
	}
	s.V(2).Info("received job event", "run", job.JobID(), "type", ev.Type, "status", job.Status())

	switch ev.Type {
	case otf.EventPlanQueued, otf.EventApplyQueued:
		s.queue <- job
	case otf.EventRunCanceled:
		s.cancelations <- job
	}
}
