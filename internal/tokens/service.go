package tokens

import (
	"github.com/leg100/otf/internal/logr"
	"github.com/leg100/otf/internal/resource"
	"github.com/lestrrat-go/jwx/v2/jwk"
)

type (
	// Alias service to permit embedding it with other services in a struct
	// without a name clash.
	TokensService = Service

	Service struct {
		*registry
		*factory
		Middleware *Middleware
		logger     logr.Logger
	}

	Options struct {
		Logger logr.Logger
		Secret []byte
	}
)

func NewService(opts Options) (*Service, error) {
	svc := Service{
		logger: opts.Logger,
	}
	key, err := jwk.FromRaw([]byte(opts.Secret))
	if err != nil {
		return nil, err
	}
	svc.factory = &factory{key: key}
	svc.registry = &registry{
		kinds: make(map[resource.Kind]SubjectGetter),
		key:   key,
	}
	svc.Middleware = &Middleware{
		logger: opts.Logger,
		authenticators: []authenticator{
			&bearerAuthenticator{
				Client: &svc,
			},
		},
	}
	return &svc, nil
}

func (s *Service) AddAuthenticator(authenticator authenticator) {
	s.Middleware.authenticators = append(s.Middleware.authenticators, authenticator)
}
