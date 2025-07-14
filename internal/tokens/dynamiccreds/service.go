package dynamiccreds

import (
	"errors"
	"os"

	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal"
	"github.com/lestrrat-go/jwx/v2/jwk"
)

type (
	Service struct {
		handlers   *Handlers
		isEnabled  bool
		privateKey jwk.Key
	}

	Options struct {
		HostnameService *internal.HostnameService
		PublicKeyPath   string
		PrivateKeyPath  string
	}
)

func NewService(opts Options) (*Service, error) {
	var svc Service
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
		pubKey, err := jwk.FromRaw(pubKeyRaw)
		if err != nil {
			return nil, err
		}
		privKey, err := jwk.FromRaw(privKeyRaw)
		if err != nil {
			return nil, err
		}
		svc.privateKey = privKey
		svc.handlers = &Handlers{
			hostnameService: opts.HostnameService,
			publicKey:       pubKey,
		}
	}

	return &svc, nil
}

func (s *Service) PrivateKey() jwk.Key { return s.privateKey }

func (s *Service) AddHandlers(r *mux.Router) {
	if s.privateKey != nil {
		s.handlers.addHandlers(r)
	}
}
