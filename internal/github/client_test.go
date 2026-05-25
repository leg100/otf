package github

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/google/go-github/v65/github"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/authenticator"
	"github.com/leg100/otf/internal/github/testserver"
	"github.com/leg100/otf/internal/user"
	"github.com/leg100/otf/internal/vcs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
)

func TestGetUser(t *testing.T) {
	username := user.MustUsername("bobby")
	client := newTestServerClient(t, testserver.WithUsername(username))

	got, err := client.GetCurrentUser(t.Context())
	require.NoError(t, err)

	want := authenticator.UserInfo{Username: username}
	assert.Equal(t, want, got)
}

func TestGetDefaultBranch(t *testing.T) {
	want, err := os.ReadFile("../testdata/github.tar.gz")
	require.NoError(t, err)
	client := newTestServerClient(t,
		testserver.WithRepo(vcs.NewMustRepo("acme", "terraform")),
		testserver.WithDefaultBranch("master"),
		testserver.WithArchive(want),
	)

	got, err := client.GetDefaultBranch(t.Context(), "acme/terraform")
	require.NoError(t, err)

	assert.Equal(t, "master", got)
}

func TestGetRepoTarball(t *testing.T) {
	want, err := os.ReadFile("../testdata/github.tar.gz")
	require.NoError(t, err)
	client := newTestServerClient(t,
		testserver.WithRepo(vcs.NewMustRepo("acme", "terraform")),
		testserver.WithArchive(want),
	)

	got, ref, err := client.GetRepoTarball(t.Context(), vcs.GetRepoTarballOptions{
		Repo: vcs.NewMustRepo("acme", "terraform"),
	})
	require.NoError(t, err)
	assert.Equal(t, "0335fb07bb0244b7a169ee89d15c7703e4aaf7de", ref)

	dst := t.TempDir()
	err = internal.Unpack(bytes.NewReader(got), dst)
	require.NoError(t, err)
	assert.FileExists(t, path.Join(dst, "main.tf"))
}

// TestGetRepoTarball_AnnotatedTag verifies that when GetRepoTarball is
// called with an annotated/signed tag ref, the tag is peeled to its
// underlying commit SHA before the archive endpoint is hit. Without
// peeling, Github's archive endpoint is handed a ref that points at the
// tag object rather than a commit, which fails in the wild (see #911).
func TestGetRepoTarball_AnnotatedTag(t *testing.T) {
	tarballBytes, err := os.ReadFile("../testdata/github.tar.gz")
	require.NoError(t, err)

	const (
		commitSHA = "deadbeefdeadbeefdeadbeefdeadbeefdeadbeef"
		tagObjSHA = "abc123abc123abc123abc123abc123abc123abc1"
		owner     = "acme"
		repoName  = "terraform"
		tagName   = "v1.0.0"
	)

	var capturedTarballRef string

	mux := http.NewServeMux()
	// Github responds to ListMatchingRefs for an annotated tag with a ref
	// whose object is itself a tag object (rather than a commit).
	mux.HandleFunc("/api/v3/repos/"+owner+"/"+repoName+"/git/matching-refs/tags/"+tagName, func(w http.ResponseWriter, r *http.Request) {
		refs := []*github.Reference{{
			Ref: new("refs/tags/" + tagName),
			Object: &github.GitObject{
				Type: new("tag"),
				SHA:  new(tagObjSHA),
			},
		}}
		require.NoError(t, json.NewEncoder(w).Encode(refs))
	})
	// Github exposes the underlying commit of an annotated tag via the
	// tag-object endpoint.
	mux.HandleFunc("/api/v3/repos/"+owner+"/"+repoName+"/git/tags/"+tagObjSHA, func(w http.ResponseWriter, r *http.Request) {
		tag := &github.Tag{
			SHA: new(tagObjSHA),
			Object: &github.GitObject{
				Type: new("commit"),
				SHA:  new(commitSHA),
			},
		}
		require.NoError(t, json.NewEncoder(w).Encode(tag))
	})
	// Capture whatever ref the archive endpoint is called with, then serve
	// up a real tarball so GetRepoTarball can complete.
	tarballPrefix := "/api/v3/repos/" + owner + "/" + repoName + "/tarball/"
	mux.HandleFunc(tarballPrefix, func(w http.ResponseWriter, r *http.Request) {
		capturedTarballRef = strings.TrimPrefix(r.URL.Path, tarballPrefix)
		http.Redirect(w, r, (&url.URL{Scheme: "https", Host: r.Host, Path: "/archive"}).String(), http.StatusFound)
	})
	mux.HandleFunc("/archive", func(w http.ResponseWriter, r *http.Request) {
		w.Write(tarballBytes)
	})

	server := httptest.NewTLSServer(mux)
	t.Cleanup(server.Close)

	u, err := url.Parse(server.URL)
	require.NoError(t, err)
	client, err := NewClient(ClientOptions{
		BaseURL:             &internal.WebURL{URL: *u},
		SkipTLSVerification: true,
		OAuthToken:          &oauth2.Token{AccessToken: "fake-token"},
	})
	require.NoError(t, err)

	ref := "tags/" + tagName
	_, _, err = client.GetRepoTarball(t.Context(), vcs.GetRepoTarballOptions{
		Repo: vcs.NewMustRepo(owner, repoName),
		Ref:  &ref,
	})
	require.NoError(t, err)
	assert.Equal(t, commitSHA, capturedTarballRef,
		"annotated tag must be peeled to its commit SHA before the archive endpoint is called")
}

func TestCreateWebhook(t *testing.T) {
	client := newTestServerClient(t,
		testserver.WithRepo(vcs.NewMustRepo("acme", "terraform")),
	)

	_, err := client.CreateWebhook(t.Context(), vcs.CreateWebhookOptions{
		Repo:   vcs.NewMustRepo("acme", "terraform"),
		Secret: "me-secret",
	})
	require.NoError(t, err)
}

func TestGetWebhook(t *testing.T) {
	client := newTestServerClient(t,
		testserver.WithRepo(vcs.NewMustRepo("acme", "terraform")),
		testserver.WithHook(testserver.Hook{
			Hook: &github.Hook{
				Config: &github.HookConfig{
					URL: new("https://otf-server/hooks"),
				},
			},
		}),
	)

	_, err := client.GetWebhook(t.Context(), vcs.GetWebhookOptions{
		Repo: vcs.NewMustRepo("acme", "terraform"),
		ID:   "123",
	})
	require.NoError(t, err)
}

// newTestServerClient creates a github server for testing purposes and
// returns a client configured to access the server.
func newTestServerClient(t *testing.T, opts ...testserver.TestServerOption) *Client {
	_, u := testserver.NewTestServer(t, opts...)

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
