package repo

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/google/uuid"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/cloud"
	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/pubsub"
	"github.com/leg100/otf/internal/sql"
	"github.com/leg100/otf/internal/sql/pggen"
	"github.com/leg100/otf/internal/vcsprovider"
)

type (
	RepoService = Service

	// Service manages VCS repositories
	Service interface {
		// Connect adds a connection between a VCS repo and an OTF resource. A
		// webhook is created if one doesn't exist already.
		Connect(ctx context.Context, opts ConnectOptions) (*Connection, error)
		// Disconnect removes a connection between a VCS repo and an OTF
		// resource. If there are no more connections then its
		// webhook is removed.
		Disconnect(ctx context.Context, opts DisconnectOptions) error

		deleteUnreferencedWebhooks(ctx context.Context) error
	}

	service struct {
		logr.Logger
		vcsprovider.Service

		*db

		*handler      // handles incoming vcs events
		factory       // produce new hooks
		*synchroniser // synchronise hooks
	}

	Options struct {
		logr.Logger

		CloudService cloud.Service

		*sql.DB
		*pubsub.Broker
		internal.HostnameService
		VCSProviderService vcsprovider.Service
		organization.OrganizationService
	}
)

func NewService(ctx context.Context, opts Options) *service {
	factory := factory{
		HostnameService: opts.HostnameService,
		Service:         opts.CloudService,
	}
	db := &db{opts.DB, factory}
	handler := &handler{
		Logger:    opts.Logger,
		Publisher: opts.Broker,
		handlerDB: db,
	}
	svc := &service{
		Logger:       opts.Logger,
		Service:      opts.VCSProviderService,
		db:           db,
		factory:      factory,
		handler:      handler,
		synchroniser: &synchroniser{Logger: opts.Logger, syncdb: db},
	}

	// Delete webhooks prior to the deletion of VCS providers. VCS providers are
	// necessary for the deletion of webhooks from VCS repos. Hence we need to
	// first delete webhooks that reference the VCS provider before the VCS
	// provider is deleted.
	opts.VCSProviderService.BeforeDeleteHook(svc.deleteProviderWebhooks)
	// Delete webhooks prior to the deletion of organizations. Deleting
	// organizations cascades deletion of VCS providers (see above).
	opts.OrganizationService.BeforeDeleteHook(svc.deleteOrganizationWebhooks)

	// Register with broker - when a repo connection is deleted in postgres, a
	// postgres trigger sends a message to the broker, which calls this function
	// to convert the message into an event. The purger subsystem then uses this
	// event to delete the corresponding webhook if it is no longer in use.
	deleteEventFunc := func(ctx context.Context, rawWebhookID string, action pubsub.DBAction) (any, error) {
		webhookID, err := uuid.Parse(rawWebhookID)
		if err != nil {
			return nil, err
		}
		if action != pubsub.DeleteDBAction {
			return nil, fmt.Errorf("trigger not registered for action: %s", action)
		}
		return &repoConnectionEvent{webhookID: webhookID}, nil
	}
	opts.Register("repo_connections", pubsub.GetterFunc(deleteEventFunc))

	return svc
}

// Connect an OTF resource to a VCS repo.
func (s *service) Connect(ctx context.Context, opts ConnectOptions) (*Connection, error) {
	vcsProvider, err := s.GetVCSProvider(ctx, opts.VCSProviderID)
	if err != nil {
		return nil, fmt.Errorf("retrieving vcs provider: %w", err)
	}
	client, err := s.GetVCSClient(ctx, opts.VCSProviderID)
	if err != nil {
		return nil, fmt.Errorf("retrieving vcs client: %w", err)
	}
	_, err = client.GetRepository(ctx, opts.RepoPath)
	if err != nil {
		return nil, fmt.Errorf("checking repository exists: %w", err)
	}

	hook, err := s.newHook(newHookOptions{
		identifier:    opts.RepoPath,
		cloud:         vcsProvider.CloudConfig.Name,
		vcsProviderID: vcsProvider.ID,
	})
	if err != nil {
		return nil, fmt.Errorf("constructing webhook: %w", err)
	}

	// lock webhooks table to prevent concurrent updates (a row-level lock is
	// insufficient)
	err = s.db.Lock(ctx, "webhooks", func(ctx context.Context, q pggen.Querier) error {
		hook, err = s.db.getOrCreateHook(ctx, hook)
		if err != nil {
			return fmt.Errorf("getting or creating webhook: %w", err)
		}
		if err := s.sync(ctx, client, hook); err != nil {
			return fmt.Errorf("synchronising webhook: %w", err)
		}
		return s.db.createConnection(ctx, hook.id, opts)
	})
	if err != nil {
		return nil, err
	}
	return &Connection{
		Repo:          opts.RepoPath,
		VCSProviderID: opts.VCSProviderID,
	}, nil
}

// Disconnect resource from repo
func (s *service) Disconnect(ctx context.Context, opts DisconnectOptions) error {
	return s.db.deleteConnection(ctx, opts)
}

func (s *service) deleteOrganizationWebhooks(ctx context.Context, org string) error {
	providers, err := s.ListVCSProviders(ctx, org)
	if err != nil {
		return err
	}
	hooks, err := s.db.listHooks(ctx)
	if err != nil {
		return err
	}
	for _, p := range providers {
		for _, h := range hooks {
			if h.vcsProviderID == p.ID {
				if err := s.deleteWebhook(ctx, h); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (s *service) deleteProviderWebhooks(ctx context.Context, providerID string) error {
	hooks, err := s.db.listHooks(ctx)
	if err != nil {
		return err
	}
	for _, h := range hooks {
		if h.vcsProviderID == providerID {
			if err := s.deleteWebhook(ctx, h); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *service) deleteUnreferencedWebhooks(ctx context.Context) error {
	hooks, err := s.db.listUnreferencedWebhooks(ctx)
	if err != nil {
		return fmt.Errorf("listing unreferenced webhooks: %w", err)
	}
	for _, h := range hooks {
		if err := s.deleteWebhook(ctx, h); err != nil {
			return err
		}
	}
	return nil
}

func (s *service) deleteWebhook(ctx context.Context, webhook *hook) error {
	if err := s.db.deleteHook(ctx, webhook.id); err != nil {
		return fmt.Errorf("deleting webhook from db: %w", err)
	}
	client, err := s.GetVCSClient(ctx, webhook.vcsProviderID)
	if err != nil {
		return fmt.Errorf("retrieving vcs client from db: %w", err)
	}
	err = client.DeleteWebhook(ctx, cloud.DeleteWebhookOptions{
		Repo: webhook.identifier,
		ID:   *webhook.cloudID,
	})
	if err != nil {
		s.Error(err, "deleting webhook", "repo", webhook.identifier, "cloud", webhook.cloud)
	} else {
		s.V(0).Info("deleted webhook", "repo", webhook.identifier, "cloud", webhook.cloud)
	}
	// Failure to delete the webhook from the cloud provider is not deemed a
	// fatal error.
	return nil
}
