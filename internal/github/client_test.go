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
					URL: new("https://otf-server/hooks"),
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
		BaseURL:             &internal.WebURL{URL: *u},
		SkipTLSVerification: true,
		OAuthToken:          &oauth2.Token{AccessToken: "fake-token"},
	})
	require.NoError(t, err)

	return client
}

func TestSetClientURLs(t *testing.T) {
	for _, test := range []struct {
		name          string
		url           string
		wantBaseURL   string
		wantUploadURL string
	}{
		{
			name:          "URL is the default public github URL, prefix host with api.",
			url:           "https://github.com/",
			wantBaseURL:   "https://api.github.com/",
			wantUploadURL: "https://api.github.com/",
		},
		{
			name:          "does not modify properly formed URLs",
			url:           "https://custom-url/",
			wantBaseURL:   "https://custom-url/api/v3/",
			wantUploadURL: "https://custom-url/api/uploads/",
		},
		{
			name:          "adds trailing slash",
			url:           "https://custom-url/",
			wantBaseURL:   "https://custom-url/api/v3/",
			wantUploadURL: "https://custom-url/api/uploads/",
		},
		{
			name:          "adds enterprise suffix",
			url:           "https://custom-url/",
			wantBaseURL:   "https://custom-url/api/v3/",
			wantUploadURL: "https://custom-url/api/uploads/",
		},
		{
			name:          "adds enterprise suffix and trailing slash",
			url:           "https://custom-url",
			wantBaseURL:   "https://custom-url/api/v3/",
			wantUploadURL: "https://custom-url/api/uploads/",
		},
		{
			name:          "URL has existing API prefix, adds trailing slash",
			url:           "https://api.custom-url",
			wantBaseURL:   "https://api.custom-url/",
			wantUploadURL: "https://api.custom-url/",
		},
		{
			name:          "URL has existing API prefix and trailing slash",
			url:           "https://api.custom-url/",
			wantBaseURL:   "https://api.custom-url/",
			wantUploadURL: "https://api.custom-url/",
		},
		{
			name:          "URL has API subdomain, adds trailing slash",
			url:           "https://catalog.api.custom-url",
			wantBaseURL:   "https://catalog.api.custom-url/",
			wantUploadURL: "https://catalog.api.custom-url/",
		},
		{
			name:          "URL has API subdomain and trailing slash",
			url:           "https://catalog.api.custom-url/",
			wantBaseURL:   "https://catalog.api.custom-url/",
			wantUploadURL: "https://catalog.api.custom-url/",
		},
		{
			name:          "URL is not a proper API subdomain, adds enterprise suffix and slash",
			url:           "https://cloud-api.custom-url",
			wantBaseURL:   "https://cloud-api.custom-url/api/v3/",
			wantUploadURL: "https://cloud-api.custom-url/api/uploads/",
		},
		{
			name:          "URL is not a proper API subdomain, adds enterprise suffix",
			url:           "https://cloud-api.custom-url/",
			wantBaseURL:   "https://cloud-api.custom-url/api/v3/",
			wantUploadURL: "https://cloud-api.custom-url/api/uploads/",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			baseURL, uploadURL := setClientURLs(internal.MustWebURL(test.url))
			assert.Equal(t, test.wantBaseURL, baseURL.String())
			assert.Equal(t, test.wantUploadURL, uploadURL.String())
		})
	}
}
