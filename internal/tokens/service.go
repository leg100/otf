package tokens

import (
	"errors"
	"os"

	"github.com/go-logr/logr"
	"github.com/gorilla/mux"
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

		Secret         []byte
		PublicKeyPath  string
		PrivateKeyPath string
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

	var (
		pubKey  jwk.Key
		privKey jwk.Key
	)
	switch {
	case opts.PublicKeyPath != "" && opts.PrivateKeyPath == "":
		return nil, errors.New("must provide both private and public key paths")
	case opts.PublicKeyPath == "" && opts.PrivateKeyPath != "":
		return nil, errors.New("must provide both private and public key paths")
	case opts.PublicKeyPath != "" && opts.PrivateKeyPath != "":
		pubKeyRaw, err := os.ReadFile(opts.PublicKeyPath)
		if err != nil {
			return nil, err
		}
		privKeyRaw, err := os.ReadFile(opts.PrivateKeyPath)
		if err != nil {
			return nil, err
		}
		pubKey, err = jwk.FromRaw(pubKeyRaw)
		if err != nil {
			return nil, err
		}
		privKey, err = jwk.FromRaw(privKeyRaw)
		if err != nil {
			return nil, err
		}
	}

	svc.tokenFactory = &tokenFactory{
		symKey:     key,
		PrivateKey: privKey,
	}
	svc.registry = &registry{
		kinds: make(map[resource.Kind]SubjectGetter),
	}
	svc.middleware = newMiddleware(middlewareOptions{
		Logger:          opts.Logger,
		GoogleIAPConfig: opts.GoogleIAPConfig,
		key:             key,
		publicKey:       pubKey,
		registry:        svc.registry,
	})
	return &svc, nil
}

// Middleware returns middleware for authenticating tokens
func (a *Service) Middleware() mux.MiddlewareFunc { return a.middleware }
