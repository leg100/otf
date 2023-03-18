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
	VCSProviderService Service

	Service interface {
		// GetVCSClient combines retrieving a vcs provider and construct a cloud
		// client from that provider.
		//
		// TODO: rename vcs provider to cloud client; the central purpose of the vcs
		// provider is, after all, to construct a cloud client.
		GetVCSClient(ctx context.Context, providerID string) (cloud.Client, error)
		ListVCSProviders(ctx context.Context, organization string) ([]*VCSProvider, error)
		GetVCSProvider(ctx context.Context, id string) (*VCSProvider, error)
		create(ctx context.Context, opts createOptions) (*VCSProvider, error)
		get(ctx context.Context, id string) (*VCSProvider, error)
		list(ctx context.Context, organization string) ([]*VCSProvider, error)
		delete(ctx context.Context, id string) (*VCSProvider, error)
	}

	service struct {
		logr.Logger

		organization otf.Authorizer
		db           *pgdb

		*factory
		web *webHandlers
	}

	Options struct {
		cloud.Service
		otf.DB
		otf.Renderer
		logr.Logger
	}
)

func NewService(opts Options) *service {
	svc := service{
		Logger:       opts.Logger,
		organization: &organization.Authorizer{opts.Logger},
		db:           newDB(opts.DB, opts.Service),
		factory: &factory{
			Service: opts.Service,
		},
	}

	svc.web = &webHandlers{
		Renderer: opts.Renderer,
		svc:      &svc,
	}
	return &svc
}

func (a *service) AddHandlers(r *mux.Router) {
	a.web.addHandlers(r)
}

func (a *service) ListVCSProviders(ctx context.Context, organization string) ([]*VCSProvider, error) {
	return a.list(ctx, organization)
}

func (a *service) GetVCSProvider(ctx context.Context, id string) (*VCSProvider, error) {
	return a.get(ctx, id)
}

func (a *service) GetVCSClient(ctx context.Context, providerID string) (cloud.Client, error) {
	provider, err := a.get(ctx, providerID)
	if err != nil {
		return nil, err
	}
	return provider.NewClient(ctx)
}

func (a *service) create(ctx context.Context, opts createOptions) (*VCSProvider, error) {
	subject, err := a.organization.CanAccess(ctx, rbac.CreateVCSProviderAction, opts.Organization)
	if err != nil {
		return nil, err
	}

	provider, err := a.new(opts)
	if err != nil {
		return nil, err
	}

	if err := a.db.create(ctx, provider); err != nil {
		a.Error(err, "creating vcs provider", "organization", opts.Organization, "id", provider.ID, "subject", subject)
		return nil, err
	}
	a.V(0).Info("created vcs provider", "organization", opts.Organization, "id", provider.ID, "subject", subject)
	return provider, nil
}

func (a *service) get(ctx context.Context, id string) (*VCSProvider, error) {
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

func (a *service) list(ctx context.Context, organization string) ([]*VCSProvider, error) {
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

func (a *service) delete(ctx context.Context, id string) (*VCSProvider, error) {
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
