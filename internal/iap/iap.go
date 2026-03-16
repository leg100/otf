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

type (
	// Authenticator authenticates Google IAP requests.
	Authenticator struct {
		audience  string
		users     UserClient
		validator tokenValidator
	}

	UserClient interface {
		GetOrCreateUser(ctx context.Context, username string) (authz.Subject, error)
	}

	tokenValidator interface {
		Validate(ctx context.Context, idToken string, audience string) (*idtoken.Payload, error)
	}

	tokenValidatorFunc func(context.Context, string, string) (*idtoken.Payload, error)
)

func (f tokenValidatorFunc) Validate(ctx context.Context, idToken string, audience string) (*idtoken.Payload, error) {
	return f(ctx, idToken, audience)
}

func NewAuthenticator(audience string, users UserClient) *Authenticator {
	return &Authenticator{
		audience:  audience,
		users:     users,
		validator: tokenValidatorFunc(idtoken.Validate),
	}
}

func (a *Authenticator) Authenticate(w http.ResponseWriter, r *http.Request) (authz.Subject, error) {
	token := r.Header.Get(header)
	if token == "" {
		// Not an IAP request.
		return nil, nil
	}
	payload, err := a.validator.Validate(r.Context(), token, a.audience)
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
	return a.users.GetOrCreateUser(r.Context(), emailString)
}
