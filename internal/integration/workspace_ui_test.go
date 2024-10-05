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
	page := browser.New(t, ctx)
		createWorkspace(t, daemon.System.Hostname(), org.Name, "workspace-1"),
		createWorkspace(t, daemon.System.Hostname(), org.Name, "workspace-12"),
		createWorkspace(t, daemon.System.Hostname(), org.Name, "workspace-2"),
		_, err = page.Goto(workspacesURL(daemon.System.Hostname(), org.Name))
require.NoError(t, err)
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
	page := browser.New(t, ctx)
		connectWorkspaceTasks(t, daemon.System.Hostname(), org.Name, "workspace-1", provider.String()),
		_, err = page.Goto(workspaceURL(daemon.System.Hostname(), org.Name, "workspace-1"))
require.NoError(t, err)
		// go to workspace settings
		err := page.Locator(`//a[text()='settings']`).Click()
require.NoError(t, err)
		// default should be set to always trigger runs
		chromedp.WaitVisible(`input#vcs-triggers-always:checked`, chromedp.ByQuery),
		// select trigger patterns strategy
		err := page.Locator(`input#vcs-triggers-patterns`).Click()
require.NoError(t, err)
		// add glob patterns
		chromedp.Focus(`#new_path`, chromedp.NodeVisible),
		input.InsertText(`/foo/*.tf`),
		err := page.Locator(`button#add-pattern`).Click()
require.NoError(t, err)
		input.InsertText(`/bar/*.tf`),
		err := page.Locator(`button#add-pattern`).Click()
require.NoError(t, err)
		input.InsertText(`/baz/*.tf`),
		err := page.Locator(`button#add-pattern`).Click()
require.NoError(t, err)
		//screenshot(t, "workspace_edit_trigger_patterns"),
		// check patterns are listed
		matchText(t, `span#trigger-pattern-1`, `/foo/\*.tf`, chromedp.ByQuery),
		matchText(t, `span#trigger-pattern-2`, `/bar/\*.tf`, chromedp.ByQuery),
		matchText(t, `span#trigger-pattern-3`, `/baz/\*.tf`, chromedp.ByQuery),
		// delete glob pattern
		err := page.Locator(`button#delete-pattern-2`).Click()
require.NoError(t, err)
		// check pattern is removed from list
		matchText(t, `span#trigger-pattern-1`, `/foo/\*.tf`, chromedp.ByQuery),
		matchText(t, `span#trigger-pattern-2`, `/baz/\*.tf`, chromedp.ByQuery),
		// submit
		chromedp.Submit(`//button[text()='Save changes']`),
		// confirm updated
		matchText(t, "//div[@role='alert']", "updated workspace"),
	})
	// check UI has correctly updated the workspace resource
	ws, err := daemon.Workspaces.GetByName(ctx, org.Name, "workspace-1")
	require.NoError(t, err)
	require.Len(t, ws.TriggerPatterns, 2)
	require.Contains(t, ws.TriggerPatterns, "/foo/*.tf")
	require.Contains(t, ws.TriggerPatterns, "/baz/*.tf")

	// set vcs trigger to use tag regex
	page := browser.New(t, ctx)
		_, err = page.Goto(workspaceURL(daemon.System.Hostname(), org.Name, "workspace-1"))
require.NoError(t, err)
		// go to workspace settings
		err := page.Locator(`//a[text()='settings']`).Click()
require.NoError(t, err)
		// trigger patterns strategy should be set
		chromedp.WaitVisible(`input#vcs-triggers-patterns:checked`, chromedp.ByQuery),
		// select tag trigger strategy
		err := page.Locator(`input#vcs-triggers-tag`).Click()
require.NoError(t, err)
		// select tag prefix pattern
		err := page.Locator(`input#tags-regex-prefix`).Click()
require.NoError(t, err)
		// submit
		chromedp.Submit(`//button[text()='Save changes']`),
		// confirm updated
		matchText(t, "//div[@role='alert']", "updated workspace"),
		// tag prefix pattern should be set
		chromedp.WaitVisible(`input#vcs-triggers-tag:checked`, chromedp.ByQuery),
		chromedp.WaitVisible(`input#tags-regex-prefix:checked`, chromedp.ByQuery),
	})
	// check UI has correctly updated the workspace resource
	ws, err = daemon.Workspaces.GetByName(ctx, org.Name, "workspace-1")
	require.NoError(t, err)
	require.Len(t, ws.TriggerPatterns, 0)
	require.NotNil(t, ws.Connection)
	require.Equal(t, `\d+\.\d+\.\d+$`, ws.Connection.TagsRegex)

	// set vcs branch
	page := browser.New(t, ctx)
		_, err = page.Goto(workspaceURL(daemon.System.Hostname(), org.Name, "workspace-1"))
require.NoError(t, err)
		// go to workspace settings
		err := page.Locator(`//a[text()='settings']`).Click()
require.NoError(t, err)
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
	ws, err = daemon.Workspaces.GetByName(ctx, org.Name, "workspace-1")
	require.NoError(t, err)
	require.Equal(t, "dev", ws.Connection.Branch)

	// permit applies from the CLI
	page := browser.New(t, ctx)
		_, err = page.Goto(workspaceURL(daemon.System.Hostname(), org.Name, "workspace-1"))
require.NoError(t, err)
		// go to workspace settings
		err := page.Locator(`//a[text()='settings']`).Click()
require.NoError(t, err)
		// allow applies from the CLI
		err := page.Locator(`input#allow-cli-apply`).Click()
require.NoError(t, err)
		chromedp.Submit(`//button[text()='Save changes']`),
		matchText(t, "//div[@role='alert']", "updated workspace"),
		// checkbox should be checked
		chromedp.WaitVisible(`input#allow-cli-apply:checked`, chromedp.ByQuery),
	})
	// check UI has correctly updated the workspace resource
	ws, err = daemon.Workspaces.GetByName(ctx, org.Name, "workspace-1")
	require.NoError(t, err)
	require.Equal(t, true, ws.Connection.AllowCLIApply)

	// set description
	page := browser.New(t, ctx)
		_, err = page.Goto(workspaceURL(daemon.System.Hostname(), org.Name, "workspace-1"))
require.NoError(t, err)
		// go to workspace settings
		err := page.Locator(`//a[text()='settings']`).Click()
require.NoError(t, err) waitLoaded,
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
	ws, err = daemon.Workspaces.GetByName(ctx, org.Name, "workspace-1")
	require.NoError(t, err)
	require.Equal(t, "my big fat workspace", ws.Description)
}
