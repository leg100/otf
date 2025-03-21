// Package connections manages connections between VCS repositories and OTF
// resources, e.g. workspaces, modules.
package connections

import (
	"context"
	"fmt"

	"github.com/leg100/otf/internal/logr"
	"github.com/leg100/otf/internal/repohooks"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/sql"
	"github.com/leg100/otf/internal/vcsprovider"
)

type (
	// Connection is a connection between a VCS repo and an OTF resource.
	Connection struct {
		VCSProviderID resource.TfeID
		Repo          string
	}

	ConnectOptions struct {
		VCSProviderID resource.TfeID // vcs provider of repo
		ResourceID    resource.TfeID // ID of OTF resource to connect.
		RepoPath      string
	}

	DisconnectOptions struct {
		ResourceID resource.TfeID // ID of OTF resource to disconnect
	}

	SynchroniseOptions struct {
		VCSProviderID resource.TfeID // vcs provider of repo
		RepoPath      string
	}

	Options struct {
		logr.Logger
		*sql.DB

		VCSProviderService *vcsprovider.Service
		RepoHooksService   *repohooks.Service
	}

	Service struct {
		logr.Logger

		*db

		repohooks    *repohooks.Service
		vcsproviders *vcsprovider.Service
	}
)

func NewService(ctx context.Context, opts Options) *Service {
	return &Service{
		Logger:       opts.Logger,
		vcsproviders: opts.VCSProviderService,
		repohooks:    opts.RepoHooksService,
		db:           &db{opts.DB},
	}
}

// Connect an OTF resource to a VCS repo.
func (s *Service) Connect(ctx context.Context, opts ConnectOptions) (*Connection, error) {
	// check vcs provider is valid
	provider, err := s.vcsproviders.Get(ctx, opts.VCSProviderID)
	if err != nil {
		return nil, fmt.Errorf("retrieving vcs provider: %w", err)
	}

	err = s.db.Tx(ctx, func(ctx context.Context, _ sql.Connection) error {
		// github app vcs provider does not require a repohook to be created
		if provider.GithubApp == nil {
			_, err := s.repohooks.CreateRepohook(ctx, repohooks.CreateRepohookOptions{
				VCSProviderID: opts.VCSProviderID,
				RepoPath:      opts.RepoPath,
			})
			if err != nil {
				return fmt.Errorf("creating webhook: %w", err)
			}
		}
		return s.db.createConnection(ctx, opts)
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
func (s *Service) Disconnect(ctx context.Context, opts DisconnectOptions) error {
	return s.db.Tx(ctx, func(ctx context.Context, _ sql.Connection) error {
		if err := s.db.deleteConnection(ctx, opts); err != nil {
			return err
		}
		// now that a connection has been deleted, also delete any repohooks that
		// are no longer referenced by connections
		if err := s.repohooks.DeleteUnreferencedRepohooks(ctx); err != nil {
			return err
		}
		return nil
	})
}
