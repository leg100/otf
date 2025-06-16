package forgejo

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path"
	"testing"

	"codeberg.org/mvdkleijn/forgejo-sdk/forgejo/v2"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/user"
	"github.com/leg100/otf/internal/vcs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetDefaultBranch(t *testing.T) {
	ctx := context.Background()
	want, err := os.ReadFile("../testdata/forgejo.tar.gz")
	require.NoError(t, err)
	testuser, err := user.NewUsername("user")
	require.NoError(t, err)
	client := newTestServerClient(t,
		WithRepo(vcs.NewMustRepo("acme", "test")),
		WithDefaultBranch("main"),
		WithUsername(testuser),
		WithCommit("000111222333444555666777888999aaabbbcccd"),
		WithArchive(want),
	)

	got, err := client.GetDefaultBranch(ctx, "acme/test")
	require.NoError(t, err)

	assert.Equal(t, "main", got)

	got, err = client.GetDefaultBranch(ctx, "acme/nonexistent-repo")
	require.Error(t, err)
	require.Zero(t, got)

	got, err = client.GetDefaultBranch(ctx, "nonexistent-org/test")
	require.Error(t, err)
	require.Zero(t, got)
}

func TestListRepositories(t *testing.T) {
	ctx := context.Background()
	want, err := os.ReadFile("../testdata/forgejo.tar.gz")
	require.NoError(t, err)
	testuser, err := user.NewUsername("user")
	require.NoError(t, err)
	client := newTestServerClient(t,
		WithRepo(vcs.NewMustRepo("acme", "test")),
		WithDefaultBranch("main"),
		WithUsername(testuser),
		WithCommit("000111222333444555666777888999aaabbbcccd"),
		WithArchive(want),
	)

	got, err := client.ListRepositories(ctx, vcs.ListRepositoriesOptions{})
	require.NoError(t, err)

	assert.Equal(t, []vcs.Repo{vcs.NewMustRepo("acme", "test")}, got)
}

func TestGetRepoTarball(t *testing.T) {
	ctx := context.Background()
	want, err := os.ReadFile("../testdata/forgejo.tar.gz")
	require.NoError(t, err)
	client := newTestServerClient(t,
		WithRepo(vcs.NewMustRepo("acme", "test")),
		WithDefaultBranch("main"),
		WithCommit("000111222333444555666777888999aaabbbcccd"),
		WithArchive(want),
	)

	got, gotref, err := client.GetRepoTarball(ctx, vcs.GetRepoTarballOptions{
		Repo: vcs.NewMustRepo("acme", "test"),
	})
	require.NoError(t, err)
	// archive returned and non-empty
	assert.True(t, len(got) > 0)

	// unpack and compare contents of file within archive
	wanttmp := t.TempDir()
	gottmp := t.TempDir()
	err = internal.Unpack(bytes.NewReader(want), wanttmp)
	require.NoError(t, err)
	err = internal.Unpack(bytes.NewReader(got), gottmp)
	require.NoError(t, err)
	wantfn := path.Join(wanttmp, "test", "main.tf") // VCS archive has a "test/" prefix
	gotfn := path.Join(gottmp, "main.tf")           // client.GetArchive stripped that prefix
	assert.FileExists(t, wantfn)
	assert.FileExists(t, gotfn)
	wanttf, err := os.ReadFile(wantfn)
	require.NoError(t, err)
	gottf, err := os.ReadFile(gotfn)
	require.NoError(t, err)
	assert.Equal(t, wanttf, gottf)

	assert.True(t, len(got) > 0)
	assert.Equal(t, "main", gotref)

	got, gotref, err = client.GetRepoTarball(ctx, vcs.GetRepoTarballOptions{
		Repo: vcs.NewMustRepo("acme", "nonexistent-repo"),
	})
	require.Error(t, err)
	require.Zero(t, got)
	require.Zero(t, gotref)

	got, gotref, err = client.GetRepoTarball(ctx, vcs.GetRepoTarballOptions{
		Repo: vcs.NewMustRepo("nonexisting-org", "test"),
	})
	require.Error(t, err)
	require.Zero(t, got)
	require.Zero(t, gotref)
}

func TestListPullRequestFiles(t *testing.T) {
	ctx := context.Background()
	client := newTestServerClient(t,
		WithRepo(vcs.NewMustRepo("acme", "test")),
		WithDefaultBranch("main"),
		WithCommit("000111222333444555666777888999aaabbbcccd"),
		WithPullRequest("432", "foo", "bar"),
	)

	files, err := client.ListPullRequestFiles(ctx, vcs.NewMustRepo("acme", "test"), 432)
	require.NoError(t, err)
	// archive returned and non-empty
	assert.Equal(t, 2, len(files))
	assert.Contains(t, files, "foo")
	assert.Contains(t, files, "bar")

	_, err = client.ListPullRequestFiles(ctx, vcs.NewMustRepo("acme", "test"), 234)
	require.Error(t, err)
}

func TestGetCommit(t *testing.T) {
	sha := "000111222333444555666777888999aaabbbcccd"
	ctx := context.Background()
	testuser, err := user.NewUsername("user")
	require.NoError(t, err)
	client := newTestServerClient(t,
		WithRepo(vcs.NewMustRepo("acme", "test")),
		WithDefaultBranch("main"),
		WithUsername(testuser),
		WithCommit(sha),
		WithRefs(testref{ref: "main", object: sha}),
	)

	// empty refname returns HEAD of default branch
	got, err := client.GetCommit(ctx, vcs.NewMustRepo("acme", "test"), "")
	require.NoError(t, err)

	assert.Equal(t, sha, got.SHA)
	assert.Equal(t, "user", got.Author.Username)

	// branch name returns HEAD of that branch
	got, err = client.GetCommit(ctx, vcs.NewMustRepo("acme", "test"), "main")
	require.NoError(t, err)

	assert.Equal(t, sha, got.SHA)
	assert.Equal(t, "user", got.Author.Username)

	// fully qualified branch name returns the HEAD of that branch
	got, err = client.GetCommit(ctx, vcs.NewMustRepo("acme", "test"), "refs/heads/main")
	require.NoError(t, err)

	assert.Equal(t, sha, got.SHA)
	assert.Equal(t, "user", got.Author.Username)

	// tag name returns the tag it points to
	got, err = client.GetCommit(ctx, vcs.NewMustRepo("acme", "test"), "v0.0.1")
	require.NoError(t, err)

	assert.Equal(t, sha, got.SHA)
	assert.Equal(t, "user", got.Author.Username)

	got, err = client.GetCommit(ctx, vcs.NewMustRepo("acme", "nonexistent-repo"), "")
	require.Error(t, err)
	require.Zero(t, got)

	got, err = client.GetCommit(ctx, vcs.NewMustRepo("nonexistent-org", "test"), "")
	require.Error(t, err)
	require.Zero(t, got)
}

func TestCreateWebhook(t *testing.T) {
	ctx := context.Background()

	client := newTestServerClient(t,
		WithRepo(vcs.NewMustRepo("acme", "test")),
	)

	hookid, err := client.CreateWebhook(ctx, vcs.CreateWebhookOptions{
		Repo:   vcs.NewMustRepo("acme", "test"),
		Secret: "so-sneaky",
	})
	require.NoError(t, err)
	assert.Equal(t, fmt.Sprintf("%x", 123), hookid)
}

func TestGetWebhook(t *testing.T) {
	ctx := context.Background()

	client := newTestServerClient(t,
		WithRepo(vcs.NewMustRepo("acme", "test")),
		WithHook(hook{
			Hook: &forgejo.Hook{
				Config: map[string]string{
					"url": "https://otf-server/hooks",
				},
				Events: []string{
					"push",
					"pull_request",
				},
			},
		}),
	)

	hook, err := client.GetWebhook(ctx, vcs.GetWebhookOptions{
		Repo: vcs.NewMustRepo("acme", "test"),
		ID:   fmt.Sprintf("%x", 123),
	})
	require.NoError(t, err)
	require.Contains(t, hook.Events, vcs.EventTypePush)
	require.Contains(t, hook.Events, vcs.EventTypePull)
}

func TestUpdateWebhook(t *testing.T) {
	ctx := context.Background()

	client := newTestServerClient(t,
		WithRepo(vcs.NewMustRepo("acme", "test")),
		WithHook(hook{
			Hook: &forgejo.Hook{
				Config: map[string]string{
					"url": "https://otf-server/hooks",
				},
				Events: []string{
					"push",
					"pull_request",
				},
			},
		}),
	)
	hookid := fmt.Sprintf("%x", 123)

	err := client.UpdateWebhook(ctx, hookid, vcs.UpdateWebhookOptions{
		Repo:   vcs.NewMustRepo("acme", "test"),
		Secret: "so-sneaky",
		Events: []vcs.EventType{vcs.EventTypePush},
	})
	require.NoError(t, err)

	hook, err := client.GetWebhook(ctx, vcs.GetWebhookOptions{
		Repo: vcs.NewMustRepo("acme", "test"),
		ID:   fmt.Sprintf("%x", 123),
	})
	require.NoError(t, err)
	require.Contains(t, hook.Events, vcs.EventTypePush)
	require.NotContains(t, hook.Events, vcs.EventTypePull)
}

func TestDeleteWebhook(t *testing.T) {
	ctx := context.Background()

	client := newTestServerClient(t,
		WithRepo(vcs.NewMustRepo("acme", "test")),
		WithHook(hook{
			Hook: &forgejo.Hook{
				Config: map[string]string{
					"url": "https://otf-server/hooks",
				},
			},
		}),
	)
	hookid := fmt.Sprintf("%x", 123)

	_, err := client.GetWebhook(ctx, vcs.GetWebhookOptions{
		Repo: vcs.NewMustRepo("acme", "test"),
		ID:   hookid,
	})
	require.NoError(t, err)
	err = client.DeleteWebhook(ctx, vcs.DeleteWebhookOptions{
		Repo: vcs.NewMustRepo("acme", "test"),
		ID:   hookid,
	})
	require.NoError(t, err)
	_, err = client.GetWebhook(ctx, vcs.GetWebhookOptions{
		Repo: vcs.NewMustRepo("acme", "test"),
		ID:   hookid,
	})
	require.Error(t, err)
}

// func (c *Client) ListTags(ctx context.Context, opts vcs.ListTagsOptions) ([]string, error)

func TestListTags(t *testing.T) {
	ctx := context.Background()
	sha := "000111222333444555666777888999aaabbbcccd"
	client := newTestServerClient(t,
		WithRepo(vcs.NewMustRepo("acme", "test")),
		WithDefaultBranch("main"),
		WithCommit(sha),
		WithRefs(
			testref{
				ref:    "refs/heads/main",
				object: sha,
			},
			testref{
				ref:    "refs/tags/v0.0.1",
				object: sha,
			},
		),
	)

	// no prefix, return all tags
	tags, err := client.ListTags(ctx, vcs.ListTagsOptions{
		Repo: vcs.NewMustRepo("acme", "test"),
	})
	require.NoError(t, err)
	assert.Equal(t, []string{"tags/v0.0.1"}, tags)

	// matching prefix, still return the tag
	tags, err = client.ListTags(ctx, vcs.ListTagsOptions{
		Repo:   vcs.NewMustRepo("acme", "test"),
		Prefix: "v",
	})
	require.NoError(t, err)
	assert.Equal(t, []string{"tags/v0.0.1"}, tags)

	// non-matching prefix, don't return the tag
	tags, err = client.ListTags(ctx, vcs.ListTagsOptions{
		Repo:   vcs.NewMustRepo("acme", "test"),
		Prefix: "q",
	})
	require.NoError(t, err)
	assert.Equal(t, []string{}, tags)
}

func TestSetStatus(t *testing.T) {
	ctx := context.Background()
	sha := "000111222333444555666777888999aaabbbcccd"
	url := "https://foo.bar.com/1234"
	testuser, err := user.NewUsername("user")
	require.NoError(t, err)
	client, server := newTestServerClientPair(t,
		WithRepo(vcs.NewMustRepo("acme", "test")),
		WithDefaultBranch("main"),
		WithUsername(testuser),
		WithCommit(sha),
	)

	err = client.SetStatus(ctx, vcs.SetStatusOptions{
		Repo:        vcs.NewMustRepo("acme", "test"),
		Ref:         sha,
		Status:      vcs.PendingStatus,
		TargetURL:   url,
		Description: "it's so exciting",
	})
	require.NoError(t, err)

	status := server.GetStatus(t, ctx)
	assert.NotNil(t, status)
	assert.Equal(t, forgejo.StatusPending, status.State)
}
