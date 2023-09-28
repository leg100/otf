package github

import (
	"context"
	"errors"

	"github.com/go-logr/logr"
	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/http/html"
	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/rbac"
	"github.com/leg100/otf/internal/sql"
	"github.com/leg100/otf/internal/vcs"
)

type (
	// Alias services so they don't conflict when nested together in struct
	GithubAppService Service

	Service interface {
		CreateGithubApp(ctx context.Context, opts CreateAppOptions) (*App, error)
		// GetGithubApp returns the github app. If no github app has been
		// created then nil is returned without an error.
		GetGithubApp(ctx context.Context) (*App, error)
		DeleteGithubApp(ctx context.Context) error

		ListInstallations(ctx context.Context) ([]*Installation, error)
		DeleteInstallation(ctx context.Context, installID int64) error
	}

	service struct {
		logr.Logger

		site            internal.Authorizer
		organization    internal.Authorizer
		db              *pgdb
		web             *webHandlers
		appEventHandler *appEventHandler
	}

	Options struct {
		internal.HostnameService
		*sql.DB
		html.Renderer
		logr.Logger
		vcs.Publisher
	}
)

func NewService(opts Options) *service {
	svc := service{
		Logger:       opts.Logger,
		site:         &internal.SiteAuthorizer{Logger: opts.Logger},
		organization: &organization.Authorizer{Logger: opts.Logger},
		db:           &pgdb{opts.DB},
	}
	svc.web = &webHandlers{
		Renderer:        opts.Renderer,
		HostnameService: opts.HostnameService,
		svc:             &svc,
	}
	svc.appEventHandler = &appEventHandler{
		Service:   &svc,
		Publisher: opts.Publisher,
	}
	return &svc
}

func (a *service) AddHandlers(r *mux.Router) {
	a.web.addHandlers(r)
	a.appEventHandler.addHandlers(r)
}

func (a *service) CreateGithubApp(ctx context.Context, opts CreateAppOptions) (*App, error) {
	subject, err := a.site.CanAccess(ctx, rbac.CreateGithubAppAction, "")
	if err != nil {
		return nil, err
	}

	app := newApp(opts)

	if err := a.db.create(ctx, app); err != nil {
		a.Error(err, "creating github app", "app", app, "subject", subject)
		return nil, err
	}
	a.V(0).Info("created github app", "app", app, "subject", subject)
	return app, nil
}

func (a *service) GetGithubApp(ctx context.Context) (*App, error) {
	subject, err := a.site.CanAccess(ctx, rbac.GetGithubAppAction, "")
	if err != nil {
		return nil, err
	}

	app, err := a.db.get(ctx)
	if errors.Is(err, internal.ErrResourceNotFound) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	a.V(9).Info("retrieved github app", "app", app, "subject", subject)

	return app, nil
}

func (a *service) DeleteGithubApp(ctx context.Context) error {
	subject, err := a.site.CanAccess(ctx, rbac.DeleteGithubAppAction, "")
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

func (a *service) ListInstallations(ctx context.Context) ([]*Installation, error) {
	app, err := a.db.get(ctx)
	if errors.Is(err, internal.ErrResourceNotFound) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	client, err := NewClient(ctx, ClientOptions{
		AppCredentials: &AppCredentials{
			ID:         app.ID,
			PrivateKey: app.PrivateKey,
		},
	})
	if err != nil {
		return nil, err
	}
	from, err := client.ListInstallations(ctx)
	if err != nil {
		return nil, err
	}
	to := make([]*Installation, len(from))
	for i, f := range from {
		to[i] = &Installation{Installation: f}
	}
	return to, nil
}

func (a *service) DeleteInstallation(ctx context.Context, installID int64) error {
	app, err := a.db.get(ctx)
	if err != nil {
		return err
	}
	client, err := NewClient(ctx, ClientOptions{
		AppCredentials: &AppCredentials{
			ID:         app.ID,
			PrivateKey: app.PrivateKey,
		},
	})
	if err != nil {
		return err
	}
	if err := client.DeleteInstallation(ctx, installID); err != nil {
		return err
	}
	return nil
}
