package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/leg100/otf"
	"github.com/leg100/otf/rbac"
	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwt"
)

func (a *service) CreateRegistrySessionToken(ctx context.Context, opts CreateRegistrySessionOptions) ([]byte, error) {
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
