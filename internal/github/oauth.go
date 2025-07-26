package github

import (
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/authenticator"

	oauth2github "golang.org/x/oauth2/github"
)

var (
	OAuthEndpoint = oauth2github.Endpoint
	// TODO: don't think read:org scope is necessary any more...not since OTF
	// stopped sync'ing org and team memberships from github.
	OAuthScopes = []string{"user:email", "read:org"}
)

func RegisterOAuthHandler(
	authenticatorService *authenticator.Service,
	baseURL *internal.WebURL,
	clientID, clientSecret string,
	skipTLSVerification bool,
) error {
	return authenticatorService.RegisterOAuthClient(authenticator.OpaqueHandlerConfig{
		ClientConstructor: NewOAuthClient,
		OAuthConfig: authenticator.OAuthConfig{
			Name:                "github",
			BaseURL:             baseURL,
			Endpoint:            OAuthEndpoint,
			Scopes:              OAuthScopes,
			ClientID:            clientID,
			ClientSecret:        clientSecret,
			Icon:                Icon(),
			SkipTLSVerification: skipTLSVerification,
		},
	})
}
