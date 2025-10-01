package github

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/authz"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/sql"
	"github.com/leg100/otf/internal/vcs"
)

type (
	// Service is the service for github app management
	Service struct {
		logr.Logger
		*authz.Authorizer

		db  *appDB
		web *webHandlers
	}

	Options struct {
		*sql.DB
		logr.Logger
		vcs.Publisher
		*internal.HostnameService

		GithubAPIURL        *internal.WebURL
		SkipTLSVerification bool
		Authorizer          *authz.Authorizer
		VCSService          *vcs.Service
		VCSEventBroker      *vcs.Broker
	}
)

func NewService(opts Options) *Service {
	svc := Service{
		Logger:     opts.Logger,
		Authorizer: opts.Authorizer,
		db: &appDB{
			DB:                  opts.DB,
			baseURL:             opts.GithubAPIURL,
			skipTLSVerification: opts.SkipTLSVerification,
		},
	}
	svc.web = &webHandlers{
		authorizer:          opts.Authorizer,
		HostnameService:     opts.HostnameService,
		githubAPIURL:        opts.GithubAPIURL,
		svc:                 &svc,
		skipTLSVerification: opts.SkipTLSVerification,
	}
	registerVCSKinds(&svc, opts.VCSService, opts.GithubAPIURL, opts.SkipTLSVerification)

	// delete github app vcs providers when the app is uninstalled
	opts.VCSEventBroker.Subscribe(func(event vcs.Event) {
		// ignore events other than uninstallation events
		if event.Type != vcs.EventTypeInstallation || event.Action != vcs.ActionDeleted {
			return
		}
		// create user with unlimited permissions
		user := &authz.Superuser{Username: "vcs-provider-service"}
		ctx := authz.AddSubjectToContext(context.Background(), user)
		// list all vcsproviders using the app install
		providers, err := opts.VCSService.ListByInstall(ctx, *event.GithubAppInstallID)
		if err != nil {
			return
		}
		// and delete them
		for _, prov := range providers {
			if _, err = opts.VCSService.Delete(ctx, prov.ID); err != nil {
				return
			}
		}
	})
	return &svc
}

func (a *Service) AddHandlers(r *mux.Router) {
	a.web.addHandlers(r)
}

func (a *Service) CreateApp(ctx context.Context, opts CreateAppOptions) (*App, error) {
	subject, err := a.Authorize(ctx, authz.CreateGithubAppAction, resource.SiteID)
	if err != nil {
		return nil, err
	}

	app, err := newApp(opts)
	if err != nil {
		a.Error(err, "creating github app", "app", app, "subject", subject)
		return nil, err
	}

	if err := a.db.create(ctx, app); err != nil {
		a.Error(err, "creating github app", "app", app, "subject", subject)
		return nil, err
	}
	a.V(0).Info("created github app", "app", app, "subject", subject)
	return app, nil
}

func (a *Service) GetApp(ctx context.Context) (*App, error) {
	subject, err := a.Authorize(ctx, authz.GetGithubAppAction, resource.SiteID)
	if err != nil {
		return nil, err
	}

	app, err := a.db.get(ctx)
	if err != nil {
		return nil, fmt.Errorf("retrieving github app: %w", err)
	}
	a.V(9).Info("retrieved github app", "app", app, "subject", subject)

	return app, nil
}

func (a *Service) DeleteApp(ctx context.Context) error {
	subject, err := a.Authorize(ctx, authz.DeleteGithubAppAction, resource.SiteID)
	if err != nil {
		return err
	}

	err = a.db.delete(ctx)
	if err != nil {
		a.Error(err, "deleting github app", "subject", subject)
		return err
	}
	a.V(0).Info("deleted github app", "subject", subject)
	return nil
}

func (a *Service) ListInstallations(ctx context.Context) ([]vcs.Installation, error) {
	app, err := a.db.get(ctx)
	if err != nil {
		return nil, err
	}
	return app.ListInstallations(ctx)
}

func (a *Service) GetInstallation(ctx context.Context, installID int64) (vcs.Installation, error) {
	app, err := a.db.get(ctx)
	if err != nil {
		return vcs.Installation{}, fmt.Errorf("retrieving github app: %w", err)
	}
	return app.GetInstallation(ctx, installID)
}

func (a *Service) DeleteInstallation(ctx context.Context, installID int64) error {
	app, err := a.db.get(ctx)
	if err != nil {
		return err
	}
	return app.DeleteInstallation(ctx, installID)
}
