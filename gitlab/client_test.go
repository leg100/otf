package gitlab

import (
	"bytes"
	"context"
	"os"
	"path"
	"testing"

	"github.com/leg100/otf"
	"github.com/leg100/otf/cloud"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xanzy/go-gitlab"
)

func TestClient(t *testing.T) {
	ctx := context.Background()

	t.Run("GetUser", func(t *testing.T) {
		want := cloud.User{
			Name: "fake-user",
			Teams: []cloud.Team{
				{
					Name:         "maintainers",
					Organization: "fake-org",
				},
			},
			Organizations: []string{"fake-org"},
		}

		provider := newTestClient(t, WithGitlabUser(&want))

		user, err := provider.GetUser(ctx)
		require.NoError(t, err)

		assert.Equal(t, "fake-user", user.Name)
		if assert.Equal(t, 1, len(user.Organizations)) {
			assert.Equal(t, "fake-org", user.Organizations[0])
		}
		if assert.Equal(t, 1, len(user.Teams)) {
			assert.Equal(t, "maintainers", user.Teams[0].Name)
		}
	})

	t.Run("GetRepository", func(t *testing.T) {
		want := "acme/terraform"

		provider := newTestClient(t, WithGitlabRepo(want))

		got, err := provider.GetRepository(ctx, string(want))
		require.NoError(t, err)

		assert.Equal(t, want, got)
	})

	t.Run("ListRepositories", func(t *testing.T) {
		want := []string{"acme/terraform"}

		provider := newTestClient(t, WithGitlabRepo(want[0]))

		got, err := provider.ListRepositories(ctx, cloud.ListRepositoriesOptions{})
		require.NoError(t, err)

		assert.Equal(t, want, got)
	})
	t.Run("GetRepoTarball", func(t *testing.T) {
		want, err := os.ReadFile("../testdata/gitlab.tar.gz")
		require.NoError(t, err)

		client := newTestClient(t,
			WithGitlabRepo("acme/terraform"),
			WithGitlabTarball(want),
		)

		got, err := client.GetRepoTarball(ctx, cloud.GetRepoTarballOptions{
			Repo: "acme/terraform",
		})
		require.NoError(t, err)

		dst := t.TempDir()
		err = otf.Unpack(bytes.NewReader(got), dst)
		require.NoError(t, err)
		assert.FileExists(t, path.Join(dst, "afile"))
		assert.FileExists(t, path.Join(dst, "bfile"))
	})
}

func newTestClient(t *testing.T, opts ...TestGitlabServerOption) *Client {
	srv := NewTestServer(t, opts...)
	t.Cleanup(srv.Close)

	client, err := gitlab.NewOAuthClient("fake-oauth-token", gitlab.WithBaseURL(srv.URL))
	require.NoError(t, err)

	return &Client{client: client}
}
