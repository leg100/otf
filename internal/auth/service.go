package auth

import (
	"github.com/go-logr/logr"
	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/http/html"
	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/sql"
	"github.com/leg100/otf/internal/tfeapi"
)

type (
	// Aliases to disambiguate service names when embedded together.
	OrganizationService organization.Service

	AuthService interface {
		TeamService
		UserService
	}

	service struct {
		logr.Logger

		site         internal.Authorizer // authorizes site access
		organization internal.Authorizer // authorizes org access

		db  *pgdb
		web *webHandlers
		api *tfe
	}

	Options struct {
		*sql.DB
		*tfeapi.Responder
		html.Renderer
		internal.HostnameService
		organization.OrganizationService
		logr.Logger
	}
)

func NewService(opts Options) *service {
	svc := service{
		Logger:       opts.Logger,
		organization: &organization.Authorizer{Logger: opts.Logger},
		site:         &internal.SiteAuthorizer{Logger: opts.Logger},
		db:           newDB(opts.DB, opts.Logger),
	}
	svc.web = &webHandlers{
		Renderer: opts.Renderer,
		svc:      &svc,
	}
	svc.api = &tfe{
		AuthService: &svc,
		Responder:   opts.Responder,
	}

	// Whenever an organization is created, also create an owners team.
	opts.OrganizationService.AfterCreateOrganization(svc.createOwnersTeam)
	// Fetch users when API calls request users be included in the
	// response
	opts.Responder.Register(tfeapi.IncludeUsers, svc.api.includeUsers)

	return &svc
}

func (a *service) AddHandlers(r *mux.Router) {
	a.web.addHandlers(r)
	a.api.addHandlers(r)
}
