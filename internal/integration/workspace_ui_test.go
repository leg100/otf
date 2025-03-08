package integration

import (
	"testing"
	"time"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/github"
	"github.com/leg100/otf/internal/runstatus"
	"github.com/leg100/otf/internal/testutils"
	"github.com/leg100/otf/internal/vcs"
	"github.com/leg100/otf/internal/workspace"
	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/require"
)

// TestIntegration_WorkspaceUI demonstrates management of workspaces via the UI.
func TestIntegration_WorkspaceUI(t *testing.T) {
	integrationTest(t)

	t.Run("create", func(t *testing.T) {
		daemon, org, ctx := setup(t, nil)
		browser.New(t, ctx, func(page playwright.Page) {
			_, err := page.Goto(organizationURL(daemon.System.Hostname(), org.Name))
			require.NoError(t, err)

			err = page.Locator("#menu-item-workspaces > a").Click()
			require.NoError(t, err)

			err = page.Locator("#new-workspace-button").Click()
			require.NoError(t, err)

			err = page.Locator("input#name").Fill("workspace-1")
			require.NoError(t, err)

			err = page.Locator("#create-workspace-button").Click()
			require.NoError(t, err)

			err = expect.Locator(page.GetByRole("alert")).ToHaveText("created workspace: workspace-1")
			require.NoError(t, err)
		})
	})

	// TODO: split into more smaller more targeted tests
	t.Run("listing_and_filtering_and_updating", func(t *testing.T) {
		repo := vcs.NewTestRepo()
		daemon, org, ctx := setup(t, nil,
			github.WithRepo(repo),
			github.WithArchive(testutils.ReadFile(t, "../testdata/github.tar.gz")),
		)
		// create vcs provider for authenticating to github backend
		provider := daemon.createVCSProvider(t, ctx, org)

		// create workspaces for listing and searching/filtering
		ws1, err := daemon.Workspaces.Create(ctx, workspace.CreateOptions{
			Name:         internal.String("workspace-1"),
			Organization: &org.Name,
		})
		require.NoError(t, err)
		ws2, err := daemon.Workspaces.Create(ctx, workspace.CreateOptions{
			Name:         internal.String("workspace-12"),
			Organization: &org.Name,
		})
		require.NoError(t, err)
		ws3, err := daemon.Workspaces.Create(ctx, workspace.CreateOptions{
			Name:         internal.String("workspace-2"),
			Organization: &org.Name,
		})
		require.NoError(t, err)

		// Create some runs to allow filtering workspaces by their current run
		// status
		cv1 := daemon.createAndUploadConfigurationVersion(t, ctx, ws1, nil)
		cv2 := daemon.createAndUploadConfigurationVersion(t, ctx, ws2, nil)
		cv3 := daemon.createAndUploadConfigurationVersion(t, ctx, ws3, nil)
		// A 'planned' run.
		ws1run1planned := daemon.createRun(t, ctx, ws1, cv1, nil)
		_ = daemon.waitRunStatus(t, ws1run1planned.ID, runstatus.Planned)
		// A 'planned' run.
		ws2run1planned := daemon.createRun(t, ctx, ws2, cv2, nil)
		_ = daemon.waitRunStatus(t, ws2run1planned.ID, runstatus.Planned)
		// An 'applied' run.
		ws3run1applied := daemon.createRun(t, ctx, ws3, cv3, nil)
		_ = daemon.waitRunStatus(t, ws3run1applied.ID, runstatus.Planned)
		err = daemon.Runs.Apply(ctx, ws3run1applied.ID)
		require.NoError(t, err)
		_ = daemon.waitRunStatus(t, ws3run1applied.ID, runstatus.Applied)

		// demonstrate listing and searching
		browser.New(t, ctx, func(page playwright.Page) {
			_, err := page.Goto(workspacesURL(daemon.System.Hostname(), org.Name))
			require.NoError(t, err)

			// search for 'workspace-1' which should produce two results
			err = page.Locator(`input[type="search"]`).Fill("workspace-1")
			require.NoError(t, err)

			err = page.Locator(`input[type="search"]`).Press("Enter")
			require.NoError(t, err)

			err = expect.Locator(page.Locator(`div.widget`)).ToHaveCount(2)
			require.NoError(t, err)

			// and workspace-2 should not be visible
			err = expect.Locator(page.Locator(`//*[@id="item-workspace-workspace-2"]`)).Not().ToBeVisible()
			require.NoError(t, err)

			// clear search term
			err = page.Locator(`input[type="search"]`).Clear()
			require.NoError(t, err)

			time.Sleep(time.Hour)

			// now workspace-2 should be visible (updated via ajax)
			err = expect.Locator(page.Locator(`//*[@id="item-workspace-workspace-2"]`)).ToBeVisible()
			require.NoError(t, err)

			// demonstrate setting vcs trigger patterns
			//
			connectWorkspaceTasks(t, page, daemon.System.Hostname(), org.Name, "workspace-1", provider.String())

			_, err = page.Goto(workspaceURL(daemon.System.Hostname(), org.Name, "workspace-1"))
			require.NoError(t, err)

			// go to workspace settings
			err = page.Locator(`//a[text()='settings']`).Click()
			require.NoError(t, err)

			// default should be set to always trigger runs
			err = expect.Locator(page.Locator(`input#vcs-triggers-always:checked`)).ToBeVisible()
			require.NoError(t, err)

			// select trigger patterns strategy
			err = page.Locator(`input#vcs-triggers-patterns`).Click()
			require.NoError(t, err)

			// add glob patterns
			err = page.Locator(`input#new_path`).Fill(`/foo/*.tf`)
			require.NoError(t, err)

			err = page.Locator(`button#add-pattern`).Click()
			require.NoError(t, err)

			err = page.Locator(`input#new_path`).Fill(`/bar/*.tf`)
			require.NoError(t, err)

			err = page.Locator(`button#add-pattern`).Click()
			require.NoError(t, err)

			err = page.Locator(`input#new_path`).Fill(`/baz/*.tf`)
			require.NoError(t, err)

			err = page.Locator(`button#add-pattern`).Click()
			require.NoError(t, err)

			screenshot(t, page, "workspace_edit_trigger_patterns")

			// check patterns are listed
			err = expect.Locator(page.Locator(`span#trigger-pattern-1`)).ToHaveText(`/foo/*.tf`)
			require.NoError(t, err)

			err = expect.Locator(page.Locator(`span#trigger-pattern-2`)).ToHaveText(`/bar/*.tf`)
			require.NoError(t, err)

			err = expect.Locator(page.Locator(`span#trigger-pattern-3`)).ToHaveText(`/baz/*.tf`)
			require.NoError(t, err)

			// delete glob pattern
			err = page.Locator(`button#delete-pattern-2`).Click()
			require.NoError(t, err)

			// check pattern is removed from list
			err = expect.Locator(page.Locator(`span#trigger-pattern-1`)).ToHaveText(`/foo/*.tf`)
			require.NoError(t, err)

			err = expect.Locator(page.Locator(`span#trigger-pattern-2`)).ToHaveText(`/baz/*.tf`)
			require.NoError(t, err)

			// submit
			err = page.GetByRole("button").GetByText("Save changes").Click()
			require.NoError(t, err)

			// confirm updated
			err = expect.Locator(page.GetByRole("alert")).ToHaveText("updated workspace")
			require.NoError(t, err)

			// check UI has correctly updated the workspace resource
			ws, err := daemon.Workspaces.GetByName(ctx, org.Name, "workspace-1")
			require.NoError(t, err)
			require.Len(t, ws.TriggerPatterns, 2)
			require.Contains(t, ws.TriggerPatterns, "/foo/*.tf")
			require.Contains(t, ws.TriggerPatterns, "/baz/*.tf")

			// set vcs trigger to use tag regex
			_, err = page.Goto(workspaceURL(daemon.System.Hostname(), org.Name, "workspace-1"))
			require.NoError(t, err)

			// go to workspace settings
			err = page.Locator(`//a[text()='settings']`).Click()
			require.NoError(t, err)

			// trigger patterns strategy should be set
			err = expect.Locator(page.Locator(`input#vcs-triggers-patterns:checked`)).ToBeVisible()
			require.NoError(t, err)

			// select tag trigger strategy
			err = page.Locator(`input#vcs-triggers-tag`).Click()
			require.NoError(t, err)

			// select tag prefix pattern
			err = page.Locator(`input#tags-regex-prefix`).Click()
			require.NoError(t, err)

			// submit
			err = page.Locator(`//button[text()='Save changes']`).Click()
			require.NoError(t, err)

			// confirm updated
			err = expect.Locator(page.GetByRole("alert")).ToHaveText("updated workspace")
			require.NoError(t, err)

			// tag prefix pattern should be set
			err = expect.Locator(page.Locator(`input#vcs-triggers-tag:checked`)).ToBeVisible()
			require.NoError(t, err)

			err = expect.Locator(page.Locator(`input#tags-regex-prefix:checked`)).ToBeVisible()
			require.NoError(t, err)

			// check UI has correctly updated the workspace resource
			ws, err = daemon.Workspaces.GetByName(ctx, org.Name, "workspace-1")
			require.NoError(t, err)
			require.Len(t, ws.TriggerPatterns, 0)
			require.NotNil(t, ws.Connection)
			require.Equal(t, `\d+\.\d+\.\d+$`, ws.Connection.TagsRegex)

			// set vcs branch
			//
			_, err = page.Goto(workspaceURL(daemon.System.Hostname(), org.Name, "workspace-1"))
			require.NoError(t, err)

			// go to workspace settings
			err = page.Locator(`//a[text()='settings']`).Click()
			require.NoError(t, err)

			// tag regex strategy should be set
			err = expect.Locator(page.Locator(`input#vcs-triggers-tag:checked`)).ToBeVisible()
			require.NoError(t, err)

			// set vcs branch
			err = page.Locator(`input#vcs-branch`).Fill(`dev`)
			require.NoError(t, err)

			// submit
			err = page.GetByRole("button").GetByText("Save changes").Click()
			require.NoError(t, err)

			// confirm updated
			err = expect.Locator(page.GetByRole("alert")).ToHaveText("updated workspace")
			require.NoError(t, err)

			// check UI has correctly updated the workspace resource
			ws, err = daemon.Workspaces.GetByName(ctx, org.Name, "workspace-1")
			require.NoError(t, err)
			require.Equal(t, "dev", ws.Connection.Branch)

			// permit applies from the CLI
			//
			_, err = page.Goto(workspaceURL(daemon.System.Hostname(), org.Name, "workspace-1"))
			require.NoError(t, err)
			// go to workspace settings
			err = page.Locator(`//a[text()='settings']`).Click()
			require.NoError(t, err)

			// allow applies from the CLI
			err = page.Locator(`input#allow-cli-apply`).Click()
			require.NoError(t, err)

			err = page.GetByRole("button").GetByText("Save changes").Click()
			require.NoError(t, err)

			err = expect.Locator(page.GetByRole("alert")).ToHaveText("updated workspace")
			require.NoError(t, err)

			// checkbox should be checked
			err = expect.Locator(page.Locator(`input#allow-cli-apply:checked`)).ToBeVisible()
			require.NoError(t, err)

			// check UI has correctly updated the workspace resource
			ws, err = daemon.Workspaces.GetByName(ctx, org.Name, "workspace-1")
			require.NoError(t, err)
			require.Equal(t, true, ws.Connection.AllowCLIApply)

			// set description

			_, err = page.Goto(workspaceURL(daemon.System.Hostname(), org.Name, "workspace-1"))
			require.NoError(t, err)
			// go to workspace settings
			err = page.Locator(`//a[text()='settings']`).Click()
			require.NoError(t, err)
			// enter a description
			err = page.Locator(`textarea#description`).Fill(`my big fat workspace`)
			require.NoError(t, err)

			// submit
			err = page.GetByRole("button").GetByText("Save changes").Click()
			require.NoError(t, err)

			// confirm updated
			err = expect.Locator(page.GetByRole("alert")).ToHaveText("updated workspace")
			require.NoError(t, err)

			// confirm updated description shows up
			err = expect.Locator(page.Locator(`//textarea[@id='description' and text()='my big fat workspace']`)).ToBeVisible()
			require.NoError(t, err)
		})

		// check UI has correctly updated the workspace resource
		ws, err := daemon.Workspaces.GetByName(ctx, org.Name, "workspace-1")
		require.NoError(t, err)
		require.Equal(t, "my big fat workspace", ws.Description)
	})
}
