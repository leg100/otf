package runner

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"

	"github.com/leg100/otf/internal/logr"
	"github.com/leg100/otf/internal/pubsub"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/sql"
)

const jobSignalsChannel = "job_signals"

// jobSignaler relays cancelation signals to jobs via a postgres notification
// channel.
type jobSignaler struct {
	db     *sql.DB
	logger logr.Logger

	// subscriptions awaiting a signal, keyed by job ID. There can be more than
	// one subscription per job: the operation carrying out a job re-establishes
	// its subscription whenever the connection is interrupted, e.g. by a proxy
	// timing out the long-lived request, and the new subscription can be
	// established before the old one has been torn down.
	subscribers map[resource.TfeID]map[*jobSignalSubscription]struct{}
	mu          sync.Mutex // sync access to map
}

// jobSignalSubscription is a one-shot subscription awaiting a signal for a job.
type jobSignalSubscription struct {
	// signals receives the job signal. Closed when the subscription is removed.
	signals chan JobSignal
	// done is closed when the subscription is removed, releasing the go routine
	// watching the subscriber's context.
	done chan struct{}
}

type JobSignal struct {
	JobID resource.TfeID `jsonapi:"primary,signals" json:"job_id"`
	Force bool           `jsonapi:"attribute" json:"force"`
}

func newJobSignaler(logger logr.Logger, db *sql.DB) *jobSignaler {
	return &jobSignaler{
		db:          db,
		logger:      logger.WithValues("component", "job-signaler"),
		subscribers: make(map[resource.TfeID]map[*jobSignalSubscription]struct{}),
	}
}

func (s *jobSignaler) Start(ctx context.Context) error {
	// Close database listen when exiting func
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	ch, err := s.db.Listen(ctx, jobSignalsChannel)
	if err != nil {
		return fmt.Errorf("listening to postgres channel: %w", err)
	}
	return s.relay(ch)
}

func (s *jobSignaler) relay(signals <-chan string) error {
	for payload := range signals {
		var signal JobSignal
		if err := json.Unmarshal([]byte(payload), &signal); err != nil {
			return fmt.Errorf("unmarshaling postgres event: %w", err)
		}

		s.mu.Lock()
		subs := make([]*jobSignalSubscription, 0, len(s.subscribers[signal.JobID]))
		for sub := range s.subscribers[signal.JobID] {
			subs = append(subs, sub)
		}
		s.mu.Unlock()

		for _, sub := range subs {
			// signals is buffered, so this never blocks, and the subscriber
			// still receives the buffered signal after removal closes the
			// channel.
			sub.signals <- signal
			s.unsubscribe(signal.JobID, sub)
		}
	}
	return pubsub.ErrSubscriptionTerminated
}

func (s *jobSignaler) publish(ctx context.Context, jobID resource.TfeID, force bool) error {
	err := s.db.Notify(ctx, jobSignalsChannel, JobSignal{JobID: jobID, Force: force})
	if err != nil {
		return fmt.Errorf("publishing job signal: %w", err)
	}
	return nil
}

// awaitJobSignal creates a "one-shot" subscription: it returns a function which
// itself only returns once a job signal is received for a job with the given
// ID. If the context is canceled then an error is instead returned giving the
// reason for the context cancelation.
func (s *jobSignaler) awaitJobSignal(ctx context.Context, jobID resource.TfeID) func() (JobSignal, error) {
	sub := &jobSignalSubscription{
		signals: make(chan JobSignal, 1),
		done:    make(chan struct{}),
	}

	s.mu.Lock()
	if _, ok := s.subscribers[jobID]; !ok {
		s.subscribers[jobID] = make(map[*jobSignalSubscription]struct{})
	}
	s.subscribers[jobID][sub] = struct{}{}
	s.mu.Unlock()

	// Remove the subscription when the subscriber's context is canceled. The go
	// routine also exits once the subscription has been removed by other means,
	// rather than lingering for as long as the context is alive.
	go func() {
		select {
		case <-ctx.Done():
			s.unsubscribe(jobID, sub)
		case <-sub.done:
		}
	}()

	return func() (JobSignal, error) {
		signal, ok := <-sub.signals
		if !ok {
			// The channel closes when the subscription is removed, which only
			// happens without a signal having been sent when the context has
			// been canceled.
			if err := ctx.Err(); err != nil {
				return JobSignal{}, err
			}
			return JobSignal{}, errors.New("job signal subscription closed")
		}
		return signal, nil
	}
}

func (s *jobSignaler) unsubscribe(jobID resource.TfeID, sub *jobSignalSubscription) {
	s.mu.Lock()
	defer s.mu.Unlock()

	subs, ok := s.subscribers[jobID]
	if !ok {
		return
	}
	if _, ok := subs[sub]; !ok {
		// already unsubscribed
		return
	}
	delete(subs, sub)
	if len(subs) == 0 {
		delete(s.subscribers, jobID)
	}
	close(sub.signals)
	close(sub.done)
}
