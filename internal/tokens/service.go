package tokens

import (
	"github.com/go-logr/logr"
	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal"
	"github.com/lestrrat-go/jwx/v2/jwk"
)

type (
	Service struct {
		logr.Logger
		*factory
		*registry
		*sessionFactory

		site internal.Authorizer // authorizes site access

		middleware mux.MiddlewareFunc
	}

	Options struct {
		logr.Logger
		GoogleIAPConfig

		Secret []byte
	}
)

func NewService(opts Options) (*Service, error) {
	svc := Service{
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
		Logger:          opts.Logger,
		GoogleIAPConfig: opts.GoogleIAPConfig,
		key:             key,
		registry:        svc.registry,
	})
	return &svc, nil
}

// Middleware returns middleware for authenticating tokens
func (a *Service) Middleware() mux.MiddlewareFunc { return a.middleware }
