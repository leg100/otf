package auth

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/gorilla/mux"
	"github.com/leg100/otf"
	"github.com/leg100/otf/cloud"
	"github.com/leg100/otf/organization"
)

type (
	AuthService interface {
		AgentTokenService
		RegistrySessionService
		sessionService
		TeamService
		tokenService
		UserService

		StartExpirer(context.Context)
		otf.Handlers
	}

	service struct {
		logr.Logger

		*synchroniser

		api          *api
		db           *pgdb
		organization otf.Authorizer
		web          *webHandlers
	}

	Options struct {
		Configs   []cloud.CloudOAuthConfig
		SiteToken string

		OrganizationService
		otf.DB
		otf.Renderer
		otf.HostnameService
		logr.Logger
	}

	OrganizationService organization.Service
)

func NewService(opts Options) (*service, error) {
	svc := service{Logger: opts.Logger}

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

	svc.synchroniser = &synchroniser{opts.Logger, opts.OrganizationService, &svc}
	svc.api = &api{svc: &svc}
	svc.db = db
	svc.organization = &organization.Authorizer{opts.Logger}
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
