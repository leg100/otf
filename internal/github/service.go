package github

import (
	"context"
	"errors"
	"fmt"

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

		GetInstallCredentials(ctx context.Context, installID int64) (*InstallCredentials, error)
	}

	service struct {
		logr.Logger

		GithubHostname string

		site         internal.Authorizer
		organization internal.Authorizer
		db           *pgdb
		web          *webHandlers
	}

	Options struct {
		internal.HostnameService
		*sql.DB
		html.Renderer
		logr.Logger
		vcs.Publisher
		GithubHostname string
	}
)

func NewService(opts Options) *service {
	svc := service{
		Logger:         opts.Logger,
		GithubHostname: opts.GithubHostname,
		site:           &internal.SiteAuthorizer{Logger: opts.Logger},
		organization:   &organization.Authorizer{Logger: opts.Logger},
		db:             &pgdb{opts.DB},
	}
	svc.web = &webHandlers{
		Renderer:        opts.Renderer,
		HostnameService: opts.HostnameService,
		GithubHostname:  opts.GithubHostname,
		svc:             &svc,
	}
	return &svc
}

func (a *service) AddHandlers(r *mux.Router) {
	a.web.addHandlers(r)
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
	client, err := a.newClient(app)
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

func (a *service) GetInstallCredentials(ctx context.Context, installID int64) (*InstallCredentials, error) {
	app, err := a.db.get(ctx)
	if err != nil {
		return nil, err
	}
	client, err := a.newClient(app)
	if err != nil {
		return nil, err
	}
	install, err := client.GetInstallation(ctx, installID)
	if err != nil {
		return nil, err
	}
	creds := InstallCredentials{
		ID: installID,
		AppCredentials: AppCredentials{
			ID:         app.ID,
			PrivateKey: app.PrivateKey,
		},
	}
	switch install.GetTargetType() {
	case "Organization":
		creds.Organization = install.GetAccount().Login
	case "User":
		creds.User = install.GetAccount().Login
	default:
		return nil, fmt.Errorf("unexpected target type: %s", install.GetTargetType())
	}
	return &creds, nil
}

func (a *service) DeleteInstallation(ctx context.Context, installID int64) error {
	app, err := a.db.get(ctx)
	if err != nil {
		return err
	}
	client, err := a.newClient(app)
	if err != nil {
		return err
	}
	if err := client.DeleteInstallation(ctx, installID); err != nil {
		return err
	}
	return nil
}

func (a *service) newClient(app *App) (*Client, error) {
	return NewClient(ClientOptions{
		Hostname:            a.GithubHostname,
		SkipTLSVerification: true,
		AppCredentials: &AppCredentials{
			ID:         app.ID,
			PrivateKey: app.PrivateKey,
		},
	})
}
