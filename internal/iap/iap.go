// Package iap contains Google Cloud IAP stuff.
package iap

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/leg100/otf/internal/authz"
	"google.golang.org/api/idtoken"
)

// HTTP header in Google Cloud IAP request containing JWT
const header string = "x-goog-iap-jwt-assertion"

// Authenticator authenticates Google IAP requests.
type Authenticator struct {
	Audience string
	Client   AuthenticatorClient
}

type AuthenticatorClient interface {
	GetOrCreateUser(ctx context.Context, username string) (authz.Subject, error)
}

func (a *Authenticator) Authenticate(w http.ResponseWriter, r *http.Request) (authz.Subject, error) {
	token := r.Header.Get(header)
	if token == "" {
		// Not an IAP request.
		return nil, nil
	}
	payload, err := idtoken.Validate(r.Context(), token, a.Audience)
	if err != nil {
		return nil, err
	}
	email, ok := payload.Claims["email"]
	if !ok {
		return nil, errors.New("IAP token is missing email claim")
	}
	emailString, ok := email.(string)
	if !ok {
		return nil, fmt.Errorf("expected IAP token email to be a string: %#v", email)
	}
	return a.Client.GetOrCreateUser(r.Context(), emailString)
}
