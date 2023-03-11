package auth

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/gorilla/mux"
	"github.com/leg100/otf"
	"github.com/leg100/otf/cloud"
	"github.com/leg100/otf/organization"
)

type service interface {
	agentTokenService
	registrySessionService
	sessionService
	teamService
	tokenService
	userService
}

type Service struct {
	logr.Logger
	TokenMiddleware, SessionMiddleware mux.MiddlewareFunc

	*synchroniser

	api          *api
	db           *pgdb
	organization otf.Authorizer
	web          *webHandlers
}

func NewService(ctx context.Context, opts Options) (*Service, error) {
	svc := Service{Logger: opts.Logger}
	svc.TokenMiddleware = AuthenticateToken(&svc)
	svc.SessionMiddleware = AuthenticateSession(&svc)

	authenticators, err := newAuthenticators(authenticatorOptions{
		Logger:          opts.Logger,
		HostnameService: opts.HostnameService,
		service:         &svc,
		configs:         opts.Configs,
	})
	if err != nil {
		return nil, err
	}

	db := newDB(opts.DB, opts.Logger)
	// purge expired sessions
	go db.startExpirer(ctx, defaultExpiry)

	svc.synchroniser = &synchroniser{opts.Logger, opts.Service, &svc}
	svc.api = &api{app: &svc}
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

type Options struct {
	Configs   []*cloud.CloudOAuthConfig
	SiteToken string

	organization.Service
	otf.DB
	otf.Renderer
	otf.HostnameService
	logr.Logger
}

func (a *Service) AddHandlers(r *mux.Router) {
	a.api.addHandlers(r)
	a.web.addHandlers(r)
}
