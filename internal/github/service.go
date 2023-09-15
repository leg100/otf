package github

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/cloud"
	"github.com/leg100/otf/internal/http/html"
	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/rbac"
	"github.com/leg100/otf/internal/sql"
)

type (
	// Alias services so they don't conflict when nested together in struct
	GithubAppService Service
	CloudService     cloud.Service

	Service interface {
		CreateGithubApp(ctx context.Context, opts CreateAppOptions) (*GithubApp, error)
		GetGithubApp(ctx context.Context) (*GithubApp, error)
		DeleteGithubApp(ctx context.Context) error

		CreateGithubInstall(ctx context.Context, opts CreateInstallOptions) (*Install, error)
		DeleteGithubInstall(ctx context.Context, installID string) (*Install, error)
	}

	service struct {
		logr.Logger

		site         internal.Authorizer
		organization internal.Authorizer
		db           *pgdb
		web          *webHandlers
	}

	Options struct {
		CloudService
		internal.HostnameService
		*sql.DB
		html.Renderer
		logr.Logger
	}

	CreateAppOptions struct {
		Organization  string
		AppID         int64
		WebhookSecret string
		PrivateKey    string
	}

	CreateInstallOptions struct {
		InstallID    int64  // github's install id
		AppID        string // otf's app id
		Organization string // name of otf org to install into
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
		CloudService:    opts.CloudService,
		Renderer:        opts.Renderer,
		HostnameService: opts.HostnameService,
		svc:             &svc,
	}
	return &svc
}

func (a *service) AddHandlers(r *mux.Router) {
	a.web.addHandlers(r)
}

func (a *service) CreateGithubApp(ctx context.Context, opts CreateAppOptions) (*GithubApp, error) {
	subject, err := a.organization.CanAccess(ctx, rbac.CreateGithubAppAction, opts.Organization)
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

func (a *service) GetGithubApp(ctx context.Context) (*GithubApp, error) {
	subject, err := a.site.CanAccess(ctx, rbac.GetGithubAppAction, "")
	if err != nil {
		return nil, err
	}

	app, err := a.db.get(ctx)
	if err != nil {
		a.Error(err, "retrieving github app")
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

func (a *service) CreateGithubInstall(ctx context.Context, opts CreateInstallOptions) (*Install, error) {
	subject, err := a.organization.CanAccess(ctx, rbac.CreateGithubAppInstallAction, opts.Organization)
	if err != nil {
		return nil, err
	}

	app, err := a.db.get(ctx)
	if err != nil {
		a.Error(err, "retrieving github app", "id", opts.AppID)
		return nil, err
	}

	install := newInstall(opts.InstallID, app)

	if err := a.db.createInstall(ctx, install); err != nil {
		a.Error(err, "creating github install", "install", install, "subject", subject)
		return nil, err
	}
	a.V(0).Info("created github install", "install", install, "subject", subject)
	return &install, nil
}

func (a *service) DeleteGithubInstall(ctx context.Context, installID string) (*Install, error) {
	subject, err := a.organization.CanAccess(ctx, rbac.DeleteGithubAppInstallAction, opts.Organization)
	if err != nil {
		return nil, err
	}

	app, err := a.db.get(ctx)
	if err != nil {
		a.Error(err, "retrieving github app", "id", opts.AppID)
		return nil, err
	}

	install := newInstall(opts.InstallID, app)

	if err := a.db.createInstall(ctx, install); err != nil {
		a.Error(err, "creating github install", "install", install, "subject", subject)
		return nil, err
	}
	a.V(0).Info("created github install", "install", install, "subject", subject)
	return &install, nil
}
