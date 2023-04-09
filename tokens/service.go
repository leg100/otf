package tokens

import (
	"github.com/go-logr/logr"
	"github.com/gorilla/mux"
	"github.com/leg100/otf"
	"github.com/leg100/otf/auth"
	"github.com/leg100/otf/organization"
	"github.com/lestrrat-go/jwx/v2/jwk"
)

// session cookie stores the session token
const sessionCookie = "session"

type (
	// Aliases to disambiguate service names when embedded together.
	OrganizationService organization.Service

	TokensService interface {
		Middleware() mux.MiddlewareFunc

		agentTokenService
		RunTokenService
		sessionService
		userTokenService
	}

	service struct {
		logr.Logger

		site         otf.Authorizer // authorizes site access
		organization otf.Authorizer // authorizes org access

		api *api
		db  *pgdb
		web *webHandlers

		middleware mux.MiddlewareFunc

		key jwk.Key
	}

	Options struct {
		logr.Logger
		otf.DB
		otf.Renderer
		auth.AuthService
		GoogleIAPConfig

		SiteToken string
		Secret    string
	}
)

func NewService(opts Options) (*service, error) {
	svc := service{
		Logger:       opts.Logger,
		organization: &organization.Authorizer{Logger: opts.Logger},
		site:         &otf.SiteAuthorizer{Logger: opts.Logger},
		db:           &pgdb{opts.DB},
	}
	svc.api = &api{svc: &svc}
	svc.web = &webHandlers{
		Renderer:  opts.Renderer,
		svc:       &svc,
		siteToken: opts.SiteToken,
	}
	key, err := jwk.FromRaw([]byte(opts.Secret))
	if err != nil {
		return nil, err
	}
	svc.key = key
	svc.middleware = newMiddleware(middlewareOptions{
		agentTokenService: &svc,
		AuthService:       opts.AuthService,
		GoogleIAPConfig:   opts.GoogleIAPConfig,
		SiteToken:         opts.SiteToken,
		key:               key,
	})

	return &svc, nil
}

func (a *service) AddHandlers(r *mux.Router) {
	a.api.addHandlers(r)
	a.web.addHandlers(r)
}

// Middleware returns middleware for authenticating tokens
func (a *service) Middleware() mux.MiddlewareFunc { return a.middleware }
