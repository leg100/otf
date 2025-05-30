package authenticator

import (
	"context"

	"github.com/leg100/otf/internal/user"
	"golang.org/x/oauth2"
)

type (
	// opaqueHandler uses an 'opaque' OAuth access token to retrieve the
	// username of the authenticated user.
	opaqueHandler struct {
		OpaqueHandlerConfig
	}

	OpaqueHandlerConfig struct {
		OAuthConfig
		ClientConstructor func(cfg OAuthConfig, token *oauth2.Token) (IdentityProviderClient, error)
	}

	IdentityProviderClient interface {
		// GetCurrentUser retrieves the currently authenticated user
		GetCurrentUser(ctx context.Context) (user.Username, error)
	}
)

func (a *opaqueHandler) getUsername(ctx context.Context, token *oauth2.Token) (user.Username, error) {
	// construct client from token
	client, err := a.ClientConstructor(a.OAuthConfig, token)
	if err != nil {
		return user.Username{}, err
	}
	// get username from identity provider
	return client.GetCurrentUser(ctx)
}
