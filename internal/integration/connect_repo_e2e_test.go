package integration

import (
	"testing"

	"github.com/leg100/otf/internal/github"
	"github.com/leg100/otf/internal/run"
	"github.com/leg100/otf/internal/testutils"
	"github.com/leg100/otf/internal/vcs"
	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/require"
)

// TestConnectRepoE2E demonstrates connecting a workspace to a VCS repository, pushing a
// git commit which triggers a run on the workspace.
func TestConnectRepoE2E(t *testing.T) {
	integrationTest(t)

	// create an otf daemon with a fake github backend, serve up a repo and its
	// contents via tarball. And register a callback to test receipt of commit
	// statuses
	daemon, org, ctx := setup(t,
		withGithubOption(github.WithRepo(vcs.NewMustRepo("leg100", "tfc-workspaces"))),
		withGithubOption(github.WithCommit("0335fb07bb0244b7a169ee89d15c7703e4aaf7de")),
		withGithubOption(github.WithArchive(testutils.ReadFile(t, "../testdata/github.tar.gz"))),
	)
	// create vcs provider for authenticating to github backend
	provider := daemon.createVCSProvider(t, ctx, org, nil)

	browser.New(t, ctx, func(page playwright.Page) {
		createWorkspace(t, page, daemon.System.Hostname(), org.Name, "my-test-workspace")
		connectWorkspaceTasks(t, page, daemon.System.Hostname(), org.Name, "my-test-workspace", provider.String())
		// we can now start a run via the web ui, which'll retrieve the tarball from
		// the fake github server
		startRunTasks(t, page, daemon.System.Hostname(), org.Name, "my-test-workspace", run.PlanAndApplyOperation, true)

		// Now we test the webhook functionality by sending an event to the daemon
		// (which would usually be triggered by a git push to github). The event
		// should trigger a run on the workspace.

		// generate and send push event
		push := testutils.ReadFile(t, "fixtures/github_push.json")
		daemon.SendEvent(t, github.PushEvent, push)

		// commit-triggered run should appear as latest run on workspace
		//
		// go to workspace
		_, err := page.Goto(workspaceURL(daemon.System.Hostname(), org.Name, "my-test-workspace"))
		require.NoError(t, err)
		// branch should match that of push event
		err = expect.Locator(page.Locator(`//div[@id='latest-run']//span[@id='vcs-branch' and text()='master']`)).ToBeVisible()
		require.NoError(t, err)
		// commit should match that of push event
		err = expect.Locator(page.Locator(`//div[@id='latest-run']//a[@id='commit-sha-abbrev']`)).ToContainText("42d6fc7")
		require.NoError(t, err)
		// user should match that of push event
		err = expect.Locator(page.Locator(`//div[@id='latest-run']//a[@id='vcs-username']`)).ToHaveText("leg100")
		require.NoError(t, err)
		// because run was triggered from github, the github icon should be visible.
		err = expect.Locator(page.Locator(`//div[@id='latest-run']//*[@id='github-icon']`)).ToBeVisible()
		require.NoError(t, err)

		// GitHub should receive one pending status update followed by a final
		// update with details of planned resources.
		require.Equal(t, "pending", daemon.GetStatus(t, ctx).GetState())
		got := daemon.GetStatus(t, ctx)
		require.Equal(t, "success", got.GetState())
		require.Equal(t, "planned: +0/~0/−0", got.GetDescription())

		// Clean up after ourselves by disconnecting the workspace and deleting the
		// workspace and vcs provider
		//
		// go to workspace
		_, err = page.Goto(workspaceURL(daemon.System.Hostname(), org.Name, "my-test-workspace"))
		require.NoError(t, err)
		// go to workspace settings
		err = page.Locator(`//li[@id='menu-item-settings']/a`).Click()
		require.NoError(t, err)
		// click disconnect button
		err = page.Locator(`//button[@id='disconnect-workspace-repo-button']`).Click()
		require.NoError(t, err)
		// confirm disconnected
		err = expect.Locator(page.GetByRole("alert")).ToHaveText("disconnected workspace from repo")
		require.NoError(t, err)
		// go to workspace settings
		err = page.Locator(`//li[@id='menu-item-settings']/a`).Click()
		require.NoError(t, err)
		// delete workspace
		err = page.Locator(`//button[@id='delete-workspace-button']`).Click()
		require.NoError(t, err)
		// confirm deletion
		err = expect.Locator(page.GetByRole("alert")).ToHaveText("deleted workspace: my-test-workspace")
		require.NoError(t, err)
		//
		// delete vcs provider
		//
		// go to org
		_, err = page.Goto(organizationURL(daemon.System.Hostname(), org.Name))
		require.NoError(t, err)
		// go to vcs providers
		err = page.Locator("#menu-item-vcs-providers > a").Click()
		require.NoError(t, err)
		// edit provider
		err = page.Locator(`//button[@id='edit-button']`).Click()
		require.NoError(t, err)
		// delete provider
		err = page.Locator(`//button[@id='delete-vcs-provider-button']`).Click()
		require.NoError(t, err)
		err = expect.Locator(page.GetByRole("alert")).ToHaveText(`deleted provider: Github-Token`)
		require.NoError(t, err)
	})
}
