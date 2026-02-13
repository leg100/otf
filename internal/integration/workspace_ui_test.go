package integration

import (
	"fmt"
	"testing"

	"github.com/leg100/otf/internal/engine"
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
		daemon, org, ctx := setup(t)

		t.Run("create with no error", func(t *testing.T) {
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

		t.Run("create with error", func(t *testing.T) {
			browser.New(t, ctx, func(page playwright.Page) {
				_, err := page.Goto(organizationURL(daemon.System.Hostname(), org.Name))
				require.NoError(t, err)

				err = page.Locator("#menu-item-workspaces > a").Click()
				require.NoError(t, err)

				err = page.Locator("#new-workspace-button").Click()
				require.NoError(t, err)

				// invalid name
				err = page.Locator("input#name").Fill("$&$*(&*(@")
				require.NoError(t, err)

				err = page.Locator("#create-workspace-button").Click()
				require.NoError(t, err)

				err = expect.Locator(page.GetByRole("alert")).ToHaveText("invalid value for name")
				require.NoError(t, err)
				err = expect.Locator(page.GetByRole("alert")).ToContainClass("alert-error")
				require.NoError(t, err)
			})
		})
	})

	t.Run("get with latest run", func(t *testing.T) {
		daemon, org, ctx := setup(t)
		ws := daemon.createWorkspace(t, ctx, org)
		run := daemon.createRun(t, ctx, ws, nil, nil)

		browser.New(t, ctx, func(page playwright.Page) {
			_, err := page.Goto(workspaceURL(daemon.System.Hostname(), org.Name, ws.Name))
			require.NoError(t, err)

			err = expect.Locator(page.Locator("//div[@id='latest-run']//tbody/tr")).ToHaveId("run-item-" + run.ID.String())
			require.NoError(t, err)

			// confirm 'overview' submenu button is active
			err = expect.Locator(page.Locator(`//*[@id="menu-item-overview"]/a`)).ToHaveClass(`menu-active`)
			require.NoError(t, err)
		})
	})

	t.Run("listing_and_filtering_and_updating", func(t *testing.T) {
		daemon, org, ctx := setup(t)

		// Create lots of workspaces for filtering and updating
		workspaces := make([]*workspace.Workspace, 101)
		for i := range 101 {
			// create workspaces workspaces-{1-101}
			ws, err := daemon.Workspaces.Create(ctx, workspace.CreateOptions{
				Name:         new(fmt.Sprintf("workspace-%d", i+1)),
				Organization: &org.Name,
			})
			require.NoError(t, err)
			workspaces[i] = ws
		}

		// Create some runs to allow filtering workspaces by their current run
		// status
		cv1 := daemon.createAndUploadConfigurationVersion(t, ctx, workspaces[0], nil)
		cv2 := daemon.createAndUploadConfigurationVersion(t, ctx, workspaces[1], nil)
		cv3 := daemon.createAndUploadConfigurationVersion(t, ctx, workspaces[2], nil)
		// A 'planned' run.
		ws1run1planned := daemon.createRun(t, ctx, workspaces[0], cv1, nil)
		_ = daemon.waitRunStatus(t, ctx, ws1run1planned.ID, runstatus.Planned)
		// A 'planned' run.
		ws2run1planned := daemon.createRun(t, ctx, workspaces[1], cv2, nil)
		_ = daemon.waitRunStatus(t, ctx, ws2run1planned.ID, runstatus.Planned)
		// An 'applied' run.
		ws3run1applied := daemon.createRun(t, ctx, workspaces[2], cv3, nil)
		_ = daemon.waitRunStatus(t, ctx, ws3run1applied.ID, runstatus.Planned)
		err := daemon.Runs.Apply(ctx, ws3run1applied.ID)
		require.NoError(t, err)
		_ = daemon.waitRunStatus(t, ctx, ws3run1applied.ID, runstatus.Applied)

		// navigate through different pages and back
		browser.New(t, ctx, func(page playwright.Page) {
			_, err := page.Goto(workspacesURL(daemon.System.Hostname(), org.Name))
			require.NoError(t, err)

			steps := []struct {
				info       string
				goNext     bool
				goPrevious bool
			}{
				{
					info:   "1-20 of 101",
					goNext: true,
				},
				{
					info:   "21-40 of 101",
					goNext: true,
				},
				{
					info:   "41-60 of 101",
					goNext: true,
				},
				{
					info:   "61-80 of 101",
					goNext: true,
				},
				{
					info:   "81-100 of 101",
					goNext: true,
				},
				{
					info:       "101-101 of 101",
					goPrevious: true,
				},
				{
					info:       "81-100 of 101",
					goPrevious: true,
				},
				{
					info:       "61-80 of 101",
					goPrevious: true,
				},
				{
					info:       "41-60 of 101",
					goPrevious: true,
				},
				{
					info:       "21-40 of 101",
					goPrevious: true,
				},
				{
					info: "1-20 of 101",
				},
			}
			for _, step := range steps {
				err = expect.Locator(page.Locator(`#page-info`)).ToHaveText(step.info)
				require.NoError(t, err)

				if step.goNext {
					err = page.Locator("#next-page-link").Click()
					require.NoError(t, err)
				} else if step.goPrevious {
					err = page.Locator("#prev-page-link").Click()
					require.NoError(t, err)
				}
			}
		})

		// demonstrate listing and searching
		browser.New(t, ctx, func(page playwright.Page) {
			_, err := page.Goto(workspacesURL(daemon.System.Hostname(), org.Name))
			require.NoError(t, err)

			// search for 'workspace-1'
			err = page.Locator(`input[type="search"]`).Fill("workspace-1")
			require.NoError(t, err)

			err = page.Locator(`input[type="search"]`).Press("Enter")
			require.NoError(t, err)

			// search for 'workspace-1' should produce 13 results (1,
			// 10-19, 100, 101)
			err = expect.Locator(page.Locator(`//table//tbody//tr`)).ToHaveCount(13)
			require.NoError(t, err)

			// and workspace-2 should not be visible
			err = expect.Locator(page.Locator(`//*[@id="item-workspace-workspace-2"]`)).Not().ToBeVisible()
			require.NoError(t, err)

			// clear search term
			err = page.Locator(`input[type="search"]`).Clear()
			require.NoError(t, err)

			// now workspace-2 should be visible (updated via ajax)
			err = expect.Locator(page.Locator(`//*[@id="item-workspace-workspace-2"]`)).ToBeVisible()
			require.NoError(t, err)
		})
	})

	t.Run("workspace settings", func(t *testing.T) {
		repo := vcs.NewRandomRepo()
		daemon, org, ctx := setup(t, withGithubOptions(
			github.WithRepo(repo),
			github.WithArchive(testutils.ReadFile(t, "../testdata/github.tar.gz")),
		))
		// create vcs provider for authenticating to github backend
		provider := daemon.createVCSProvider(t, ctx, org, nil)
		// create workspace on which edit settings
		ws1 := daemon.createWorkspace(t, ctx, org)

		browser.New(t, ctx, func(page playwright.Page) {
			// demonstrate setting vcs trigger patterns
			//
			connectWorkspaceTasks(t, page, daemon.System.Hostname(), org.Name, ws1.Name, provider.String())

			_, err := page.Goto(workspaceURL(daemon.System.Hostname(), org.Name, ws1.Name))
			require.NoError(t, err)

			// go to workspace settings
			err = page.Locator(`//li[@id='menu-item-settings']/a`).Click()
			require.NoError(t, err)

			// confirm 'settings' submenu button is active
			err = expect.Locator(page.Locator(`//li[@id='menu-item-settings']/a`)).ToHaveClass(`menu-active`)
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
			ws, err := daemon.Workspaces.GetByName(ctx, org.Name, ws1.Name)
			require.NoError(t, err)
			require.Len(t, ws.TriggerPatterns, 2)
			require.Contains(t, ws.TriggerPatterns, "/foo/*.tf")
			require.Contains(t, ws.TriggerPatterns, "/baz/*.tf")

			// set vcs trigger to use tag regex
			_, err = page.Goto(workspaceURL(daemon.System.Hostname(), org.Name, ws1.Name))
			require.NoError(t, err)

			// go to workspace settings
			err = page.Locator(`//li[@id='menu-item-settings']/a`).Click()
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
			ws, err = daemon.Workspaces.GetByName(ctx, org.Name, ws1.Name)
			require.NoError(t, err)
			require.Len(t, ws.TriggerPatterns, 0)
			require.NotNil(t, ws.Connection)
			require.Equal(t, `\d+\.\d+\.\d+$`, ws.Connection.TagsRegex)

			// set vcs branch
			//
			_, err = page.Goto(workspaceURL(daemon.System.Hostname(), org.Name, ws1.Name))
			require.NoError(t, err)

			// go to workspace settings
			err = page.Locator(`//li[@id='menu-item-settings']/a`).Click()
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
			ws, err = daemon.Workspaces.GetByName(ctx, org.Name, ws1.Name)
			require.NoError(t, err)
			require.Equal(t, "dev", ws.Connection.Branch)

			// permit applies from the CLI
			//
			_, err = page.Goto(workspaceURL(daemon.System.Hostname(), org.Name, ws1.Name))
			require.NoError(t, err)
			// go to workspace settings
			err = page.Locator(`//li[@id='menu-item-settings']/a`).Click()
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
			ws, err = daemon.Workspaces.GetByName(ctx, org.Name, ws1.Name)
			require.NoError(t, err)
			require.Equal(t, true, ws.Connection.AllowCLIApply)

			// set description

			_, err = page.Goto(workspaceURL(daemon.System.Hostname(), org.Name, ws1.Name))
			require.NoError(t, err)

			// go to workspace settings
			err = page.Locator(`//li[@id='menu-item-settings']/a`).Click()
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

		t.Run("engine settings", func(t *testing.T) {
			// create workspace on which edit engine settings
			ws := daemon.createWorkspace(t, ctx, org)

			browser.New(t, ctx, func(page playwright.Page) {
				// go to workspace settings
				_, err := page.Goto(workspaceURL(daemon.System.Hostname(), org.Name, ws.Name))
				require.NoError(t, err)
				err = page.Locator(`//li[@id='menu-item-settings']/a`).Click()
				require.NoError(t, err)

				// switch engine from terraform to tofu

				// terraform should be current engine
				err = expect.Locator(page.Locator(`//*[@id='engine-selector']//input[@id='terraform']`)).ToBeChecked()
				require.NoError(t, err)

				// make tofu the current engine instead
				err = page.Locator(`//*[@id='engine-selector']//input[@id='tofu']`).Click()
				require.NoError(t, err)

				// submit
				err = page.GetByRole("button").GetByText("Save changes").Click()
				require.NoError(t, err)

				// confirm tofu is now current engine
				err = expect.Locator(page.Locator(`//*[@id='engine-selector']//input[@id='tofu']`)).ToBeChecked()
				require.NoError(t, err)

				// confirm tofu version is the default version (integration test
				// disables the latest version checker, so the latest version
				// defaults to the default version)
				err = expect.Locator(page.Locator(`//*[@id='engine-version-selector']//input[@id='engine-specific-version']`)).ToHaveValue(engine.Tofu.DefaultVersion())
				require.NoError(t, err)

				// switch tofu version to v2.1.0
				err = page.Locator(`//*[@id='engine-version-selector']//input[@id='engine-specific-version']`).Fill(`2.1.0`)
				require.NoError(t, err)

				// submit
				err = page.GetByRole("button").GetByText("Save changes").Click()
				require.NoError(t, err)

				// expect tofu version to have been updated
				err = expect.Locator(page.Locator(`//*[@id='engine-version-selector']//input[@id='engine-specific-version']`)).ToHaveValue(`2.1.0`)
				require.NoError(t, err)

				// switch tofu version to track latest
				err = page.Locator(`//*[@id='engine-version-selector']//input[@id='engine-version-latest-true']`).Click()
				require.NoError(t, err)

				// submit
				err = page.GetByRole("button").GetByText("Save changes").Click()
				require.NoError(t, err)

				// expect tofu version to now track latest
				err = expect.Locator(page.Locator(`//*[@id='engine-version-selector']//input[@id='engine-version-latest-true']`)).ToBeChecked()
				require.NoError(t, err)
			})
		})
	})

	t.Run("workspace locking", func(t *testing.T) {
		daemon, org, ctx := setup(t)
		ws1 := daemon.createWorkspace(t, ctx, org)

		browser.New(t, ctx, func(page playwright.Page) {
			_, err := page.Goto(workspaceURL(daemon.System.Hostname(), org.Name, ws1.Name))
			require.NoError(t, err)

			// expect workspace to be unlocked by default
			err = expect.Locator(page.Locator(`#lock-state`)).ToContainText("Unlocked")
			require.NoError(t, err)

			// lock workspace
			err = page.Locator(`#lock-button`).Click()
			require.NoError(t, err)

			// expect workspace to now be locked
			err = expect.Locator(page.Locator(`#lock-state`)).ToContainText("Locked")
			require.NoError(t, err)

			// unlock workspace
			err = page.Locator(`#lock-button`).Click()
			require.NoError(t, err)

			// expect workspace to now be unlocked
			err = expect.Locator(page.Locator(`#lock-state`)).ToContainText("Unlocked")
			require.NoError(t, err)
		})
	})
}
