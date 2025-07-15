package dynamiccreds

import (
	"crypto/x509"
	"encoding/pem"
	"errors"
	"os"

	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/logr"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"golang.org/x/crypto/ssh"
)

type (
	Service struct {
		handlers   *Handlers
		privateKey jwk.Key
	}

	Options struct {
		Logger          logr.Logger
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
		// parse and assign public key
		{
			raw, err := os.ReadFile(opts.PublicKeyPath)
			if err != nil {
				return nil, err
			}
			decoded, _ := pem.Decode(raw)
			parsed, err := x509.ParsePKIXPublicKey(decoded.Bytes)
			if err != nil {
				return nil, err
			}
			key, err := jwk.FromRaw(parsed)
			if err != nil {
				return nil, err
			}
			svc.handlers = &Handlers{
				hostnameService: opts.HostnameService,
				publicKey:       key,
			}
		}
		// parse and assign private key
		{
			raw, err := os.ReadFile(opts.PrivateKeyPath)
			if err != nil {
				return nil, err
			}
			parsed, err := ssh.ParseRawPrivateKey(raw)
			if err != nil {
				return nil, err
			}
			key, err := jwk.FromRaw(parsed)
			if err != nil {
				return nil, err
			}
			svc.privateKey = key
		}
		opts.Logger.Info("enabled dynamic provider credentials")
	}

	return &svc, nil
}

func (s *Service) PrivateKey() jwk.Key { return s.privateKey }

func (s *Service) AddHandlers(r *mux.Router) {
	if s.privateKey != nil {
		s.handlers.addHandlers(r)
	}
}
