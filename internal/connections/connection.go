// Package connections manages connections between VCS repositories and OTF
// resources, e.g. workspaces, modules.
package connections

import (
	"context"
	"fmt"

	"github.com/leg100/otf/internal/logr"
	"github.com/leg100/otf/internal/repo"
	"github.com/leg100/otf/internal/sql"
	"github.com/leg100/otf/internal/sql/pggen"
	"github.com/leg100/otf/internal/vcsprovider"
)

const (
	WorkspaceConnection ConnectionType = iota
	ModuleConnection
)

type (
	// ConnectionType identifies the OTF resource type in a VCS connection.
	ConnectionType int

	// Connection is a connection between a VCS repo and an OTF resource.
	Connection struct {
		VCSProviderID string
		Repo          string
	}

	ConnectOptions struct {
		ConnectionType // OTF resource type

		VCSProviderID string // vcs provider of repo
		ResourceID    string // ID of OTF resource
		RepoPath      string
	}

	DisconnectOptions struct {
		ConnectionType // OTF resource type

		ResourceID string // ID of OTF resource
	}

	SynchroniseOptions struct {
		VCSProviderID string // vcs provider of repo
		RepoPath      string
	}

	ConnectionService Service

	// Service manages connections between OTF resources and VCS repos
	Service interface {
		// Connect adds a connection between a VCS repo and an OTF resource. A
		// webhook is created if one doesn't exist already.
		Connect(ctx context.Context, opts ConnectOptions) (*Connection, error)

		// Disconnect removes a connection between a VCS repo and an OTF
		// resource. If there are no more connections then its
		// webhook is removed.
		Disconnect(ctx context.Context, opts DisconnectOptions) error
	}

	Options struct {
		logr.Logger
		vcsprovider.VCSProviderService
		repo.RepoService
		*sql.DB
	}

	service struct {
		logr.Logger
		vcsprovider.Service
		repo.RepoService

		*db
	}
)

func NewService(ctx context.Context, opts Options) *service {
	return &service{
		Logger:      opts.Logger,
		Service:     opts.VCSProviderService,
		RepoService: opts.RepoService,
		db:          &db{opts.DB},
	}
}

// Connect an OTF resource to a VCS repo.
func (s *service) Connect(ctx context.Context, opts ConnectOptions) (*Connection, error) {
	// check vcs provider is valid
	provider, err := s.GetVCSProvider(ctx, opts.VCSProviderID)
	if err != nil {
		return nil, fmt.Errorf("retrieving vcs provider: %w", err)
	}

	err = s.db.Tx(ctx, func(ctx context.Context, q pggen.Querier) error {
		// github app vcs provider does not require a repohook to be created
		if provider.GithubApp == nil {
			_, err := s.RepoService.CreateWebhook(ctx, repo.CreateWebhookOptions{
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
func (s *service) Disconnect(ctx context.Context, opts DisconnectOptions) error {
	return s.db.Tx(ctx, func(ctx context.Context, q pggen.Querier) error {
		if err := s.db.deleteConnection(ctx, opts); err != nil {
			return err
		}
		// now that a connection has been deleted, also delete any webhooks that
		// are no longer referenced by connections
		if err := s.RepoService.DeleteUnreferencedWebhooks(ctx); err != nil {
			return err
		}
		return nil
	})
}
