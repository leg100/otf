package notifications

import (
	"context"
	"time"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/logr"
	"github.com/leg100/otf/internal/run"
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
		NotificationService
	}

	// NotifierOptions are options for constructing a notifier
	NotifierOptions struct {
		logr.Logger
		internal.DB
		internal.Subscriber
		NotificationService
	}
)

// Start the notifier daemon. Should be started in a go-routine.
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
		if err := s.handle(ctx, event); err != nil {
			s.Error(err, "handling event", event.Type)
		}
	}
	return nil
}

func (s *notifier) handle(ctx context.Context, event internal.Event) error {
	run, ok := event.Payload.(*run.Run)
	if !ok {
		// ignore non-run events
		return nil
	}
	if run.Queued() {
		// ignore queued events
		return nil
	}
	configs, err := s.ListNotificationConfigurations(ctx, run.WorkspaceID)
	if err != nil {
		return err
	}
	for _, cfg := range configs {
		if !cfg.Enabled {
			// skip disabled config
			continue
		}
		if len(cfg.Triggers) == 0 {
			// skip config with no triggers
			continue
		}
		if !cfg.isTriggered(run) {
			// skip config with no matching triggers
			continue
		}
	}
	return nil
}
