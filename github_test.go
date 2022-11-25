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

// TODO: refactor tests below, create helpers for creating server and client

func TestGithub_GetUser(t *testing.T) {
	ctx := context.Background()
	org := NewTestOrganization(t)
	team := NewTeam("fake-team", org)
	want := NewUser("fake-user", WithOrganizationMemberships(org), WithTeamMemberships(team))
	srv := NewTestGithubServer(t, WithGithubUser(want))

	u, err := url.Parse(srv.URL)
	require.NoError(t, err)

	client, err := NewGithubClient(ctx, ClientConfig{
		Hostname:            u.Host,
		SkipTLSVerification: true,
		OAuthToken:          &oauth2.Token{AccessToken: "fake-token"},
	})
	require.NoError(t, err)

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

	srv := NewTestGithubServer(t,
		WithGithubRepo(&Repo{Identifier: "acme/terraform", Branch: "master"}),
		WithGithubArchive(want),
	)
	u, err := url.Parse(srv.URL)
	require.NoError(t, err)

	client, err := NewGithubClient(ctx, ClientConfig{
		Hostname:            u.Host,
		SkipTLSVerification: true,
		OAuthToken:          &oauth2.Token{AccessToken: "fake-token"},
	})
	require.NoError(t, err)

	got, err := client.GetRepoTarball(ctx, &VCSRepo{
		Identifier: "acme/terraform",
		Branch:     "master",
	})
	require.NoError(t, err)

	dst := t.TempDir()
	err = Unpack(bytes.NewReader(got), dst)
	require.NoError(t, err)
}

func TestGithub_CreateWebhook(t *testing.T) {
	ctx := context.Background()

	srv := NewTestGithubServer(t,
		WithGithubRepo(&Repo{Identifier: "acme/terraform", Branch: "master"}),
	)
	u, err := url.Parse(srv.URL)
	require.NoError(t, err)

	client, err := NewGithubClient(ctx, ClientConfig{
		Hostname:            u.Host,
		SkipTLSVerification: true,
		OAuthToken:          &oauth2.Token{AccessToken: "fake-token"},
	})
	require.NoError(t, err)

	err = client.CreateWebhook(ctx, CreateWebhookOptions{
		Identifier: "acme/terraform",
		Host:       "https://me-server/me-webhook",
		Secret:     "me-secret",
	})
	require.NoError(t, err)
}
