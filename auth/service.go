package auth

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/gorilla/mux"
	"github.com/leg100/otf"
	"github.com/leg100/otf/cloud"
	"github.com/leg100/otf/organization"
	"github.com/leg100/otf/orgcreator"
)

type (
	// Aliases to disambiguate service names when embedded together.
	OrganizationService        organization.Service
	OrganizationCreatorService orgcreator.Service

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

		*synchroniser

		site         otf.Authorizer // authorizes site access
		organization otf.Authorizer // authorizes org access

		api *api
		db  *pgdb
		web *webHandlers
	}

	Options struct {
		Configs   []cloud.CloudOAuthConfig
		SiteToken string

		OrganizationService
		OrganizationCreatorService
		otf.DB
		otf.Renderer
		otf.HostnameService
		logr.Logger
	}
)

func NewService(opts Options) (*service, error) {
	svc := service{
		Logger:       opts.Logger,
		organization: &organization.Authorizer{opts.Logger},
		site:         &otf.SiteAuthorizer{opts.Logger},
	}

	authenticators, err := newAuthenticators(authenticatorOptions{
		Logger:          opts.Logger,
		HostnameService: opts.HostnameService,
		AuthService:     &svc,
		configs:         opts.Configs,
	})
	if err != nil {
		return nil, err
	}

	db := newDB(opts.DB, opts.Logger)

	svc.synchroniser = &synchroniser{
		Logger:                     opts.Logger,
		OrganizationService:        opts.OrganizationService,
		OrganizationCreatorService: opts.OrganizationCreatorService,
		AuthService:                &svc,
	}
	svc.api = &api{svc: &svc}
	svc.db = db
	svc.web = &webHandlers{
		Renderer:       opts.Renderer,
		svc:            &svc,
		authenticators: authenticators,
		siteToken:      opts.SiteToken,
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
