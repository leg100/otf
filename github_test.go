package otf

import (
	"bytes"
	"context"
	"net/url"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
)

func TestGithub_GetUser(t *testing.T) {
	ctx := context.Background()
	org := NewTestOrganization(t)
	team := NewTeam("fake-team", org)
	want := NewUser("fake-user", WithOrganizationMemberships(org), WithTeamMemberships(team))
	client := newTestGithubServerClient(t, WithGithubUser(want))

	got, err := client.GetUser(ctx)
	require.NoError(t, err)

	assert.Equal(t, want.Username(), got.Username())
	if assert.Equal(t, 1, len(got.Organizations())) {
		assert.Equal(t, org.Name(), got.Organizations()[0].Name())
	}
	if assert.Equal(t, 1, len(got.Teams())) {
		assert.Equal(t, team.Name(), got.Teams()[0].Name())
	}
}

func TestGithub_GetRepoTarball(t *testing.T) {
	ctx := context.Background()
	want, err := os.ReadFile("testdata/github.tar.gz")
	require.NoError(t, err)
	client := newTestGithubServerClient(t,
		WithGithubRepo(&Repo{Identifier: "acme/terraform", Branch: "master"}),
		WithGithubArchive(want),
	)

	got, err := client.GetRepoTarball(ctx, GetRepoTarballOptions{
		Identifier: "acme/terraform",
		Ref:        "master",
	})
	require.NoError(t, err)

	dst := t.TempDir()
	err = Unpack(bytes.NewReader(got), dst)
	require.NoError(t, err)
}

func TestGithub_CreateWebhook(t *testing.T) {
	ctx := context.Background()

	client := newTestGithubServerClient(t,
		WithGithubRepo(&Repo{Identifier: "acme/terraform", Branch: "master"}),
	)

	_, err := client.CreateWebhook(ctx, CreateWebhookOptions{
		Identifier: "acme/terraform",
		Secret:     "me-secret",
	})
	require.NoError(t, err)
}

// newTestGithubServerClient creates a github server for testing purposes and
// returns a client configured to access the server.
func newTestGithubServerClient(t *testing.T, opts ...TestGithubServerOption) *GithubClient {
	srv := NewTestGithubServer(t, opts...)

	u, err := url.Parse(srv.URL)
	require.NoError(t, err)

	client, err := NewGithubClient(context.Background(), CloudClientOptions{
		Hostname:            u.Host,
		SkipTLSVerification: true,
		CloudCredentials: CloudCredentials{
			OAuthToken: &oauth2.Token{AccessToken: "fake-token"},
		},
	})
	require.NoError(t, err)

	return client
}
