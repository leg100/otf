package integration

import (
	"strings"
	"testing"

	"github.com/chromedp/cdproto/input"
	"github.com/chromedp/chromedp"
	"github.com/chromedp/chromedp/kb"
	"github.com/leg100/otf/internal/github"
	"github.com/leg100/otf/internal/testutils"
	"github.com/leg100/otf/internal/vcs"
	"github.com/stretchr/testify/require"
)

// TestIntegration_WorkspaceUI demonstrates management of workspaces via the UI.
func TestIntegration_WorkspaceUI(t *testing.T) {
	integrationTest(t)

	repo := vcs.NewTestRepo()
	daemon, org, ctx := setup(t, nil,
		github.WithRepo(repo),
		github.WithArchive(testutils.ReadFile(t, "../testdata/github.tar.gz")),
	)
	// create vcs provider for authenticating to github backend
	provider := daemon.createVCSProvider(t, ctx, org)

	// demonstrate listing and searching
	browser.Run(t, ctx, chromedp.Tasks{
		createWorkspace(t, daemon.Hostname(), org.Name, "workspace-1"),
		createWorkspace(t, daemon.Hostname(), org.Name, "workspace-12"),
		createWorkspace(t, daemon.Hostname(), org.Name, "workspace-2"),
		chromedp.Navigate(workspacesURL(daemon.Hostname(), org.Name)),
		// search for 'workspace-1' which should produce two results
		chromedp.Focus(`input[type="search"]`, chromedp.NodeVisible, chromedp.ByQuery),
		input.InsertText("workspace-1"),
		chromedp.Submit(`input[type="search"]`, chromedp.ByQuery),
		chromedp.WaitVisible(`div.widget`, chromedp.AtLeast(2)),
		// and workspace-2 should not be visible
		chromedp.WaitNotPresent(`//*[@id="item-workspace-workspace-2"]`),
		// clear search term
		chromedp.SendKeys(`input[type="search"]`, strings.Repeat(kb.Delete, len("workspace-1")), chromedp.ByQuery),
		// now workspace-2 should be visible (updated via ajax)
		chromedp.WaitVisible(`//*[@id="item-workspace-workspace-2"]`),
	})
	// demonstrate setting vcs trigger patterns
	browser.Run(t, ctx, chromedp.Tasks{
		connectWorkspaceTasks(t, daemon.Hostname(), org.Name, "workspace-1", provider.String()),
		chromedp.Navigate(workspaceURL(daemon.Hostname(), org.Name, "workspace-1")),
		// go to workspace settings
		chromedp.Click(`//a[text()='settings']`),
		// default should be set to always trigger runs
		chromedp.WaitVisible(`input#vcs-triggers-always:checked`, chromedp.ByQuery),
		// select trigger patterns strategy
		chromedp.Click(`input#vcs-triggers-patterns`, chromedp.ByQuery),
		// add glob patterns
		chromedp.Focus(`#new_path`, chromedp.NodeVisible),
		input.InsertText(`/foo/*.tf`),
		chromedp.Click(`button#add-pattern`, chromedp.ByQuery),
		input.InsertText(`/bar/*.tf`),
		chromedp.Click(`button#add-pattern`, chromedp.ByQuery),
		input.InsertText(`/baz/*.tf`),
		chromedp.Click(`button#add-pattern`, chromedp.ByQuery),
		screenshot(t, "workspace_edit_trigger_patterns"),
		// check patterns are listed
		matchText(t, `span#trigger-pattern-1`, `/foo/\*.tf`, chromedp.ByQuery),
		matchText(t, `span#trigger-pattern-2`, `/bar/\*.tf`, chromedp.ByQuery),
		matchText(t, `span#trigger-pattern-3`, `/baz/\*.tf`, chromedp.ByQuery),
		// delete glob pattern
		chromedp.Click(`button#delete-pattern-2`, chromedp.ByQuery),
		// check pattern is removed from list
		matchText(t, `span#trigger-pattern-1`, `/foo/\*.tf`, chromedp.ByQuery),
		matchText(t, `span#trigger-pattern-2`, `/baz/\*.tf`, chromedp.ByQuery),
		// submit
		chromedp.Submit(`//button[text()='Save changes']`),
		// confirm updated
		matchText(t, "//div[@role='alert']", "updated workspace"),
	})
	// check UI has correctly updated the workspace resource
	ws, err := daemon.GetWorkspaceByName(ctx, org.Name, "workspace-1")
	require.NoError(t, err)
	require.Len(t, ws.TriggerPatterns, 2)
	require.Contains(t, ws.TriggerPatterns, "/foo/*.tf")
	require.Contains(t, ws.TriggerPatterns, "/baz/*.tf")

	// set vcs trigger to use tag regex
	browser.Run(t, ctx, chromedp.Tasks{
		chromedp.Navigate(workspaceURL(daemon.Hostname(), org.Name, "workspace-1")),
		// go to workspace settings
		chromedp.Click(`//a[text()='settings']`),
		// trigger patterns strategy should be set
		chromedp.WaitVisible(`input#vcs-triggers-patterns:checked`, chromedp.ByQuery),
		// select tag trigger strategy
		chromedp.Click(`input#vcs-triggers-tag`, chromedp.ByQuery),
		// select tag prefix pattern
		chromedp.Click(`input#tags-regex-prefix`, chromedp.ByQuery),
		// submit
		chromedp.Submit(`//button[text()='Save changes']`),
		// confirm updated
		matchText(t, "//div[@role='alert']", "updated workspace"),
		// tag prefix pattern should be set
		chromedp.WaitVisible(`input#vcs-triggers-tag:checked`, chromedp.ByQuery),
		chromedp.WaitVisible(`input#tags-regex-prefix:checked`, chromedp.ByQuery),
	})
	// check UI has correctly updated the workspace resource
	ws, err = daemon.GetWorkspaceByName(ctx, org.Name, "workspace-1")
	require.NoError(t, err)
	require.Len(t, ws.TriggerPatterns, 0)
	require.NotNil(t, ws.Connection)
	require.Equal(t, `\d+\.\d+\.\d+$`, ws.Connection.TagsRegex)

	// set vcs branch
	browser.Run(t, ctx, chromedp.Tasks{
		chromedp.Navigate(workspaceURL(daemon.Hostname(), org.Name, "workspace-1")),
		// go to workspace settings
		chromedp.Click(`//a[text()='settings']`),
		// tag regex strategy should be set
		chromedp.WaitVisible(`input#vcs-triggers-tag:checked`, chromedp.ByQuery),
		// set vcs branch
		chromedp.Focus(`input#vcs-branch`, chromedp.ByQuery, chromedp.NodeVisible),
		input.InsertText(`dev`),
		// submit
		chromedp.Submit(`//button[text()='Save changes']`),
		// confirm updated
		matchText(t, "//div[@role='alert']", "updated workspace"),
	})
	// check UI has correctly updated the workspace resource
	ws, err = daemon.GetWorkspaceByName(ctx, org.Name, "workspace-1")
	require.NoError(t, err)
	require.Equal(t, "dev", ws.Connection.Branch)

	// permit applies from the CLI
	browser.Run(t, ctx, chromedp.Tasks{
		chromedp.Navigate(workspaceURL(daemon.Hostname(), org.Name, "workspace-1")),
		// go to workspace settings
		chromedp.Click(`//a[text()='settings']`),
		// allow applies from the CLI
		chromedp.Click(`input#allow-cli-apply`, chromedp.ByQuery),
		chromedp.Submit(`//button[text()='Save changes']`),
		matchText(t, "//div[@role='alert']", "updated workspace"),
		// checkbox should be checked
		chromedp.WaitVisible(`input#allow-cli-apply:checked`, chromedp.ByQuery),
	})
	// check UI has correctly updated the workspace resource
	ws, err = daemon.GetWorkspaceByName(ctx, org.Name, "workspace-1")
	require.NoError(t, err)
	require.Equal(t, true, ws.Connection.AllowCLIApply)

	// set description
	browser.Run(t, ctx, chromedp.Tasks{
		chromedp.Navigate(workspaceURL(daemon.Hostname(), org.Name, "workspace-1")),
		// go to workspace settings
		chromedp.Click(`//a[text()='settings']`), waitLoaded,
		// enter a description
		chromedp.Focus(`textarea#description`, chromedp.ByQuery, chromedp.NodeVisible),
		input.InsertText(`my big fat workspace`),
		// submit
		chromedp.Submit(`//button[text()='Save changes']`), waitLoaded,
		// confirm updated
		matchText(t, "//div[@role='alert']", "updated workspace"),
		// confirm updated description shows up
		chromedp.WaitVisible(`//textarea[@id='description' and text()='my big fat workspace']`),
	})
	// check UI has correctly updated the workspace resource
	ws, err = daemon.GetWorkspaceByName(ctx, org.Name, "workspace-1")
	require.NoError(t, err)
	require.Equal(t, "my big fat workspace", ws.Description)
}
