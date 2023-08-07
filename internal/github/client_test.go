package github

import (
	"bytes"
	"context"
	"os"
	"path"
	"testing"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/cloud"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
)

func TestGetUser(t *testing.T) {
	ctx := context.Background()
	want := cloud.User{Name: "fake-user"}
	client := newTestServerClient(t, WithUser(&want))

	got, err := client.GetCurrentUser(ctx)
	require.NoError(t, err)

	assert.Equal(t, want.Name, got.Name)
}

func TestGetRepository(t *testing.T) {
	ctx := context.Background()
	want, err := os.ReadFile("../testdata/github.tar.gz")
	require.NoError(t, err)
	client := newTestServerClient(t,
		WithRepo("acme/terraform"),
		WithDefaultBranch("master"),
		WithArchive(want),
	)

	got, err := client.GetRepository(ctx, "acme/terraform")
	require.NoError(t, err)

	assert.Equal(t, "acme/terraform", got.Path)
	assert.Equal(t, "master", got.DefaultBranch)
}

func TestGetRepoTarball(t *testing.T) {
	ctx := context.Background()
	want, err := os.ReadFile("../testdata/github.tar.gz")
	require.NoError(t, err)
	client := newTestServerClient(t,
		WithRepo("acme/terraform"),
		WithArchive(want),
	)

	got, ref, err := client.GetRepoTarball(ctx, cloud.GetRepoTarballOptions{
		Repo: "acme/terraform",
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
	_, cfg := NewTestServer(t, opts...)

	client, err := NewClient(context.Background(), cloud.ClientOptions{
		Hostname:            cfg.Hostname,
		SkipTLSVerification: true,
		Credentials: cloud.Credentials{
			OAuthToken: &oauth2.Token{AccessToken: "fake-token"},
		},
	})
	require.NoError(t, err)

	return client
}
