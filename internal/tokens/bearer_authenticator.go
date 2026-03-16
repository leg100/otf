package tokens

import (
	"context"
	"net/http"

	"github.com/leg100/otf/internal/authz"
)

// bearerAuthenticator authenticates requests possessing a header with a bearer
// token (i.e. API requests).
type bearerAuthenticator struct {
	Client bearerAuthenticatorClient
}

type bearerAuthenticatorClient interface {
	GetSubject(ctx context.Context, token []byte) (authz.Subject, error)
}

func (a *bearerAuthenticator) Authenticate(w http.ResponseWriter, r *http.Request) (authz.Subject, error) {
	bearer := r.Header.Get("Authorization")
	if bearer == "" {
		return nil, nil
	}
	token, err := ParseBearerToken(bearer)
	if err != nil {
		return nil, err
	}
	return a.Client.GetSubject(r.Context(), []byte(token))
}
