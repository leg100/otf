package auth

import (
	"github.com/go-logr/logr"
	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/http/html"
	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/sql"
	"github.com/leg100/otf/internal/tfeapi"
	"github.com/leg100/otf/internal/tokens"
)

type (
	// Aliases to disambiguate service names when embedded together.
	OrganizationService organization.Service

	AuthService interface {
		TeamService
		UserService

		userTokenService
		teamTokenService
	}

	service struct {
		logr.Logger

		site         internal.Authorizer // authorizes site access
		organization internal.Authorizer // authorizes org access
		team         internal.Authorizer // authorizes team access

		db     *pgdb
		web    *webHandlers
		tfeapi *tfe
		api    *api

		*userTokenFactory
		*teamTokenFactory
	}

	Options struct {
		SiteToken string

		*sql.DB
		*tfeapi.Responder
		html.Renderer
		internal.HostnameService
		organization.OrganizationService
		tokens.TokensService
		logr.Logger
	}
)

func NewService(opts Options) *service {
	svc := service{
		Logger:       opts.Logger,
		organization: &organization.Authorizer{Logger: opts.Logger},
		site:         &internal.SiteAuthorizer{Logger: opts.Logger},
		team:         &TeamAuthorizer{Logger: opts.Logger},
		db:           newDB(opts.DB, opts.Logger),
		userTokenFactory: &userTokenFactory{
			TokensService: opts.TokensService,
		},
		teamTokenFactory: &teamTokenFactory{
			TokensService: opts.TokensService,
		},
	}
	svc.web = &webHandlers{
		Renderer: opts.Renderer,
		svc:      &svc,
	}
	svc.tfeapi = &tfe{
		AuthService: &svc,
		Responder:   opts.Responder,
	}
	svc.api = &api{
		AuthService: &svc,
		Responder:   opts.Responder,
	}

	// Whenever an organization is created, also create an owners team.
	opts.OrganizationService.AfterCreateOrganization(svc.createOwnersTeam)
	// Fetch users when API calls request users be included in the
	// response
	opts.Responder.Register(tfeapi.IncludeUsers, svc.tfeapi.includeUsers)
	// Register site token and site admin with the auth middleware, to permit
	// the latter to authenticate using the former.
	opts.TokensService.RegisterSiteToken(opts.SiteToken, &SiteAdmin)

	return &svc
}

func (a *service) AddHandlers(r *mux.Router) {
	a.web.addHandlers(r)
	a.tfeapi.addHandlers(r)
	a.api.addHandlers(r)
}
