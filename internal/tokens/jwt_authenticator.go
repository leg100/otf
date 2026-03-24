package tokens

import (
	"context"
	"fmt"
	"net/http"

	"github.com/leg100/otf/internal/authz"
)

// JWTAuthenticator authenticates requests possessing a header with a JWT
// token (i.e. API requests).
type JWTAuthenticator struct {
	Client JWTAuthenticatorClient
}

type JWTAuthenticatorClient interface {
	GetSubject(ctx context.Context, token []byte) (authz.Subject, error)
}

func (a *JWTAuthenticator) Authenticate(w http.ResponseWriter, r *http.Request) (authz.Subject, error) {
	token, err := ParseBearerToken(r)
	if err != nil {
		return nil, fmt.Errorf("parsing bearer token: %w", err)
	}
	if token == "" {
		// No Authorization header found.
		return nil, nil
	}
	return a.Client.GetSubject(r.Context(), []byte(token))
}
