package auth

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/gorilla/mux"
	"github.com/leg100/otf"
	"github.com/leg100/otf/organization"
)

type (
	// Aliases to disambiguate service names when embedded together.
	OrganizationService organization.Service

	AuthService interface {
		AgentTokenService
		RegistrySessionService
		sessionService
		TeamService
		tokenService
		UserService

		StartExpirer(context.Context)
	}

	service struct {
		logr.Logger

		site         otf.Authorizer // authorizes site access
		organization otf.Authorizer // authorizes org access

		api *api
		db  *pgdb
		web *webHandlers
	}

	Options struct {
		SiteToken string

		otf.DB
		otf.Renderer
		otf.HostnameService
		logr.Logger
	}
)

func NewService(opts Options) (*service, error) {
	svc := service{
		Logger:       opts.Logger,
		organization: &organization.Authorizer{Logger: opts.Logger},
		site:         &otf.SiteAuthorizer{Logger: opts.Logger},
		db:           newDB(opts.DB, opts.Logger),
	}
	svc.api = &api{svc: &svc}
	svc.web = &webHandlers{
		Renderer:  opts.Renderer,
		svc:       &svc,
		siteToken: opts.SiteToken,
	}

	return &svc, nil
}

func (a *service) StartExpirer(ctx context.Context) {
	// purge expired sessions on regular interval
	go a.db.startExpirer(ctx, defaultExpiry)
}

func (a *service) AddHandlers(r *mux.Router) {
	a.api.addHandlers(r)
	a.web.addHandlers(r)
}
