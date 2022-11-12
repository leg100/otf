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
	srv := NewTestGithubServer(t, WithGithubUser(want))

	u, err := url.Parse(srv.URL)
	require.NoError(t, err)

	cloud := &GithubCloud{
		&GithubConfig{
			cloudConfig{
				hostname:            u.Host,
				skipTLSVerification: true,
			},
		},
	}
	client, err := cloud.NewDirectoryClient(ctx, DirectoryClientOptions{
		OAuthToken: &oauth2.Token{AccessToken: "fake-token"},
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

	cloud := &GithubCloud{
		&GithubConfig{
			cloudConfig{
				hostname:            u.Host,
				skipTLSVerification: true,
			},
		},
	}
	client, err := cloud.NewDirectoryClient(ctx, DirectoryClientOptions{
		OAuthToken: &oauth2.Token{AccessToken: "fake-token"},
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
