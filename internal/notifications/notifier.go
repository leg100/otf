package notifications

import (
	"context"
	"fmt"
	"time"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/logr"
	"github.com/leg100/otf/internal/run"
	"github.com/leg100/otf/internal/workspace"
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
		workspace.WorkspaceService // for retrieving workspace name

		*cache
	}

	// notifierOptions are options for constructing a notifier
	notifierOptions struct {
		logr.Logger
		internal.Subscriber
		workspace.WorkspaceService // for retrieving workspace name

		db *pgdb
	}
)

// start the notifier daemon. Should be started in a go-routine.
func start(ctx context.Context, opts notifierOptions) error {
	ctx = internal.AddSubjectToContext(ctx, &internal.Superuser{Username: "notifier"})

	sched := &notifier{
		Logger:           opts.Logger.WithValues("component", "notifier"),
		Subscriber:       opts.Subscriber,
		WorkspaceService: opts.WorkspaceService,
	}
	sched.V(2).Info("started")

	op := func() error {
		// block on getting an exclusive lock
		err := opts.db.WaitAndLock(ctx, lockID, func() error {
			return sched.reinitialize(ctx, opts.db)
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

func (s *notifier) reinitialize(ctx context.Context, db *pgdb) error {
	// Unsubscribe Subscribe() whenever exiting this routine.
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// subscribe to both run events and notification config events
	sub, err := s.Subscribe(ctx, "notifier-")
	if err != nil {
		return err
	}

	// populate cache with existing notification configs
	cache, err := newCache(ctx, db, &defaultFactory{})
	if err != nil {
		return err
	}
	s.cache = cache

	// block on handling events
	for event := range sub {
		if err := s.handle(ctx, event); err != nil {
			s.Error(err, "handling event", "event", event.Type)
		}
	}
	return nil
}

func (s *notifier) handle(ctx context.Context, event internal.Event) error {
	switch payload := event.Payload.(type) {
	case *run.Run:
		return s.handleRun(ctx, payload)
	case *Config:
		return s.handleConfig(ctx, payload, event.Type)
	default:
		return nil
	}
}

func (s *notifier) handleConfig(ctx context.Context, cfg *Config, eventType internal.EventType) error {
	switch eventType {
	case internal.CreatedEvent:
		return s.add(cfg)
	case internal.DeletedEvent:
		return s.remove(cfg.ID)
	default:
		return nil
	}
}

func (s *notifier) handleRun(ctx context.Context, r *run.Run) error {
	if r.Queued() {
		// ignore queued events
		return nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for _, cfg := range s.configs {
		if cfg.WorkspaceID != r.WorkspaceID {
			// skip configs for other workspaces
			continue
		}
		if !cfg.Enabled {
			// skip disabled config
			continue
		}
		if len(cfg.Triggers) == 0 {
			// skip config with no triggers
			continue
		}
		trigger, matches := cfg.matchTrigger(r)
		if !matches {
			// skip config with no matching trigger
			continue
		}
		// retrieve workspace because client might want to provide workspace
		// info in the notification
		//
		// TODO: this is rather expensive. We should either:
		// (a) cache workspaces, either in the notifier or upstream; or
		// (b) add workspace info to run itself
		ws, err := s.GetWorkspace(ctx, r.WorkspaceID)
		if err != nil {
			return err
		}
		client, ok := s.clients[*cfg.URL]
		if !ok {
			// should never happen
			return fmt.Errorf("client not found for url: %s", *cfg.URL)
		}
		return client.Publish(ctx, &notification{
			run:       r,
			workspace: ws,
			trigger:   trigger,
			config:    cfg,
		})
	}
	return nil
}
