package gitlab

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"path"
	"testing"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/authenticator"
	"github.com/leg100/otf/internal/testutils"
	"github.com/leg100/otf/internal/user"
	"github.com/leg100/otf/internal/vcs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_GetCurrentUser(t *testing.T) {
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/user", func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "GET", r.Method)
		fmt.Fprint(w, `{"username":"bobby","avatar_url": "https://mymugshot.com"}`)
	})

	got, err := client.GetCurrentUser(context.Background())
	require.NoError(t, err)

	want := authenticator.UserInfo{
		Username:  user.MustUsername("bobby"),
		AvatarURL: new("https://mymugshot.com"),
	}
	assert.Equal(t, want, got)
}

func TestClient_GetDefaultBranch(t *testing.T) {
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/acme%2Fterraform", func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "GET", r.Method)
		fmt.Fprint(w, `{"path_with_namespace":"acme/terraform","default_branch":"master"}`)
	})

	got, err := client.GetDefaultBranch(context.Background(), "acme/terraform")
	require.NoError(t, err)

	assert.Equal(t, "master", got)
}

func TestClient_ListRepositories(t *testing.T) {
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects", func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "GET", r.Method)
		fmt.Fprint(w, `[{"namespace": {"path": "acme"}, "path":"terraform"}]`)
	})

	got, err := client.ListRepositories(context.Background(), vcs.ListRepositoriesOptions{})
	require.NoError(t, err)

	want := vcs.NewMustRepo("acme", "terraform")
	assert.Equal(t, []vcs.Repo{want}, got)
}

func TestClient_GetRepoTarball(t *testing.T) {
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/acme%2Fterraform/repository/archive.tar.gz", func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "GET", r.Method)
		w.Write(testutils.ReadFile(t, "../testdata/gitlab.tar.gz"))
	})

	got, ref, err := client.GetRepoTarball(context.Background(), vcs.GetRepoTarballOptions{
		Repo: vcs.NewMustRepo("acme", "terraform"),
	})
	require.NoError(t, err)
	assert.Equal(t, "0335fb07bb0244b7a169ee89d15c7703e4aaf7de", ref)

	dst := t.TempDir()
	err = internal.Unpack(bytes.NewReader(got), dst)
	require.NoError(t, err)
	assert.FileExists(t, path.Join(dst, "afile"))
	assert.FileExists(t, path.Join(dst, "bfile"))
}

func TestClient_CreateWebhook(t *testing.T) {
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/acme%2Fterraform/hooks", func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "POST", r.Method)
		fmt.Fprint(w, `{"id":1}`)
	})

	got, err := client.CreateWebhook(context.Background(), vcs.CreateWebhookOptions{
		Repo: vcs.NewMustRepo("acme", "terraform"),
	})
	require.NoError(t, err)
	assert.Equal(t, "1", got)
}

func TestClient_UpdateWebhook(t *testing.T) {
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/acme%2Fterraform/hooks/1", func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "PUT", r.Method)
		fmt.Fprint(w, `{"id":1}`)
	})

	err := client.UpdateWebhook(context.Background(), "1", vcs.UpdateWebhookOptions{
		Repo: vcs.NewMustRepo("acme", "terraform"),
	})
	require.NoError(t, err)
}

func TestClient_GetWebhook(t *testing.T) {
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/acme%2Fterraform/hooks/1", func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "GET", r.Method)
		fmt.Fprint(w, `{"id":1}`)
	})

	_, err := client.GetWebhook(context.Background(), vcs.GetWebhookOptions{
		ID:   "1",
		Repo: vcs.NewMustRepo("acme", "terraform"),
	})
	require.NoError(t, err)
}

func TestClient_DeleteWebhook(t *testing.T) {
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/acme%2Fterraform/hooks/1", func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "DELETE", r.Method)
		fmt.Fprint(w, `{"id":1}`)
	})

	err := client.DeleteWebhook(context.Background(), vcs.DeleteWebhookOptions{
		ID:   "1",
		Repo: vcs.NewMustRepo("acme", "terraform"),
	})
	require.NoError(t, err)
}

func TestClient_ListPullRequestFiles(t *testing.T) {
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/acme%2Fterraform/merge_requests/1/diffs", func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "GET", r.Method)
		fmt.Fprint(w, `[{"old_path":"main.tf","new_path":"main.tf"},{"old_path":"dev.tf","new_path":"prod.tf"}]`)
	})

	repo := vcs.NewMustRepo("acme", "terraform")
	got, err := client.ListPullRequestFiles(context.Background(), repo, 1)
	require.NoError(t, err)
	assert.Equal(t, []string{"dev.tf", "main.tf", "prod.tf"}, got)
}

func TestClient_GetCommit(t *testing.T) {
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/acme%2Fterraform/repository/commits/abc123", func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "GET", r.Method)
		fmt.Fprint(w, `{"id":"abc123","web_url":"https://gitlab.com/commits/abc123"}`)
	})

	repo := vcs.NewMustRepo("acme", "terraform")
	got, err := client.GetCommit(context.Background(), repo, "abc123")
	require.NoError(t, err)
	want := vcs.Commit{
		SHA: "abc123",
		URL: "https://gitlab.com/commits/abc123",
	}
	assert.Equal(t, want, got)
}
