package vcsprovider

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/cloud"
	"github.com/leg100/otf/internal/hooks"
	"github.com/leg100/otf/internal/http/html"
	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/rbac"
	"github.com/leg100/otf/internal/sql"
	"github.com/leg100/otf/internal/tfeapi"
)

type (
	// Alias services so they don't conflict when nested together in struct
	VCSProviderService Service
	CloudService       cloud.Service

	Service interface {
		CreateVCSProvider(ctx context.Context, opts CreateOptions) (*VCSProvider, error)
		UpdateVCSProvider(ctx context.Context, id string, opts UpdateOptions) (*VCSProvider, error)
		GetVCSProvider(ctx context.Context, id string) (*VCSProvider, error)
		ListVCSProviders(ctx context.Context, organization string) ([]*VCSProvider, error)
		ListAllVCSProviders(ctx context.Context) ([]*VCSProvider, error)
		DeleteVCSProvider(ctx context.Context, id string) (*VCSProvider, error)

		// GetVCSClient combines retrieving a vcs provider and construct a cloud
		// client from that provider.
		//
		// TODO: rename vcs provider to cloud client; the central purpose of the vcs
		// provider is, after all, to construct a cloud client.
		GetVCSClient(ctx context.Context, providerID string) (cloud.Client, error)

		BeforeDeleteVCSProvider(l hooks.Listener[*VCSProvider])
	}

	service struct {
		logr.Logger
		CloudService

		site         internal.Authorizer
		organization internal.Authorizer
		db           *pgdb
		web          *webHandlers
		api          *tfe
		deleteHook   *hooks.Hook[*VCSProvider]
	}

	Options struct {
		CloudService
		internal.HostnameService
		*sql.DB
		*tfeapi.Responder
		html.Renderer
		logr.Logger
	}
)

func NewService(opts Options) *service {
	svc := service{
		Logger:       opts.Logger,
		site:         &internal.SiteAuthorizer{Logger: opts.Logger},
		organization: &organization.Authorizer{Logger: opts.Logger},
		db:           newDB(opts.DB, opts.CloudService),
		CloudService: opts.CloudService,
		deleteHook:   hooks.NewHook[*VCSProvider](opts.DB),
	}

	svc.web = &webHandlers{
		CloudService:    opts.CloudService,
		Renderer:        opts.Renderer,
		HostnameService: opts.HostnameService,
		svc:             &svc,
	}
	svc.api = &tfe{
		Service:   &svc,
		Responder: opts.Responder,
	}

	return &svc
}

func (a *service) AddHandlers(r *mux.Router) {
	a.web.addHandlers(r)
	a.api.addHandlers(r)
}

func (a *service) BeforeDeleteVCSProvider(l hooks.Listener[*VCSProvider]) {
	a.deleteHook.Before(l)
}

func (a *service) CreateVCSProvider(ctx context.Context, opts CreateOptions) (*VCSProvider, error) {
	subject, err := a.organization.CanAccess(ctx, rbac.CreateVCSProviderAction, opts.Organization)
	if err != nil {
		return nil, err
	}

	provider, err := newProvider(ctx, opts)
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

func (a *service) UpdateVCSProvider(ctx context.Context, id string, opts UpdateOptions) (*VCSProvider, error) {
	var (
		subject internal.Subject
		before  VCSProvider
		after   *VCSProvider
	)
	err := a.db.update(ctx, id, func(provider *VCSProvider) (err error) {
		subject, err = a.organization.CanAccess(ctx, rbac.UpdateVariableSetAction, provider.Organization)
		if err != nil {
			return err
		}
		// keep copy for logging the differences before and after update
		before = *provider
		after = provider
		if err := after.Update(opts); err != nil {
			return err
		}
		return err
	})
	if err != nil {
		a.Error(err, "updating vcs provider", "vcs_provider_id", id)
		return nil, err
	}
	a.V(0).Info("updated vcs provider", "before", &before, "after", after, "subject", subject)
	return after, nil
}

func (a *service) ListVCSProviders(ctx context.Context, organization string) ([]*VCSProvider, error) {
	subject, err := a.organization.CanAccess(ctx, rbac.ListVCSProvidersAction, organization)
	if err != nil {
		return nil, err
	}

	providers, err := a.db.listByOrganization(ctx, organization)
	if err != nil {
		a.Error(err, "listing vcs providers", "organization", organization, "subject", subject)
		return nil, err
	}
	a.V(9).Info("listed vcs providers", "organization", organization, "subject", subject)
	return providers, nil
}

func (a *service) ListAllVCSProviders(ctx context.Context) ([]*VCSProvider, error) {
	subject, err := a.site.CanAccess(ctx, rbac.ListVCSProvidersAction, "")
	if err != nil {
		return nil, err
	}

	providers, err := a.db.list(ctx)
	if err != nil {
		a.Error(err, "listing vcs providers", "subject", subject)
		return nil, err
	}
	a.V(9).Info("listed vcs providers", "subject", subject)
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
	a.V(9).Info("retrieved vcs provider", "provider", provider, "subject", subject)

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

	err = a.deleteHook.Dispatch(ctx, provider, func(ctx context.Context) error {
		return a.db.delete(ctx, id)
	})
	if err != nil {
		a.Error(err, "deleting vcs provider", "provider", provider, "subject", subject)
		return nil, err
	}
	a.V(0).Info("deleted vcs provider", "provider", provider, "subject", subject)
	return provider, nil
}
