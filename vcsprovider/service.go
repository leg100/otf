package vcsprovider

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/gorilla/mux"
	"github.com/leg100/otf"
	"github.com/leg100/otf/cloud"
	"github.com/leg100/otf/organization"
	"github.com/leg100/otf/rbac"
)

type (
	// Alias services so they don't conflict when nested together in struct
	VCSProviderService Service
	CloudService       cloud.Service

	Service interface {
		CreateVCSProvider(ctx context.Context, opts CreateOptions) (*VCSProvider, error)
		GetVCSProvider(ctx context.Context, id string) (*VCSProvider, error)
		ListVCSProviders(ctx context.Context, organization string) ([]*VCSProvider, error)
		DeleteVCSProvider(ctx context.Context, id string) (*VCSProvider, error)

		// GetVCSClient combines retrieving a vcs provider and construct a cloud
		// client from that provider.
		//
		// TODO: rename vcs provider to cloud client; the central purpose of the vcs
		// provider is, after all, to construct a cloud client.
		GetVCSClient(ctx context.Context, providerID string) (cloud.Client, error)
	}

	service struct {
		logr.Logger

		organization otf.Authorizer
		db           *pgdb

		*factory
		web *webHandlers
	}

	Options struct {
		CloudService
		otf.DB
		otf.Renderer
		logr.Logger
	}
)

func NewService(opts Options) *service {
	svc := service{
		Logger:       opts.Logger,
		organization: &organization.Authorizer{Logger: opts.Logger},
		db:           newDB(opts.DB, opts.CloudService),
		factory: &factory{
			CloudService: opts.CloudService,
		},
	}

	svc.web = &webHandlers{
		CloudService: opts.CloudService,
		Renderer:     opts.Renderer,
		svc:          &svc,
	}
	return &svc
}

func (a *service) AddHandlers(r *mux.Router) {
	a.web.addHandlers(r)
}

func (a *service) CreateVCSProvider(ctx context.Context, opts CreateOptions) (*VCSProvider, error) {
	subject, err := a.organization.CanAccess(ctx, rbac.CreateVCSProviderAction, opts.Organization)
	if err != nil {
		return nil, err
	}

	provider, err := a.new(opts)
	if err != nil {
		return nil, err
	}

	if err := a.db.create(ctx, provider); err != nil {
		a.Error(err, "creating vcs provider", "provider", provider, "subject", subject)
		return nil, err
	}
	a.V(0).Info("created vcs provider", "provider", provider, "subject", subject)
	return provider, nil
}

func (a *service) ListVCSProviders(ctx context.Context, organization string) ([]*VCSProvider, error) {
	subject, err := a.organization.CanAccess(ctx, rbac.ListVCSProvidersAction, organization)
	if err != nil {
		return nil, err
	}

	providers, err := a.db.list(ctx, organization)
	if err != nil {
		a.Error(err, "listing vcs providers", "organization", organization, "subject", subject)
		return nil, err
	}
	a.V(2).Info("listed vcs providers", "organization", organization, "subject", subject)
	return providers, nil
}

func (a *service) GetVCSProvider(ctx context.Context, id string) (*VCSProvider, error) {
	// Parameters only include VCS Provider ID, so we can only determine
	// authorization _after_ retrieving the provider
	provider, err := a.db.get(ctx, id)
	if err != nil {
		a.Error(err, "retrieving vcs provider", "id", id)
		return nil, err
	}

	subject, err := a.organization.CanAccess(ctx, rbac.GetVCSProviderAction, provider.Organization)
	if err != nil {
		return nil, err
	}
	a.V(2).Info("retrieved vcs provider", "provider", provider, "subject", subject)

	return provider, nil
}

func (a *service) GetVCSClient(ctx context.Context, providerID string) (cloud.Client, error) {
	provider, err := a.GetVCSProvider(ctx, providerID)
	if err != nil {
		return nil, err
	}
	return provider.NewClient(ctx)
}

func (a *service) DeleteVCSProvider(ctx context.Context, id string) (*VCSProvider, error) {
	// retrieve vcs provider first in order to get organization for authorization
	provider, err := a.db.get(ctx, id)
	if err != nil {
		a.Error(err, "retrieving vcs provider", "id", id)
		return nil, err
	}

	subject, err := a.organization.CanAccess(ctx, rbac.DeleteVCSProviderAction, provider.Organization)
	if err != nil {
		return nil, err
	}

	if err := a.db.delete(ctx, id); err != nil {
		a.Error(err, "deleting vcs provider", "provider", provider, "subject", subject)
		return nil, err
	}
	a.V(0).Info("deleted vcs provider", "provider", provider, "subject", subject)
	return provider, nil
}
