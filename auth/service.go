package auth

import (
	"context"
	"net/http"
	"time"

	"github.com/go-logr/logr"
	"github.com/gorilla/mux"
	"github.com/leg100/otf"
	"github.com/leg100/otf/organization"
	"github.com/lestrrat-go/jwx/v2/jwk"
)

const (
	// default user session expiry
	defaultExpiry = 24 * time.Hour

	userSessionKind     authKind = "user_session"
	registrySessionKind authKind = "registry_session"
	agentTokenKind      authKind = "agent_token"
	userTokenKind       authKind = "user_token"
)

type (
	// Aliases to disambiguate service names when embedded together.
	OrganizationService organization.Service

	AuthService interface {
		AgentTokenService
		TeamService
		tokenService
		UserService

		StartSession(w http.ResponseWriter, r *http.Request, opts StartUserSessionOptions) error
		CreateRegistryToken(ctx context.Context, opts CreateRegistryTokenOptions) ([]byte, error)
	}

	service struct {
		logr.Logger

		site         otf.Authorizer // authorizes site access
		organization otf.Authorizer // authorizes org access

		api *api
		db  *pgdb
		web *webHandlers

		key jwk.Key
	}

	Options struct {
		SiteToken string
		Secret    string

		otf.DB
		otf.Renderer
		otf.HostnameService
		logr.Logger
	}

	// the kind of authentication token: user session, user token, agent token, etc
	authKind string
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
	key, err := jwk.FromRaw([]byte(opts.Secret))
	if err != nil {
		return nil, err
	}
	svc.key = key

	return &svc, nil
}

func (a *service) AddHandlers(r *mux.Router) {
	a.api.addHandlers(r)
	a.web.addHandlers(r)
}
