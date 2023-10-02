package gitlab

import (
	"bytes"
	"context"
	"os"
	"path"
	"testing"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/cloud"
	"github.com/leg100/otf/internal/vcs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xanzy/go-gitlab"
)

func TestClient(t *testing.T) {
	ctx := context.Background()

	t.Run("GetUser", func(t *testing.T) {
		want := cloud.User{Name: "fake-user"}

		provider := newTestClient(t, WithGitlabUser(&want))

		user, err := provider.GetCurrentUser(ctx)
		require.NoError(t, err)

		assert.Equal(t, "fake-user", user.Name)
	})

	t.Run("GetRepository", func(t *testing.T) {
		provider := newTestClient(t, WithGitlabRepo("acme/terraform", "master"))

		got, err := provider.GetRepository(ctx, "acme/terraform")
		require.NoError(t, err)

		assert.Equal(t, "acme/terraform", got.Path)
		assert.Equal(t, "master", got.DefaultBranch)
	})

	t.Run("ListRepositories", func(t *testing.T) {
		want := []string{"acme/terraform"}

		provider := newTestClient(t, WithGitlabRepo(want[0], ""))

		got, err := provider.ListRepositories(ctx, vcs.ListRepositoriesOptions{})
		require.NoError(t, err)

		assert.Equal(t, want, got)
	})
	t.Run("GetRepoTarball", func(t *testing.T) {
		want, err := os.ReadFile("../testdata/gitlab.tar.gz")
		require.NoError(t, err)

		client := newTestClient(t,
			WithGitlabRepo("acme/terraform", ""),
			WithGitlabTarball(want),
		)

		got, ref, err := client.GetRepoTarball(ctx, vcs.GetRepoTarballOptions{
			Repo: "acme/terraform",
		})
		require.NoError(t, err)
		assert.Equal(t, "0335fb07bb0244b7a169ee89d15c7703e4aaf7de", ref)

		dst := t.TempDir()
		err = internal.Unpack(bytes.NewReader(got), dst)
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
