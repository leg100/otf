package notifications

import (
	"context"
	"fmt"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/logr"
	"github.com/leg100/otf/internal/pubsub"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/run"
	"github.com/leg100/otf/internal/runstatus"
	"github.com/leg100/otf/internal/sql"
	"github.com/leg100/otf/internal/workspace"
)

type (
	// Notifier relays run events onto interested parties
	Notifier struct {
		logr.Logger

		workspaces    notifierWorkspaceClient
		runs          notifierRunClient
		notifications notifierNotificationClient
		system        notifierHostnameClient

		*cache
		db *pgdb
	}

	NotifierOptions struct {
		RunClient          notifierRunClient
		WorkspaceClient    notifierWorkspaceClient
		NotificationClient notifierNotificationClient

		logr.Logger
		*internal.HostnameService
		*sql.DB
	}

	notifierWorkspaceClient interface {
		Get(ctx context.Context, workspaceID resource.TfeID) (*workspace.Workspace, error)
	}

	notifierRunClient interface {
		Watch(context.Context) (<-chan pubsub.Event[*run.Event], func())
		Get(context.Context, resource.TfeID) (*run.Run, error)
	}

	notifierNotificationClient interface {
		Watch(context.Context) (<-chan pubsub.Event[*Config], func())
	}

	notifierHostnameClient interface {
		Hostname() string
	}
)

func NewNotifier(opts NotifierOptions) *Notifier {
	return &Notifier{
		Logger:        opts.Logger.WithValues("component", "notifier"),
		workspaces:    opts.WorkspaceClient,
		system:        opts.HostnameService,
		runs:          opts.RunClient,
		notifications: opts.NotificationClient,
		db:            &pgdb{opts.DB},
	}
}

// Start the notifier daemon. Should be started in a go-routine.
func (s *Notifier) Start(ctx context.Context) error {
	// subscribe to notification config events
	subRuns, unsubRuns := s.runs.Watch(ctx)
	defer unsubRuns()
	subConfigs, unsubConfigs := s.notifications.Watch(ctx)
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
			if err := s.handleRunEvent(ctx, event.Payload); err != nil {
				s.Error(err, "handling event", "event", event.Type)
			}
		case event, ok := <-subConfigs:
			if !ok {
				return pubsub.ErrSubscriptionTerminated
			}
			if err := s.handleConfigEvent(event); err != nil {
				s.Error(err, "handling event", "event", event.Type)
			}
		}
	}
}

func (s *Notifier) handleConfigEvent(event pubsub.Event[*Config]) error {
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

func (s *Notifier) handleRunEvent(ctx context.Context, runEvent *run.Event) error {
	if runstatus.Queued(runEvent.Status) {
		// ignore queued events
		return nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	var ws *workspace.Workspace
	for _, cfg := range s.configs {
		if cfg.WorkspaceID != runEvent.WorkspaceID {
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
		trigger, matches := cfg.matchTrigger(runEvent.Status)
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
			ws, err = s.workspaces.Get(ctx, runEvent.WorkspaceID)
			if err != nil {
				return err
			}
		}
		run, err := s.runs.Get(ctx, runEvent.ID)
		if err != nil {
			return err
		}
		client, ok := s.clients[*cfg.URL]
		if !ok {
			// should never happen
			return fmt.Errorf("client not found for url: %s", *cfg.URL)
		}
		msg := &notification{
			run:       run,
			workspace: ws,
			trigger:   trigger,
			config:    cfg,
			hostname:  s.system.Hostname(),
		}
		s.V(3).Info("publishing notification", "notification", msg)
		if err := client.Publish(ctx, msg); err != nil {
			return err
		}
	}
	return nil
}
