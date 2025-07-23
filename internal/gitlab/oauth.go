package gitlab

import (
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/authenticator"
)

func RegisterOAuthHandler(
	authenticatorService *authenticator.Service,
	hostname *internal.WebURL,
	clientID, clientSecret string,
	skipTLSVerification bool,
) error {
	return authenticatorService.RegisterOAuthClient(authenticator.OpaqueHandlerConfig{
		ClientConstructor: NewOAuthClient,
		OAuthConfig: authenticator.OAuthConfig{
			Name:                "gitlab",
			BaseURL:             hostname,
			Endpoint:            OAuthEndpoint,
			Scopes:              OAuthScopes,
			ClientID:            clientID,
			ClientSecret:        clientSecret,
			Icon:                Icon(),
			SkipTLSVerification: skipTLSVerification,
		},
	})
}
