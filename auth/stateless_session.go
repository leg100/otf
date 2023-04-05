package auth

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/go-logr/logr"
	"github.com/leg100/otf"
	"github.com/leg100/otf/http/html"
	"github.com/leg100/otf/rbac"
	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jwt"
)

const (
	userSessionKind     tokenKind = "user"
	registrySessionKind tokenKind = "registry"
)

type (
	StatelessSessionService interface {
		StartSession(w http.ResponseWriter, r *http.Request, opts CreateStatelessSessionOptions) error
		CreateRegistrySessionToken(ctx context.Context, opts CreateRegistrySessionOptions) ([]byte, error)
	}
	CreateStatelessSessionOptions struct {
		Username *string
		Expiry   *time.Time
	}
	statelessSessionService struct {
		logr.Logger
		key          jwk.Key
		organization otf.Authorizer
	}

	// the kind of authentication token: user session, user token, agent token, etc
	tokenKind string
)

func newStatelessSessionService(logger logr.Logger, secret string) (*statelessSessionService, error) {
	key, err := jwk.FromRaw([]byte(secret))
	if err != nil {
		return nil, err
	}
	return &statelessSessionService{Logger: logger, key: key}, nil
}

func (a *statelessSessionService) StartSession(w http.ResponseWriter, r *http.Request, opts CreateStatelessSessionOptions) error {
	if opts.Username == nil {
		return fmt.Errorf("missing username")
	}
	expiry := otf.CurrentTimestamp().Add(defaultExpiry)
	if opts.Expiry != nil {
		expiry = *opts.Expiry
	}

	token, err := jwt.NewBuilder().
		Subject(*opts.Username).
		Claim("kind", userSessionKind).
		IssuedAt(time.Now()).
		Expiration(expiry).
		Build()
	if err != nil {
		return err
	}
	serialized, err := jwt.Sign(token, jwt.WithKey(jwa.HS256, a.key))
	if err != nil {
		return err
	}

	// Set cookie to expire at same time as token
	html.SetCookie(w, sessionCookie, string(serialized), otf.Time(expiry))
	html.ReturnUserOriginalPage(w, r)

	a.V(2).Info("started session", "username", *opts.Username)

	return nil
}

func (a *statelessSessionService) CreateRegistrySessionToken(ctx context.Context, opts CreateRegistrySessionOptions) ([]byte, error) {
	if opts.Organization == nil {
		return nil, fmt.Errorf("missing organization")
	}

	subject, err := a.organization.CanAccess(ctx, rbac.CreateRegistrySessionAction, *opts.Organization)
	if err != nil {
		return nil, err
	}

	expiry := otf.CurrentTimestamp().Add(defaultRegistrySessionExpiry)
	if opts.Expiry != nil {
		expiry = *opts.Expiry
	}

	token, err := jwt.NewBuilder().
		Claim("kind", registrySessionKind).
		Claim("organization", *opts.Organization).
		IssuedAt(time.Now()).
		Expiration(expiry).
		Build()
	if err != nil {
		return nil, err
	}
	serialized, err := jwt.Sign(token, jwt.WithKey(jwa.HS256, a.key))
	if err != nil {
		return nil, err
	}

	a.V(2).Info("created registry session", "subject", subject, "run")

	return serialized, nil
}
