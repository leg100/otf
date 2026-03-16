package integration

import (
	"testing"

	"github.com/leg100/otf/internal/github/testserver"
	"github.com/leg100/otf/internal/testutils"
	"github.com/leg100/otf/internal/ui/paths"
	"github.com/leg100/otf/internal/vcs"
	"github.com/leg100/otf/internal/workspace"
	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGithubPullRequest demonstrates github pull request events triggering runs.
func TestGithubPullRequest(t *testing.T) {
	integrationTest(t)

	// create an OTF daemon with a fake github backend, serve up a repo and its
	// contents via tarball, and setup a fake pull request with a list of files
	// it has changed.
	daemon, org, ctx := setup(t, withGithubOptions(
		testserver.WithRepo(vcs.NewMustRepo("leg100", "otf-workspaces")),
		testserver.WithArchive(testutils.ReadFile(t, "../testdata/github.tar.gz")),
		testserver.WithPullRequest("2", "/nomatch.tf", "/foo/bar/match.tf"),
	))

	provider := daemon.createVCSProvider(t, ctx, org, nil)
	ws, err := daemon.Workspaces.CreateWorkspace(ctx, workspace.CreateOptions{
		Name:            new("dev"),
		Organization:    &org.Name,
		TriggerPatterns: []string{"/foo/**/*.tf"},
		ConnectOptions: &workspace.ConnectOptions{
			VCSProviderID: &provider.ID,
			RepoPath:      new(vcs.NewMustRepo("leg100", "otf-workspaces")),
		},
	})
	require.NoError(t, err)

	// send events
	events := []struct {
		path   string
		commit string
	}{
		{
			path:   "fixtures/github_pull_opened.json",
			commit: "c560613",
		},
		{
			path:   "fixtures/github_pull_update.json",
			commit: "067e2b4",
		},
	}
	for _, event := range events {
		pull := testutils.ReadFile(t, event.path)
		daemon.SendEvent(t, testserver.PullRequest, pull)

		// commit-triggered run should appear as latest run on workspace
		browser.New(t, ctx, func(page playwright.Page) {
			// go to runs
			_, err = page.Goto(daemon.URL(paths.Runs(ws.ID)))
			require.NoError(t, err)
			// should be one run widget with info matching the pull request
			err = expect.Locator(page.Locator(`//a[@id='pull-request-link']`)).ToHaveText(`#2`)
			require.NoError(t, err)
			err = expect.Locator(page.Locator(`//a[@id='vcs-username']`)).ToHaveText(`leg100`)
			require.NoError(t, err)
			err = expect.Locator(page.Locator(`//a[@id='commit-sha-abbrev']`)).ToContainText(event.commit)
			require.NoError(t, err)
		})

		// github should receive several pending status updates followed by a final
		// update with details of planned resources
		var pending int
		for {
			got := daemon.GetStatus(t, ctx)
			switch got.GetState() {
			case "pending":
				pending++
			case "success":
				// Expect to have received at least one pending status update
				// before the final update
				assert.GreaterOrEqual(t, pending, 1)
				require.Equal(t, "planned: +2/~0/−0", got.GetDescription())
				return
			}
		}
	}
}
