package tokens

import (
	"github.com/go-logr/logr"
	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/auth"
	"github.com/leg100/otf/internal/http/html"
	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/sql"
	"github.com/leg100/otf/internal/tfeapi"
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
		organizationTokenService
	}

	service struct {
		logr.Logger

		site         internal.Authorizer // authorizes site access
		organization internal.Authorizer // authorizes org access

		db  *pgdb
		web *webHandlers
		api *tfe

		middleware mux.MiddlewareFunc

		key jwk.Key
	}

	Options struct {
		logr.Logger
		*sql.DB
		*tfeapi.Responder
		html.Renderer
		auth.AuthService
		GoogleIAPConfig

		SiteToken string
		Secret    []byte
	}
)

func NewService(opts Options) (*service, error) {
	svc := service{
		Logger:       opts.Logger,
		organization: &organization.Authorizer{Logger: opts.Logger},
		site:         &internal.SiteAuthorizer{Logger: opts.Logger},
		db:           &pgdb{opts.DB},
	}
	svc.web = &webHandlers{
		Renderer:  opts.Renderer,
		svc:       &svc,
		siteToken: opts.SiteToken,
	}
	svc.api = &tfe{
		TokensService: &svc,
		Responder:     opts.Responder,
	}
	key, err := jwk.FromRaw([]byte(opts.Secret))
	if err != nil {
		return nil, err
	}
	svc.key = key
	svc.middleware = newMiddleware(middlewareOptions{
		agentTokenService:        &svc,
		organizationTokenService: &svc,
		AuthService:              opts.AuthService,
		GoogleIAPConfig:          opts.GoogleIAPConfig,
		SiteToken:                opts.SiteToken,
		key:                      key,
	})

	return &svc, nil
}

func (a *service) AddHandlers(r *mux.Router) {
	a.web.addHandlers(r)
	a.api.addHandlers(r)
}

// Middleware returns middleware for authenticating tokens
func (a *service) Middleware() mux.MiddlewareFunc { return a.middleware }
