package integration

import (
	"testing"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/cloud"
	"github.com/leg100/otf/internal/github"
	"github.com/leg100/otf/internal/testutils"
	"github.com/leg100/otf/internal/workspace"
	"github.com/stretchr/testify/require"
)

// TestIntegration_GithubPR demonstrates the spawning of runs in response to
// opening and updating a pull-request on github.
func TestIntegration_GithubPR(t *testing.T) {
	t.Parallel()

	// create an otf daemon with a fake github backend, serve up a repo and its
	// contents via tarball.
	repo := cloud.NewTestRepo()
	daemon := setup(t, nil,
		github.WithRepo(repo),
		github.WithArchive(testutils.ReadFile(t, "../testdata/github.tar.gz")),
	)

	// create workspace connected to github repo
	provider := daemon.createVCSProvider(t, ctx, nil)
	_, err := daemon.CreateWorkspace(ctx, workspace.CreateOptions{
		Name:         internal.String("workspace-1"),
		Organization: &provider.Organization,
		ConnectOptions: &workspace.ConnectOptions{
			VCSProviderID: provider.ID,
			RepoPath:      repo,
		},
	})
	require.NoError(t, err)

	// a pull request is opened on github which triggers an event
	push := testutils.ReadFile(t, "./fixtures/github_pull_opened.json")
	daemon.SendEvent(t, github.PullRequest, push)

	// github should receive multiple pending status updates followed by a final
	// update with details of planned resources
	require.Equal(t, "pending", daemon.GetStatus(t, ctx).GetState())
	require.Equal(t, "pending", daemon.GetStatus(t, ctx).GetState())
	require.Equal(t, "pending", daemon.GetStatus(t, ctx).GetState())
	require.Equal(t, "pending", daemon.GetStatus(t, ctx).GetState())
	got := daemon.GetStatus(t, ctx)
	require.Equal(t, "success", got.GetState())
	require.Equal(t, "planned: +2/~0/−0", got.GetDescription())

	// the pull request is updated with another commit
	update := testutils.ReadFile(t, "./fixtures/github_pull_update.json")
	daemon.SendEvent(t, github.PullRequest, update)

	// github should receive multiple pending status updates followed by a final
	// update with details of planned resources
	require.Equal(t, "pending", daemon.GetStatus(t, ctx).GetState())
	require.Equal(t, "pending", daemon.GetStatus(t, ctx).GetState())
	require.Equal(t, "pending", daemon.GetStatus(t, ctx).GetState())
	require.Equal(t, "pending", daemon.GetStatus(t, ctx).GetState())
	got = daemon.GetStatus(t, ctx)
	require.Equal(t, "success", got.GetState())
	require.Equal(t, "planned: +2/~0/−0", got.GetDescription())
}
