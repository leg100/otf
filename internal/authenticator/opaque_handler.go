package authenticator

import (
	"context"
	"fmt"

	"github.com/leg100/otf/internal/cloud"
	"github.com/leg100/otf/internal/github"
	"github.com/leg100/otf/internal/gitlab"
	"golang.org/x/oauth2"
)

type (
	// opaqueHandler uses an 'opaque' OAuth access token to retrieve the
	// username of the authenticated user.
	opaqueHandler struct {
		OpaqueHandlerConfig
	}

	OpaqueHandlerConfig struct {
		cloud.Kind
		OAuthConfig
	}

	identityProviderClient interface {
		// GetCurrentUser retrieves the currently authenticated user
		GetCurrentUser(ctx context.Context) (cloud.User, error)
	}
)

func (a *opaqueHandler) getUsername(ctx context.Context, token *oauth2.Token) (string, error) {
	client, err := a.newCloudClient(ctx, token)
	if err != nil {
		return "", err
	}
	// Get identity provider user
	cuser, err := client.GetCurrentUser(ctx)
	if err != nil {
		return "", err
	}
	return cuser.Name, nil
}

func (a *opaqueHandler) newCloudClient(ctx context.Context, token *oauth2.Token) (identityProviderClient, error) {
	switch a.Kind {
	case cloud.GithubKind:
		return github.NewClient(ctx, github.ClientOptions{
			Hostname:            a.Hostname,
			SkipTLSVerification: a.SkipTLSVerification,
			OAuthToken:          token,
		})
	case cloud.GitlabKind:
		return gitlab.NewClient(ctx, gitlab.ClientOptions{OAuthToken: token})
	default:
		return nil, fmt.Errorf("unknown cloud kind: %s", a.Kind)
	}
}
