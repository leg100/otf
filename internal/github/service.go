package github

import (
	"context"
	"errors"
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

		GithubHostname string

		db  *pgdb
		web *webHandlers
	}

	Options struct {
		*sql.DB
		logr.Logger
		vcs.Publisher
		*internal.HostnameService

		GithubHostname      string
		SkipTLSVerification bool
		Authorizer          *authz.Authorizer
		VCSService          *vcs.Service
		VCSEventBroker      *vcs.Broker
	}
)

func NewService(opts Options) *Service {
	svc := Service{
		Logger:         opts.Logger,
		GithubHostname: opts.GithubHostname,
		Authorizer:     opts.Authorizer,
		db:             &pgdb{opts.DB},
	}
	svc.web = &webHandlers{
		authorizer:      opts.Authorizer,
		HostnameService: opts.HostnameService,
		GithubHostname:  opts.GithubHostname,
		GithubSkipTLS:   opts.SkipTLSVerification,
		svc:             &svc,
	}
	provider := provider{
		service:             &svc,
		db:                  svc.db,
		hostname:            opts.GithubHostname,
		skipTLSVerification: opts.SkipTLSVerification,
	}
	// Register providers
	opts.VCSService.RegisterKind(vcs.ProviderKind{
		Kind:             GithubAppKind,
		Name:             "GitHub (App)",
		Icon:             Icon(),
		Hostname:         opts.GithubHostname,
		InstallationKind: &provider,
		NewClient:        provider.NewClient,
		// Github apps don't need webhooks on repositories.
		SkipRepohook: true,
	})
	opts.VCSService.RegisterKind(vcs.ProviderKind{
		Kind:     GithubTokenKind,
		Name:     "GitHub (Token)",
		Icon:     Icon(),
		Hostname: opts.GithubHostname,
		TokenKind: &vcs.TokenKind{
			Description: tokenDescription(opts.GithubHostname),
		},
		NewClient: provider.NewClient,
	})
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

	app := newApp(opts)

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

func (a *Service) ListInstallations(ctx context.Context) ([]*Installation, error) {
	app, err := a.db.get(ctx)
	if errors.Is(err, internal.ErrResourceNotFound) {
		return nil, nil
	} else if err != nil {
		return nil, fmt.Errorf("retrieving github app: %w", err)
	}
	client, err := a.newClient(app)
	if err != nil {
		return nil, err
	}
	from, err := client.ListInstallations(ctx)
	if err != nil {
		return nil, fmt.Errorf("listing github app installs: %w", err)
	}
	to := make([]*Installation, len(from))
	for i, f := range from {
		to[i] = &Installation{Installation: f}
	}
	return to, nil
}

func (a *Service) GetInstallCredentials(ctx context.Context, installID int64) (*InstallCredentials, error) {
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

func (a *Service) DeleteInstallation(ctx context.Context, installID int64) error {
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

func (a *Service) newClient(app *App) (*Client, error) {
	return NewClient(ClientOptions{
		Hostname:            a.GithubHostname,
		SkipTLSVerification: true,
		AppCredentials: &AppCredentials{
			ID:         app.ID,
			PrivateKey: app.PrivateKey,
		},
	})
}
