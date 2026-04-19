package integration

import (
	"testing"

	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/runstatus"
	"github.com/leg100/otf/internal/ui/paths"
	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/require"
)

// TestRunTriggers tests run trigger functionality.
func TestRunTriggers(t *testing.T) {
	integrationTest(t)

	daemon, org, ctx := setup(t)

	triggered := daemon.createWorkspace(t, ctx, org)
	triggering := daemon.createWorkspace(t, ctx, org)

	// Create connection between workspaces using UI.
	browser.New(t, ctx, func(page playwright.Page) {
		// navigate to workspace page
		_, err := page.Goto(daemon.URL(paths.Workspace(triggered.ID)))
		require.NoError(t, err)

		// navigate to settings page
		err = page.Locator("#menu-item-settings > a").Click()
		require.NoError(t, err)

		// navigate to run triggers page
		err = page.Locator("#menu-item-run-triggers > a").Click()
		require.NoError(t, err)

		// connect to triggering workspace
		_, err = page.Locator(`//select[@id="connect_workspace"]`).SelectOption(playwright.SelectOptionValues{
			ValuesOrLabels: new([]string{triggering.ID.String()}),
		})
		require.NoError(t, err)

		// form should auto submit and redirect to same page with triggering
		// workspace listed as connected workspace.
		err = expect.Locator(page.Locator(`//table/tbody/tr/td[1]`)).ToHaveText(triggering.Name)
		require.NoError(t, err)
	})

	// Create run on triggering workspace to test that it triggers a run on the
	// triggered workspace.

	// The triggered workspace first needs some config.
	_ = daemon.createAndUploadConfigurationVersion(t, ctx, triggered, nil)

	// Create a run on the triggering workspace
	cv1 := daemon.createAndUploadConfigurationVersion(t, ctx, triggering, nil)
	triggeringRun := daemon.createRun(t, ctx, triggering, cv1, nil)

	// Wait for run to be applied on the triggering workspace.
	for re := range daemon.runEvents {
		require.True(t, re.Payload.WorkspaceID == triggering.ID)

		switch re.Payload.Status {
		case runstatus.Planned:
			err := daemon.Runs.ApplyRun(ctx, re.Payload.ID)
			require.NoError(t, err)
		case runstatus.Applied:
			goto triggeringRunApplied
		}
	}
triggeringRunApplied:

	// Run should now be automatically created on the triggered workspace.
	var triggeredRunID resource.ID
	for re := range daemon.runEvents {
		require.True(t, re.Payload.WorkspaceID == triggered.ID)

		switch re.Payload.Status {
		case runstatus.Planned:
			err := daemon.Runs.ApplyRun(ctx, re.Payload.ID)
			require.NoError(t, err)
		case runstatus.Applied:
			triggeredRunID = re.Payload.ID
			goto triggeredRunApplied
		}
	}
triggeredRunApplied:

	// Check that the triggered run page states that it was triggered by another
	// run.
	browser.New(t, ctx, func(page playwright.Page) {
		// navigate to run page
		_, err := page.Goto(daemon.URL(paths.Run(triggeredRunID)))
		require.NoError(t, err)

		err = expect.Locator(page.Locator(`//div[@id='triggering-run']/span[2]`)).ToHaveText("Triggered by " + triggeringRun.ID.String())
		require.NoError(t, err)
	})

	// Delete connection via UI.
	browser.New(t, ctx, func(page playwright.Page) {
		// navigate to triggers page
		_, err := page.Goto(daemon.URL(paths.EditTriggersWorkspace(triggered.ID)))
		require.NoError(t, err)

		// now delete the connection to the triggering workspace.
		err = page.Locator(`//table/tbody/tr/td[2]//button`).Click()
		require.NoError(t, err)

		// should redirect to same page and show alert.
		err = expect.Locator(page.GetByRole("alert")).ToContainText("deleted trigger")
		require.NoError(t, err)
	})
}
