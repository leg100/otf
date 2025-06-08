package github

import "github.com/leg100/otf/internal/authenticator"

func RegisterOAuthHandler(
	authenticatorService *authenticator.Service,
	hostname string,
	clientID, clientSecret string,
) error {
	return authenticatorService.RegisterOAuthClient(authenticator.OpaqueHandlerConfig{
		ClientConstructor: NewOAuthClient,
		OAuthConfig: authenticator.OAuthConfig{
			Name:         "Github",
			Hostname:     hostname,
			Endpoint:     OAuthEndpoint,
			Scopes:       OAuthScopes,
			ClientID:     clientID,
			ClientSecret: clientSecret,
			Icon:         Icon(),
		},
	})
}
