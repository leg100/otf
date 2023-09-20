package authenticator

import (
	"context"
	"testing"

	"github.com/leg100/otf/internal/cloud"
	"github.com/leg100/otf/internal/github"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
)

func TestOpaqueHandler_getUsername(t *testing.T) {
	want := cloud.User{Name: "fake-user"}

	_, githubURL := github.NewTestServer(t, github.WithUser(&want))

	handler := &opaqueHandler{
		OpaqueHandlerConfig: OpaqueHandlerConfig{
			Kind: cloud.GithubKind,
			OAuthConfig: OAuthConfig{
				Hostname:            githubURL,
				SkipTLSVerification: true,
			},
		},
	}

	got, err := handler.getUsername(context.Background(), &oauth2.Token{})
	require.NoError(t, err)
	assert.Equal(t, want, got)

}
