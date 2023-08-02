package integration

import (
	"testing"

	"github.com/chromedp/chromedp"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/cloud"
	"github.com/leg100/otf/internal/github"
	"github.com/leg100/otf/internal/testutils"
	"github.com/leg100/otf/internal/workspace"
	"github.com/stretchr/testify/require"
)

// TestGithubPullRequest demonstrates github pull request events triggering runs.
func TestGithubPullRequest(t *testing.T) {
	integrationTest(t)

	// create an OTF daemon with a fake github backend, serve up a repo and its
	// contents via tarball, and setup a fake pull request with a list of files
	// it has changed.
	repo := cloud.NewTestRepo()
	daemon, org, ctx := setup(t, nil,
		github.WithRepo(repo),
		github.WithArchive(testutils.ReadFile(t, "../testdata/github.tar.gz")),
		github.WithPullRequest("2", "/nomatch.tf", "/foo/bar/match.tf"),
	)

	provider := daemon.createVCSProvider(t, ctx, org)
	ws, err := daemon.CreateWorkspace(ctx, workspace.CreateOptions{
		Name:            internal.String("dev"),
		Organization:    internal.String(org.Name),
		TriggerPatterns: []string{"/foo/**/*.tf"},
		ConnectOptions: &workspace.ConnectOptions{
			VCSProviderID: &provider.ID,
			RepoPath:      &repo,
		},
	})
	require.NoError(t, err)

	// open pull request
	pull := testutils.ReadFile(t, "fixtures/github_pull_opened.json")
	daemon.SendEvent(t, github.PullRequest, pull)

	// commit-triggered run should appear as latest run on workspace
	browser.Run(t, ctx, chromedp.Tasks{
		// go to runs
		chromedp.Navigate(runsURL(daemon.Hostname(), ws.ID)),
		screenshot(t),
		// should be one run widget with info matching the pull request
		chromedp.WaitVisible(`//div[@class='widget']//a[@id='pull-request-link' and text()='#2']`),
		chromedp.WaitVisible(`//div[@class='widget']//a[@id='vcs-username' and text()='@leg100']`),
		chromedp.WaitVisible(`//div[@class='widget']//a[@id='commit-sha-abbrev' and text()='c560613']`),
		screenshot(t),
	})

	// github should receive several pending status updates followed by a final
	// update with details of planned resources
	require.Equal(t, "pending", daemon.GetStatus(t, ctx).GetState())
	require.Equal(t, "pending", daemon.GetStatus(t, ctx).GetState())
	require.Equal(t, "pending", daemon.GetStatus(t, ctx).GetState())
	require.Equal(t, "pending", daemon.GetStatus(t, ctx).GetState())
	got := daemon.GetStatus(t, ctx)
	require.Equal(t, "success", got.GetState())
	require.Equal(t, "planned: +2/~0/−0", got.GetDescription())
}
