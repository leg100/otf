package auth

import (
	"github.com/go-logr/logr"
	"github.com/gorilla/mux"
	internal "github.com/leg100/otf"
	"github.com/leg100/otf/http/html"
	"github.com/leg100/otf/organization"
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
	}

	Options struct {
		internal.DB
		html.Renderer
		internal.HostnameService
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
	return &svc
}

func (a *service) AddHandlers(r *mux.Router) {
	a.web.addHandlers(r)
}
