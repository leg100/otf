package notifications

import (
	"context"
	"time"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/logr"
	"gopkg.in/cenkalti/backoff.v1"
)

// lockID guarantees only one notifier on a cluster is running at any
// time.
const lockID int64 = 5577006791947779411

type (
	// notifier relays run events onto interested parties
	notifier struct {
		logr.Logger

		internal.Subscriber
	}

	NotifierOptions struct {
		logr.Logger
		internal.DB
		internal.Subscriber
	}
)

// Start constructs and initialises the notifier.
// start starts the notifier daemon. Should be invoked in a go routine.
func Start(ctx context.Context, opts NotifierOptions) error {
	ctx = internal.AddSubjectToContext(ctx, &internal.Superuser{Username: "notifier"})

	sched := &notifier{
		Logger:     opts.Logger.WithValues("component", "notifier"),
		Subscriber: opts.Subscriber,
	}
	sched.V(2).Info("started")

	op := func() error {
		// block on getting an exclusive lock
		err := opts.WaitAndLock(ctx, lockID, func() error {
			return sched.reinitialize(ctx)
		})
		if ctx.Err() != nil {
			return nil // exit
		}
		return err // retry
	}
	policy := backoff.WithContext(backoff.NewExponentialBackOff(), ctx)
	return backoff.RetryNotify(op, policy, func(err error, next time.Duration) {
		sched.Error(err, "restarting notifier")
	})
}

func (s *notifier) reinitialize(ctx context.Context) error {
	// Unsubscribe Subscribe() whenever exiting this routine.
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// subscribe to run events
	sub, err := s.Subscribe(ctx, "notifier-")
	if err != nil {
		return err
	}

	for event := range sub {
		queue <- event
	}
	return nil
}
