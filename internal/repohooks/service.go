package repohooks

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/google/uuid"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/github"
	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/sql"
	"github.com/leg100/otf/internal/sql/sqlc"
	"github.com/leg100/otf/internal/vcs"
	"github.com/leg100/otf/internal/vcsprovider"
)

type (
	Service struct {
		logr.Logger

		*db
		*handlers     // handles incoming vcs events
		*synchroniser // synchronise hooks

		vcsproviders *vcsprovider.Service
	}

	Options struct {
		logr.Logger

		OrganizationService *organization.Service
		VCSProviderService  *vcsprovider.Service
		GithubAppService    *github.Service
		VCSEventBroker      *vcs.Broker

		*sql.DB
		*internal.HostnameService
	}

	CreateRepohookOptions struct {
		VCSProviderID string // vcs provider of repo
		RepoPath      string
	}
)

func NewService(ctx context.Context, opts Options) *Service {
	db := &db{opts.DB, opts.HostnameService}
	svc := &Service{
		Logger:       opts.Logger,
		vcsproviders: opts.VCSProviderService,
		db:           db,
		handlers: newHandler(
			opts.Logger,
			opts.VCSEventBroker,
			db,
		),
		synchroniser: &synchroniser{Logger: opts.Logger, syncdb: db},
	}
	// Delete webhooks prior to the deletion of VCS providers. VCS providers are
	// necessary for the deletion of webhooks from VCS repos. Hence we need to
	// first delete webhooks that reference the VCS provider before the VCS
	// provider is deleted.
	opts.VCSProviderService.BeforeDeleteVCSProvider(svc.deleteProviderRepohooks)
	// Delete webhooks prior to the deletion of organizations. Deleting
	// organizations cascades deletion of VCS providers (see above).
	opts.OrganizationService.BeforeDeleteOrganization(svc.deleteOrganizationRepohooks)
	return svc
}

func (s *Service) CreateRepohook(ctx context.Context, opts CreateRepohookOptions) (uuid.UUID, error) {
	vcsProvider, err := s.vcsproviders.Get(ctx, opts.VCSProviderID)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("retrieving vcs provider: %w", err)
	}
	if vcsProvider.GithubApp != nil {
		// github apps don't need a webhook created on each repo.
		return uuid.UUID{}, nil
	}
	client, err := s.vcsproviders.GetVCSClient(ctx, opts.VCSProviderID)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("retrieving vcs client: %w", err)
	}
	_, err = client.GetRepository(ctx, opts.RepoPath)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("checking repository exists: %w", err)
	}
	hook, err := newRepohook(newRepohookOptions{
		repoPath:        opts.RepoPath,
		cloud:           vcsProvider.Kind,
		vcsProviderID:   vcsProvider.ID,
		HostnameService: s.HostnameService,
	})
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("constructing webhook: %w", err)
	}
	// lock repohooks table to prevent concurrent updates (a row-level lock is
	// insufficient)
	err = s.db.Lock(ctx, "repohooks", func(ctx context.Context, q *sqlc.Queries) error {
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

func (s *Service) RegisterCloudHandler(kind vcs.Kind, h EventUnmarshaler) {
	s.handlers.cloudHandlers.Set(kind, h)
}

func (s *Service) DeleteUnreferencedRepohooks(ctx context.Context) error {
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

func (s *Service) deleteOrganizationRepohooks(ctx context.Context, org *organization.Organization) error {
	providers, err := s.vcsproviders.List(ctx, org.Name)
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

func (s *Service) deleteProviderRepohooks(ctx context.Context, provider *vcsprovider.VCSProvider) error {
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

func (s *Service) deleteRepohook(ctx context.Context, repohook *hook) error {
	if err := s.db.deleteHook(ctx, repohook.id); err != nil {
		return fmt.Errorf("deleting webhook from db: %w", err)
	}
	client, err := s.vcsproviders.GetVCSClient(ctx, repohook.vcsProviderID)
	if err != nil {
		return fmt.Errorf("retrieving vcs client from db: %w", err)
	}
	err = client.DeleteWebhook(ctx, vcs.DeleteWebhookOptions{
		Repo: repohook.repoPath,
		ID:   *repohook.cloudID,
	})
	if err != nil {
		s.Error(err, "deleting webhook", "repo", repohook.repoPath, "cloud", repohook.cloud)
	} else {
		s.V(0).Info("deleted webhook", "repo", repohook.repoPath, "cloud", repohook.cloud)
	}
	// Failure to delete the webhook from the cloud provider is not deemed a
	// fatal error.
	return nil
}
