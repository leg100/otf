package tokens

import (
	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal/logr"
	"github.com/leg100/otf/internal/resource"
	"github.com/lestrrat-go/jwx/v2/jwk"
)

type (
	Service struct {
		logr.Logger

		*tokenFactory
		*registry

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
	}
	key, err := jwk.FromRaw([]byte(opts.Secret))
	if err != nil {
		return nil, err
	}
	svc.tokenFactory = &tokenFactory{key: key}
	svc.registry = &registry{
		kinds: make(map[resource.Kind]SubjectGetter),
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
