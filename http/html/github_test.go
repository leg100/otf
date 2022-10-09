package html

import (
	"context"
	"net/url"
	"testing"

	"github.com/leg100/otf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
)

func TestGithub_GetUser(t *testing.T) {
	ctx := context.Background()
	org := otf.NewTestOrganization(t)
	team := otf.NewTeam("fake-team", org)
	want := otf.NewUser("fake-user", otf.WithOrganizationMemberships(org), otf.WithTeamMemberships(team))
	srv := NewTestGithubServer(t, want)

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
		Token: &oauth2.Token{AccessToken: "fake-token"},
	})
	require.NoError(t, err)

	got, err := client.GetUser(ctx)
	require.NoError(t, err)

	assert.Equal(t, want.Username(), got.Username())
	if assert.Equal(t, 1, len(got.Organizations)) {
		assert.Equal(t, org.Name(), got.Organizations[0].Name())
	}
	if assert.Equal(t, 2, len(got.Teams)) {
		assert.Equal(t, "owners", got.Teams[0].Name())
		assert.Equal(t, team.Name(), got.Teams[1].Name())
	}
}
