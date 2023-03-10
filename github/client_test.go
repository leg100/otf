package github

import (
	"bytes"
	"context"
	"net/url"
	"os"
	"path"
	"testing"

	"github.com/leg100/otf"
	"github.com/leg100/otf/cloud"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
)

func TestGetUser(t *testing.T) {
	ctx := context.Background()
	want := cloud.User{
		Name: "fake-user",
		Teams: []cloud.Team{
			{
				Name:         "fake-team",
				Organization: "fake-org",
			},
		},
		Organizations: []string{"fake-org"},
	}
	client := newTestServerClient(t, WithUser(&want))

	got, err := client.GetUser(ctx)
	require.NoError(t, err)

	assert.Equal(t, want.Name, got.Name)
	if assert.Equal(t, 1, len(got.Organizations)) {
		assert.Equal(t, "fake-org", got.Organizations[0])
	}
	if assert.Equal(t, 1, len(got.Teams)) {
		assert.Equal(t, "fake-team", got.Teams[0].Name)
	}
}

func TestGetRepoTarball(t *testing.T) {
	ctx := context.Background()
	want, err := os.ReadFile("../testdata/github.tar.gz")
	require.NoError(t, err)
	client := newTestServerClient(t,
		WithRepo("acme/terraform"),
		WithArchive(want),
	)

	got, err := client.GetRepoTarball(ctx, cloud.GetRepoTarballOptions{
		Repo: "acme/terraform",
	})
	require.NoError(t, err)

	dst := t.TempDir()
	err = otf.Unpack(bytes.NewReader(got), dst)
	require.NoError(t, err)
	assert.FileExists(t, path.Join(dst, "main.tf"))
}

func TestCreateWebhook(t *testing.T) {
	ctx := context.Background()

	client := newTestServerClient(t,
		WithRepo("acme/terraform"),
	)

	_, err := client.CreateWebhook(ctx, cloud.CreateWebhookOptions{
		Repo:   "acme/terraform",
		Secret: "me-secret",
	})
	require.NoError(t, err)
}

// newTestServerClient creates a github server for testing purposes and
// returns a client configured to access the server.
func newTestServerClient(t *testing.T, opts ...TestServerOption) *Client {
	srv := NewTestServer(t, opts...)

	u, err := url.Parse(srv.URL)
	require.NoError(t, err)

	client, err := NewClient(context.Background(), cloud.ClientOptions{
		Hostname:            u.Host,
		SkipTLSVerification: true,
		Credentials: cloud.Credentials{
			OAuthToken: &oauth2.Token{AccessToken: "fake-token"},
		},
	})
	require.NoError(t, err)

	return client
}
