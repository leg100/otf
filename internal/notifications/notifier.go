package notifications

import (
	"context"
	"fmt"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/logr"
	"github.com/leg100/otf/internal/pubsub"
	"github.com/leg100/otf/internal/run"
	"github.com/leg100/otf/internal/sql"
	"github.com/leg100/otf/internal/workspace"
)

// LockID guarantees only one notifier on a cluster is running at any
// time.
const LockID int64 = 5577006791947779411

type (
	// Notifier relays run events onto interested parties
	Notifier struct {
		logr.Logger
		workspace.WorkspaceService // for retrieving workspace name
		internal.HostnameService   // for including a link in the notification
		run.RunService
		NotificationService

		*cache
		db *pgdb
	}

	NotifierOptions struct {
		logr.Logger
		workspace.WorkspaceService // for retrieving workspace name
		internal.HostnameService   // for including a link in the notification
		run.RunService
		NotificationService
		*sql.DB
	}
)

func NewNotifier(opts NotifierOptions) *Notifier {
	return &Notifier{
		Logger:              opts.Logger.WithValues("component", "notifier"),
		WorkspaceService:    opts.WorkspaceService,
		HostnameService:     opts.HostnameService,
		RunService:          opts.RunService,
		NotificationService: opts.NotificationService,
		db:                  &pgdb{opts.DB},
	}
}

// Start the notifier daemon. Should be started in a go-routine.
func (s *Notifier) Start(ctx context.Context) error {
	// subscribe to both run events and notification config events
	subRuns, unsubRuns := s.SubscribeRunEvents()
	defer unsubRuns()
	subConfigs, unsubConfigs := s.WatchNotificationConfigurations()
	defer unsubConfigs()

	// populate cache with existing notification configs
	cache, err := newCache(ctx, s.db, &defaultFactory{})
	if err != nil {
		return err
	}
	s.cache = cache

	// block on handling events
	for {
		select {
		case event, ok := <-subRuns:
			if !ok {
				return pubsub.ErrSubscriptionTerminated
			}
			if err := s.handleRun(ctx, event.Payload); err != nil {
				s.Error(err, "handling event", "event", event.Type)
			}
		case event, ok := <-subConfigs:
			if !ok {
				return pubsub.ErrSubscriptionTerminated
			}
			if err := s.handleConfig(ctx, event); err != nil {
				s.Error(err, "handling event", "event", event.Type)
			}
		}
	}
}

func (s *Notifier) handleConfig(ctx context.Context, event pubsub.Event[*Config]) error {
	switch event.Type {
	case pubsub.CreatedEvent:
		return s.add(event.Payload)
	case pubsub.UpdatedEvent:
		if err := s.remove(event.Payload.ID); err != nil {
			return err
		}
		return s.add(event.Payload)
	case pubsub.DeletedEvent:
		return s.remove(event.Payload.ID)
	default:
		return nil
	}
}

func (s *Notifier) handleRun(ctx context.Context, r *run.Run) error {
	if r.Queued() {
		// ignore queued events
		return nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	var ws *workspace.Workspace
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
		// Retrieve workspace if not already retrieved. We do this in order to
		// furnish the notification below with further information.
		//
		// TODO: this is rather expensive. We should either:
		// (a) cache workspaces, either in the notifier or upstream; or
		// (b) add workspace info to run itself
		if ws == nil {
			var err error
			ws, err = s.GetWorkspace(ctx, r.WorkspaceID)
			if err != nil {
				return err
			}
		}
		client, ok := s.clients[*cfg.URL]
		if !ok {
			// should never happen
			return fmt.Errorf("client not found for url: %s", *cfg.URL)
		}
		msg := &notification{
			run:       r,
			workspace: ws,
			trigger:   trigger,
			config:    cfg,
			hostname:  s.Hostname(),
		}
		s.V(3).Info("publishing notification", "notification", msg)
		if err := client.Publish(ctx, msg); err != nil {
			return err
		}
	}
	return nil
}
