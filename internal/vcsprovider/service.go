package vcsprovider

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/github"
	"github.com/leg100/otf/internal/http/html"
	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/rbac"
	"github.com/leg100/otf/internal/sql"
	"github.com/leg100/otf/internal/sql/pggen"
	"github.com/leg100/otf/internal/tfeapi"
	"github.com/leg100/otf/internal/vcs"
)

type (
	Service struct {
		logr.Logger

		site              internal.Authorizer
		organization      internal.Authorizer
		db                *pgdb
		web               *webHandlers
		api               *tfe
		beforeDeleteHooks []func(context.Context, *VCSProvider) error
		githubapps        *github.Service

		*internal.HostnameService
		*factory
	}

	Options struct {
		*internal.HostnameService
		*sql.DB
		*tfeapi.Responder
		html.Renderer
		logr.Logger
		vcs.Subscriber

		GithubAppService    *github.Service
		GithubHostname      string
		GitlabHostname      string
		SkipTLSVerification bool
	}
)

func NewService(opts Options) *Service {
	factory := factory{
		githubapps:          opts.GithubAppService,
		githubHostname:      opts.GithubHostname,
		gitlabHostname:      opts.GitlabHostname,
		skipTLSVerification: opts.SkipTLSVerification,
	}
	svc := Service{
		Logger:          opts.Logger,
		HostnameService: opts.HostnameService,
		githubapps:      opts.GithubAppService,
		site:            &internal.SiteAuthorizer{Logger: opts.Logger},
		organization:    &organization.Authorizer{Logger: opts.Logger},
		factory:         &factory,
		db: &pgdb{
			DB:      opts.DB,
			factory: &factory,
		},
	}
	svc.web = &webHandlers{
		Renderer:        opts.Renderer,
		HostnameService: opts.HostnameService,
		GithubHostname:  opts.GithubHostname,
		GitlabHostname:  opts.GitlabHostname,
		client:          &svc,
		githubApps:      opts.GithubAppService,
	}
	svc.api = &tfe{
		Service:   &svc,
		Responder: opts.Responder,
	}
	// delete vcs providers when a github app is uninstalled
	opts.Subscribe(func(event vcs.Event) {
		// ignore events other than uninstallation events
		if event.Type != vcs.EventTypeInstallation || event.Action != vcs.ActionDeleted {
			return
		}
		// create user with unlimited permissions
		user := &internal.Superuser{Username: "vcs-provider-service"}
		ctx := internal.AddSubjectToContext(context.Background(), user)
		// list all vcsproviders using the app install
		providers, err := svc.ListVCSProvidersByGithubAppInstall(ctx, *event.GithubAppInstallID)
		if err != nil {
			return
		}
		// and delete them
		for _, prov := range providers {
			if _, err = svc.Delete(ctx, prov.ID); err != nil {
				return
			}
		}
	})
	return &svc
}

func (a *Service) AddHandlers(r *mux.Router) {
	a.web.addHandlers(r)
	a.api.addHandlers(r)
}

func (a *Service) Create(ctx context.Context, opts CreateOptions) (*VCSProvider, error) {
	subject, err := a.organization.CanAccess(ctx, rbac.CreateVCSProviderAction, opts.Organization)
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

func (a *Service) Update(ctx context.Context, id string, opts UpdateOptions) (*VCSProvider, error) {
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

func (a *Service) List(ctx context.Context, organization string) ([]*VCSProvider, error) {
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

func (a *Service) ListAllVCSProviders(ctx context.Context) ([]*VCSProvider, error) {
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

// ListVCSProvidersByGithubAppInstall is unauthenticated: only for internal use.
func (a *Service) ListVCSProvidersByGithubAppInstall(ctx context.Context, installID int64) ([]*VCSProvider, error) {
	subject, err := internal.SubjectFromContext(ctx)
	if err != nil {
		return nil, err
	}

	providers, err := a.db.listByGithubAppInstall(ctx, installID)
	if err != nil {
		a.Error(err, "listing github app installation vcs providers", "subject", subject, "install", installID)
		return nil, err
	}
	a.V(9).Info("listed github app installation vcs providers", "count", len(providers), "subject", subject, "install", installID)
	return providers, nil
}

func (a *Service) Get(ctx context.Context, id string) (*VCSProvider, error) {
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

func (a *Service) GetVCSClient(ctx context.Context, providerID string) (vcs.Client, error) {
	provider, err := a.Get(ctx, providerID)
	if err != nil {
		return nil, err
	}
	return provider.NewClient()
}

func (a *Service) Delete(ctx context.Context, id string) (*VCSProvider, error) {
	var (
		provider *VCSProvider
		subject  internal.Subject
	)
	err := a.db.Tx(ctx, func(ctx context.Context, q pggen.Querier) (err error) {
		// retrieve vcs provider first in order to get organization for authorization
		provider, err = a.db.get(ctx, id)
		if err != nil {
			a.Error(err, "retrieving vcs provider", "id", id)
			return err
		}

		subject, err = a.organization.CanAccess(ctx, rbac.DeleteVCSProviderAction, provider.Organization)
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

func (a *Service) BeforeDeleteVCSProvider(hook func(context.Context, *VCSProvider) error) {
	a.beforeDeleteHooks = append(a.beforeDeleteHooks, hook)
}
