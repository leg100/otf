package gitlab

import (
	"context"
	"testing"

	"github.com/leg100/otf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xanzy/go-gitlab"
)

func TestClient(t *testing.T) {
	ctx := context.Background()

	t.Run("GetUser", func(t *testing.T) {
		org := otf.NewTestOrganization(t)
		team := otf.NewTeam("maintainers", org)
		want := otf.NewUser("fake-user", otf.WithOrganizationMemberships(org), otf.WithTeamMemberships(team))

		provider := newTestClient(t, WithGitlabUser(want))

		user, err := provider.GetUser(ctx)
		require.NoError(t, err)

		assert.Equal(t, "fake-user", user.Username())
		if assert.Equal(t, 1, len(user.Organizations())) {
			assert.Equal(t, org.Name(), user.Organizations()[0].Name())
		}
		if assert.Equal(t, 1, len(user.Teams())) {
			assert.Equal(t, "maintainers", user.Teams()[0].Name())
		}
	})

	t.Run("GetRepository", func(t *testing.T) {
		want := &otf.Repo{Identifier: "acme/terraform", Branch: "master"}

		provider := newTestClient(t, WithGitlabRepo(want))

		got, err := provider.GetRepository(ctx, want.Identifier)
		require.NoError(t, err)

		assert.Equal(t, want, got)
	})

	t.Run("ListRepositories", func(t *testing.T) {
		// TODO: test pagination - our test server doesn't currently implement it
		want := &otf.RepoList{
			Items:      []*otf.Repo{{Identifier: "acme/terraform", Branch: "master"}},
			Pagination: &otf.Pagination{},
		}

		provider := newTestClient(t, WithGitlabRepo(want.Items[0]))

		got, err := provider.ListRepositories(ctx, otf.ListOptions{})
		require.NoError(t, err)

		assert.Equal(t, want, got)
	})
	t.Run("GetRepoTarball", func(t *testing.T) {
		want := otf.NewTestTarball(t, `file1 contents`, `file2 contents`)
		client := newTestClient(t,
			WithGitlabRepo(&otf.Repo{Identifier: "acme/terraform", Branch: "master"}),
			WithGitlabTarball(want),
		)

		got, err := client.GetRepoTarball(ctx, otf.GetRepoTarballOptions{
			Identifier: "acme/terraform",
			Ref:        "master",
		})
		require.NoError(t, err)

		assert.Equal(t, want, got)
	})
}

func newTestClient(t *testing.T, opts ...TestGitlabServerOption) *Client {
	srv := NewTestServer(t, opts...)
	t.Cleanup(srv.Close)

	client, err := gitlab.NewOAuthClient("fake-oauth-token", gitlab.WithBaseURL(srv.URL))
	require.NoError(t, err)

	return &Client{client: client}
}
