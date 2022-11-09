package otf

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xanzy/go-gitlab"
)

func TestGitlab_GetUser(t *testing.T) {
	ctx := context.Background()

	t.Run("GetUser", func(t *testing.T) {
		org := NewTestOrganization(t)
		team := NewTeam("maintainers", org)
		want := NewUser("fake-user", WithOrganizationMemberships(org), WithTeamMemberships(team))

		provider := newTestGitlabClient(t, WithGitlabUser(want))

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
		want := &Repo{Identifier: "acme/terraform", Branch: "master"}

		provider := newTestGitlabClient(t, WithGitlabRepo(want))

		got, err := provider.GetRepository(ctx, want.Identifier)
		require.NoError(t, err)

		assert.Equal(t, want, got)
	})

	t.Run("ListRepositories", func(t *testing.T) {
		// TODO: test pagination - our test server doesn't currently implement it
		want := &RepoList{
			Items:      []*Repo{{Identifier: "acme/terraform", Branch: "master"}},
			Pagination: &Pagination{},
		}

		provider := newTestGitlabClient(t, WithGitlabRepo(want.Items[0]))

		got, err := provider.ListRepositories(ctx, ListOptions{})
		require.NoError(t, err)

		assert.Equal(t, want, got)
	})
	t.Run("GetRepoTarball", func(t *testing.T) {
		want := NewTestTarball(t, `file1 contents`, `file2 contents`)
		provider := newTestGitlabClient(t,
			WithGitlabRepo(&Repo{Identifier: "acme/terraform", Branch: "master"}),
			WithGitlabTarball(want),
		)

		got, err := provider.GetRepoTarball(ctx, &VCSRepo{
			Identifier: "acme/terraform",
			Branch:     "master",
		})
		require.NoError(t, err)

		assert.Equal(t, want, got)
	})
}

func newTestGitlabClient(t *testing.T, opts ...TestGitlabServerOption) *gitlabProvider {
	srv := NewTestGitlabServer(t, opts...)
	t.Cleanup(srv.Close)

	client, err := gitlab.NewOAuthClient("fake-oauth-token", gitlab.WithBaseURL(srv.URL))
	require.NoError(t, err)

	return &gitlabProvider{client: client}
}
