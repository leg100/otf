package github

import (
	"bytes"
	"context"
	"os"
	"path"
	"testing"

	"github.com/google/go-github/v65/github"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/authenticator"
	"github.com/leg100/otf/internal/user"
	"github.com/leg100/otf/internal/vcs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
)

func TestGetUser(t *testing.T) {
	ctx := context.Background()
	username := user.MustUsername("bobby")
	client := newTestServerClient(t, WithUsername(username))

	got, err := client.GetCurrentUser(ctx)
	require.NoError(t, err)

	want := authenticator.UserInfo{Username: username}
	assert.Equal(t, want, got)
}

func TestGetDefaultBranch(t *testing.T) {
	ctx := context.Background()
	want, err := os.ReadFile("../testdata/github.tar.gz")
	require.NoError(t, err)
	client := newTestServerClient(t,
		WithRepo(vcs.NewMustRepo("acme", "terraform")),
		WithDefaultBranch("master"),
		WithArchive(want),
	)

	got, err := client.GetDefaultBranch(ctx, "acme/terraform")
	require.NoError(t, err)

	assert.Equal(t, "master", got)
}

func TestGetRepoTarball(t *testing.T) {
	ctx := context.Background()
	want, err := os.ReadFile("../testdata/github.tar.gz")
	require.NoError(t, err)
	client := newTestServerClient(t,
		WithRepo(vcs.NewMustRepo("acme", "terraform")),
		WithArchive(want),
	)

	got, ref, err := client.GetRepoTarball(ctx, vcs.GetRepoTarballOptions{
		Repo: vcs.NewMustRepo("acme", "terraform"),
	})
	require.NoError(t, err)
	assert.Equal(t, "0335fb07bb0244b7a169ee89d15c7703e4aaf7de", ref)

	dst := t.TempDir()
	err = internal.Unpack(bytes.NewReader(got), dst)
	require.NoError(t, err)
	assert.FileExists(t, path.Join(dst, "main.tf"))
}

func TestCreateWebhook(t *testing.T) {
	ctx := context.Background()

	client := newTestServerClient(t,
		WithRepo(vcs.NewMustRepo("acme", "terraform")),
	)

	_, err := client.CreateWebhook(ctx, vcs.CreateWebhookOptions{
		Repo:   vcs.NewMustRepo("acme", "terraform"),
		Secret: "me-secret",
	})
	require.NoError(t, err)
}

func TestGetWebhook(t *testing.T) {
	ctx := context.Background()

	client := newTestServerClient(t,
		WithRepo(vcs.NewMustRepo("acme", "terraform")),
		WithHook(hook{
			Hook: &github.Hook{
				Config: &github.HookConfig{
					URL: internal.Ptr("https://otf-server/hooks"),
				},
			},
		}),
	)

	_, err := client.GetWebhook(ctx, vcs.GetWebhookOptions{
		Repo: vcs.NewMustRepo("acme", "terraform"),
		ID:   "123",
	})
	require.NoError(t, err)
}

// newTestServerClient creates a github server for testing purposes and
// returns a client configured to access the server.
func newTestServerClient(t *testing.T, opts ...TestServerOption) *Client {
	_, u := NewTestServer(t, opts...)

	client, err := NewClient(ClientOptions{
		BaseURL:             &internal.WebURL{URL: u},
		SkipTLSVerification: true,
		OAuthToken:          &oauth2.Token{AccessToken: "fake-token"},
	})
	require.NoError(t, err)

	return client
}
