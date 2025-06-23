package authenticator

import (
	"context"

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
		GetCurrentUser(ctx context.Context) (UserInfo, error)
	}
)

func (a *opaqueHandler) parseUserInfo(ctx context.Context, token *oauth2.Token) (UserInfo, error) {
	// construct client from token
	client, err := a.ClientConstructor(a.OAuthConfig, token)
	if err != nil {
		return UserInfo{}, err
	}
	// get user info from identity provider
	return client.GetCurrentUser(ctx)
}
