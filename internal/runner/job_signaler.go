package runner

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/leg100/otf/internal/logr"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/sql"
)

const jobSignalsChannel = "job_signals"

// jobSignaler relays cancelation signals to jobs via a postgres notification
// channel.
type jobSignaler struct {
	db     *sql.DB
	logger logr.Logger

	subscribers map[resource.TfeID]chan jobSignal
	mu          sync.Mutex // sync access to map
}

type jobSignal struct {
	JobID resource.TfeID `jsonapi:"primary,signals" json:"job_id"`
	Force bool           `jsonapi:"attribute" json:"force"`
}

func newJobSignaler(logger logr.Logger, db *sql.DB) *jobSignaler {
	return &jobSignaler{
		db:          db,
		logger:      logger.WithValues("component", "job-signaler"),
		subscribers: make(map[resource.TfeID]chan jobSignal),
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
		var signal jobSignal
		if err := json.Unmarshal([]byte(payload), &signal); err != nil {
			return fmt.Errorf("unmarshaling postgres event: %w", err)
		}

		s.mu.Lock()
		sub, ok := s.subscribers[signal.JobID]
		s.mu.Unlock()

		if ok {
			sub <- signal
			s.unsubscribe(signal.JobID)
		}
	}
	return nil
}

func (s *jobSignaler) publish(ctx context.Context, jobID resource.TfeID, force bool) error {
	err := s.db.Notify(ctx, jobSignalsChannel, jobSignal{JobID: jobID, Force: force})
	if err != nil {
		return fmt.Errorf("publishing job signal: %w", err)
	}
	return nil
}

// subscribe creates a "one-shot" subscription: it returns a channel on which
// any signals for a job with the given ID is sent to; once a single signal is
// sent the channel is closed. If the context is canceled before a signal is
// sent then the channel is closed.
func (s *jobSignaler) subscribe(ctx context.Context, jobID resource.TfeID) func() (jobSignal, error) {
	ch := make(chan jobSignal, 1)

	s.mu.Lock()
	s.subscribers[jobID] = ch
	s.mu.Unlock()

	go func() {
		<-ctx.Done()
		s.unsubscribe(jobID)
	}()

	return func() (jobSignal, error) {
		signal, ok := <-ch
		if !ok {
			// The only reason the channel closes is because the context has been
			// canceled, so return the reason for context being canceled.
			return jobSignal{}, ctx.Err()
		}
		return signal, nil
	}
}

func (s *jobSignaler) unsubscribe(jobID resource.TfeID) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if ch, ok := s.subscribers[jobID]; ok {
		delete(s.subscribers, jobID)
		close(ch)
	}
}
