package auth

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/gorilla/mux"
	"github.com/leg100/otf"
	"github.com/leg100/otf/cloud"
)

type app interface {
	agentTokenApp
	registrySessionApp
	sessionApp
	teamApp
	userApp
}

type Application struct {
	otf.Authorizer
	logr.Logger

	db db
	*api
	*web
	*synchroniser
}

func NewApplication(ctx context.Context, opts ApplicationOptions) (*Application, error) {
	db := newDB(opts.DB, opts.Logger)
	app := &Application{
		Authorizer: opts.Authorizer,
		Logger:     opts.Logger,
		db:         db,
	}
	app.synchroniser = &synchroniser{opts.Logger, opts.OrganizationService, app}

	authenticators, err := newAuthenticators(authenticatorOptions{
		Logger:          opts.Logger,
		HostnameService: opts.HostnameService,
		app:             app,
		configs:         opts.Configs,
	})
	if err != nil {
		return nil, err
	}

	app.api = &api{
		app: app,
	}
	app.web = &web{
		Renderer:       opts.Renderer,
		app:            app,
		authenticators: authenticators,
		siteToken:      opts.SiteToken,
	}

	// purge expired sessions
	go db.startExpirer(ctx, defaultExpiry)

	return app, nil
}

type ApplicationOptions struct {
	Configs   []*cloud.CloudOAuthConfig
	SiteToken string

	otf.OrganizationService
	otf.Authorizer
	otf.DB
	otf.Renderer
	otf.HostnameService
	logr.Logger
}

func NewApp(logger logr.Logger, db otf.DB, authorizer otf.Authorizer) *Application {
	return &Application{
		Logger:     logger,
		db:         newDB(db, logger),
		Authorizer: authorizer,
	}
}

func (h *Application) AddHandlers(r *mux.Router) {
	h.api.addHandlers(r)
	h.web.addHandlers(r)
}
