package tokens

import (
	"github.com/go-logr/logr"
	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/sql"
	"github.com/lestrrat-go/jwx/v2/jwk"
)

type (
	TokensService interface {
		Middleware() mux.MiddlewareFunc
		NewToken(NewTokenOptions) ([]byte, error)
		RegisterKind(Kind, SubjectGetter)
		RegisterSiteToken(token string, siteAdmin internal.Subject)

		sessionService
	}

	service struct {
		logr.Logger
		*factory
		*registry
		*sessionFactory

		site internal.Authorizer // authorizes site access

		middleware mux.MiddlewareFunc
	}

	Options struct {
		logr.Logger
		*sql.DB
		GoogleIAPConfig

		Secret []byte
	}
)

func NewService(opts Options) (*service, error) {
	svc := service{
		Logger: opts.Logger,
		site:   &internal.SiteAuthorizer{Logger: opts.Logger},
	}
	key, err := jwk.FromRaw([]byte(opts.Secret))
	if err != nil {
		return nil, err
	}
	svc.factory = &factory{key: key}
	svc.sessionFactory = &sessionFactory{factory: svc.factory}
	svc.registry = &registry{
		kinds: make(map[Kind]SubjectGetter),
	}
	svc.middleware = newMiddleware(middlewareOptions{
		GoogleIAPConfig: opts.GoogleIAPConfig,
		key:             key,
	})

	return &svc, nil
}

// Middleware returns middleware for authenticating tokens
func (a *service) Middleware() mux.MiddlewareFunc { return a.middleware }
