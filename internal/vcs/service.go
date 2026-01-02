package vcs

import (
	"context"

	"github.com/leg100/otf/internal/logr"
	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/authz"
	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/sql"
	"github.com/leg100/otf/internal/tfeapi"
)

type (
	Service struct {
		logr.Logger
		*authz.Authorizer

		db                *pgdb
		api               *tfe
		beforeDeleteHooks []func(context.Context, *Provider) error

		*internal.HostnameService
		*factory
		*kindDB
	}

	Options struct {
		*internal.HostnameService
		*sql.DB
		*tfeapi.Responder
		logr.Logger

		SourceIconRegistrar SourceIconRegistrar
		SkipTLSVerification bool
		Authorizer          *authz.Authorizer
	}
)

func NewService(opts Options) *Service {
	kindDB := newKindDB(opts.SourceIconRegistrar)
	factory := factory{kinds: kindDB}
	svc := Service{
		Logger:          opts.Logger,
		HostnameService: opts.HostnameService,
		Authorizer:      opts.Authorizer,
		factory:         &factory,
		db: &pgdb{
			DB:    opts.DB,
			kinds: kindDB,
		},
		kindDB: kindDB,
	}
	svc.api = &tfe{
		Service:   &svc,
		Responder: opts.Responder,
	}
	return &svc
}

func (a *Service) AddHandlers(r *mux.Router) {
	a.api.addHandlers(r)
}

func (a *Service) Create(ctx context.Context, opts CreateOptions) (*Provider, error) {
	subject, err := a.Authorize(ctx, authz.CreateVCSProviderAction, &opts.Organization)
	if err != nil {
		return nil, err
	}

	provider, err := a.newProvider(ctx, opts)
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

func (a *Service) Update(ctx context.Context, id resource.TfeID, opts UpdateOptions) (*Provider, error) {
	var (
		subject authz.Subject
		before  Provider
		after   *Provider
	)
	err := a.db.update(ctx, id, func(ctx context.Context, provider *Provider) (err error) {
		subject, err = a.Authorize(ctx, authz.UpdateVariableSetAction, &provider.Organization)
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

func (a *Service) List(ctx context.Context, organization organization.Name) ([]*Provider, error) {
	subject, err := a.Authorize(ctx, authz.ListVCSProvidersAction, organization)
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

func (a *Service) ListByInstall(ctx context.Context, installID int64) ([]*Provider, error) {
	subject, err := authz.SubjectFromContext(ctx)
	if err != nil {
		return nil, err
	}
	providers, err := a.db.listByInstall(ctx, installID)
	if err != nil {
		a.Error(err, "listing vcs providers by install", "subject", subject, "install_id", installID)
		return nil, err
	}
	a.V(9).Info("listed vcs providers by install", "count", len(providers), "subject", subject, "install_id", installID)
	return providers, nil
}

func (a *Service) Get(ctx context.Context, id resource.TfeID) (*Provider, error) {
	// Parameters only include VCS Provider ID, so we can only determine
	// authorization _after_ retrieving the provider
	provider, err := a.db.get(ctx, id)
	if err != nil {
		a.Error(err, "retrieving vcs provider", "id", id)
		return nil, err
	}

	subject, err := a.Authorize(ctx, authz.GetVCSProviderAction, &provider.Organization)
	if err != nil {
		return nil, err
	}
	a.V(9).Info("retrieved vcs provider", "provider", provider, "subject", subject)

	return provider, nil
}

func (a *Service) Delete(ctx context.Context, id resource.TfeID) (*Provider, error) {
	var (
		provider *Provider
		subject  authz.Subject
	)
	err := a.db.Tx(ctx, func(ctx context.Context) (err error) {
		// retrieve vcs provider first in order to get organization for authorization
		provider, err = a.db.get(ctx, id)
		if err != nil {
			a.Error(err, "retrieving vcs provider", "id", id)
			return err
		}

		subject, err = a.Authorize(ctx, authz.DeleteVCSProviderAction, &provider.Organization)
		if err != nil {
			return err
		}

		for _, hook := range a.beforeDeleteHooks {
			if err := hook(ctx, provider); err != nil {
				return err
			}
		}
		return a.db.delete(ctx, id)
	})
	if err != nil {
		a.Error(err, "deleting vcs provider", "provider", provider, "subject", subject)
		return nil, err
	}
	a.V(0).Info("deleted vcs provider", "provider", provider, "subject", subject)
	return provider, nil
}

func (a *Service) BeforeDeleteVCSProvider(hook func(context.Context, *Provider) error) {
	a.beforeDeleteHooks = append(a.beforeDeleteHooks, hook)
}
