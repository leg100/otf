package gitlab

import "github.com/leg100/otf/internal/authenticator"

func RegisterOAuthHandler(
	authenticatorService *authenticator.Service,
	hostname string,
	clientID, clientSecret string,
	skipTLSVerification bool,
) error {
	return authenticatorService.RegisterOAuthClient(authenticator.OpaqueHandlerConfig{
		ClientConstructor: NewOAuthClient,
		OAuthConfig: authenticator.OAuthConfig{
			Name:                "gitlab",
			Hostname:            hostname,
			Endpoint:            OAuthEndpoint,
			Scopes:              OAuthScopes,
			ClientID:            clientID,
			ClientSecret:        clientSecret,
			Icon:                Icon(),
			SkipTLSVerification: skipTLSVerification,
		},
	})
}
