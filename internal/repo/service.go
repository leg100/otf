package repo

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/google/uuid"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/github"
	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/sql"
	"github.com/leg100/otf/internal/sql/pggen"
	"github.com/leg100/otf/internal/vcs"
	"github.com/leg100/otf/internal/vcsprovider"
)

type (
	RepoService = Service

	// Service manages webhooks
	Service interface {
		// CreateWebhook creates a webhook on a VCS repository. If the webhook
		// already exists, it is updated if there are discrepancies; otherwise
		// no action is taken. In any case an identifier is returned uniquely
		// identifying the webhook.
		CreateWebhook(ctx context.Context, opts CreateWebhookOptions) (uuid.UUID, error)
		// RegisterCloudHandler registers a new cloud handler, to handle VCS
		// events for a specific vcs hosting provider.
		RegisterCloudHandler(kind vcs.Kind, h CloudHandler)
		// DeleteUnreferencedWebhooks deletes any repohooks no longer used
		// by a VCS connection
		DeleteUnreferencedWebhooks(ctx context.Context) error
	}

	service struct {
		logr.Logger
		vcsprovider.Service

		*db

		*handlers     // handles incoming vcs events
		*synchroniser // synchronise hooks
	}

	Options struct {
		logr.Logger

		*sql.DB
		VCSEventBroker *vcs.Broker
		internal.HostnameService
		VCSProviderService vcsprovider.Service
		organization.OrganizationService
		github.GithubAppService
	}

	// CreateOptions are options for creating a webhook.
	CreateOptions struct {
		Client        vcs.Client
		VCSProviderID string
		RepoPath      string
	}

	CreateWebhookOptions struct {
		VCSProviderID string // vcs provider of repo
		RepoPath      string
	}
)

func NewService(ctx context.Context, opts Options) *service {
	db := &db{opts.DB, opts.HostnameService}
	svc := &service{
		Logger:  opts.Logger,
		Service: opts.VCSProviderService,
		db:      db,
		handlers: newHandler(
			opts.Logger,
			opts.VCSEventBroker,
			opts.VCSProviderService,
			db,
			opts.GithubAppService,
		),
		synchroniser: &synchroniser{Logger: opts.Logger, syncdb: db},
	}

	// Delete webhooks prior to the deletion of VCS providers. VCS providers are
	// necessary for the deletion of webhooks from VCS repos. Hence we need to
	// first delete webhooks that reference the VCS provider before the VCS
	// provider is deleted.
	opts.VCSProviderService.BeforeDeleteVCSProvider(svc.deleteProviderWebhooks)
	// Delete webhooks prior to the deletion of organizations. Deleting
	// organizations cascades deletion of VCS providers (see above).
	opts.OrganizationService.BeforeDeleteOrganization(svc.deleteOrganizationWebhooks)

	return svc
}

func (s *service) CreateWebhook(ctx context.Context, opts CreateWebhookOptions) (uuid.UUID, error) {
	vcsProvider, err := s.GetVCSProvider(ctx, opts.VCSProviderID)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("retrieving vcs provider: %w", err)
	}
	if vcsProvider.GithubApp != nil {
		// github apps don't need a webhook created on each repo.
		return uuid.UUID{}, nil
	}
	client, err := s.GetVCSClient(ctx, opts.VCSProviderID)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("retrieving vcs client: %w", err)
	}
	_, err = client.GetRepository(ctx, opts.RepoPath)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("checking repository exists: %w", err)
	}
	hook, err := newHook(newHookOptions{
		identifier:      opts.RepoPath,
		cloud:           vcsProvider.Kind,
		vcsProviderID:   vcsProvider.ID,
		HostnameService: s.HostnameService,
	})
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("constructing webhook: %w", err)
	}
	// lock repohooks table to prevent concurrent updates (a row-level lock is
	// insufficient)
	err = s.db.Lock(ctx, "repohooks", func(ctx context.Context, q pggen.Querier) error {
		hook, err = s.db.getOrCreateHook(ctx, hook)
		if err != nil {
			return fmt.Errorf("getting or creating webhook: %w", err)
		}
		if err := s.sync(ctx, client, hook); err != nil {
			return fmt.Errorf("synchronising webhook: %w", err)
		}
		return nil
	})
	if err != nil {
		return uuid.UUID{}, err
	}
	return hook.id, nil
}

func (s *service) RegisterCloudHandler(kind vcs.Kind, h CloudHandler) {
	s.handlers.cloudHandlers.Set(kind, h)
}

func (s *service) DeleteUnreferencedWebhooks(ctx context.Context) error {
	hooks, err := s.db.listUnreferencedRepohooks(ctx)
	if err != nil {
		return fmt.Errorf("listing unreferenced webhooks: %w", err)
	}
	for _, h := range hooks {
		if err := s.deleteRepohook(ctx, h); err != nil {
			return err
		}
	}
	return nil
}

func (s *service) deleteOrganizationWebhooks(ctx context.Context, org *organization.Organization) error {
	providers, err := s.ListVCSProviders(ctx, org.Name)
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
				if err := s.deleteRepohook(ctx, h); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (s *service) deleteProviderWebhooks(ctx context.Context, provider *vcsprovider.VCSProvider) error {
	hooks, err := s.db.listHooks(ctx)
	if err != nil {
		return err
	}
	for _, h := range hooks {
		if h.vcsProviderID == provider.ID {
			if err := s.deleteRepohook(ctx, h); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *service) deleteRepohook(ctx context.Context, repohook *hook) error {
	if err := s.db.deleteHook(ctx, repohook.id); err != nil {
		return fmt.Errorf("deleting webhook from db: %w", err)
	}
	client, err := s.GetVCSClient(ctx, repohook.vcsProviderID)
	if err != nil {
		return fmt.Errorf("retrieving vcs client from db: %w", err)
	}
	err = client.DeleteWebhook(ctx, vcs.DeleteWebhookOptions{
		Repo: repohook.identifier,
		ID:   *repohook.cloudID,
	})
	if err != nil {
		s.Error(err, "deleting webhook", "repo", repohook.identifier, "cloud", repohook.cloud)
	} else {
		s.V(0).Info("deleted webhook", "repo", repohook.identifier, "cloud", repohook.cloud)
	}
	// Failure to delete the webhook from the cloud provider is not deemed a
	// fatal error.
	return nil
}
