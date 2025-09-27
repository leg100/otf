package dynamiccreds

import (
	"crypto/x509"
	"encoding/pem"
	"errors"
	"os"

	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/logr"
	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/run"
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

			// Assign kid to public key. The AWS provider insists on this
			// being present on the JWKS endpoint.
			jwk.AssignKeyID(key)

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

			// Assign kid to private key. The Azure provider insists on this
			// being present in the headers of the generated JWT.
			jwk.AssignKeyID(key)

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

func (s *Service) GenerateToken(
	issuer string,
	organization organization.Name,
	workspaceID resource.TfeID,
	workspaceName string,
	runID resource.TfeID,
	phase run.PhaseType,
	audience string,
) ([]byte, error) {
	return generateToken(
		s.privateKey,
		issuer,
		organization,
		workspaceID,
		workspaceName,
		runID,
		phase,
		audience,
	)
}
