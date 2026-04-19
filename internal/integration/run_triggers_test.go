package integration

import (
	"fmt"
	"regexp"
	"testing"

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

		// Create run on triggering workspace to test that it triggers a run on the
		// triggered workspace.

		// The triggered workspace first needs some config.
		_ = daemon.createAndUploadConfigurationVersion(t, ctx, triggered, nil)

		// Create a run on the triggering workspace
		cv1 := daemon.createAndUploadConfigurationVersion(t, ctx, triggering, nil)
		triggeringRun := daemon.createRun(t, ctx, triggering, cv1, nil)

		// navigate to run page
		_, err = page.Goto(daemon.URL(paths.Run(triggeringRun.ID)))
		require.NoError(t, err)

		// Wait for run to reach planned status
		triggeringRunStatusID := fmt.Sprintf("#%s-status", triggeringRun.ID.String())
		err = expect.Locator(page.Locator(triggeringRunStatusID)).ToHaveText("planned")
		require.NoError(t, err)

		// Apply run
		err = page.Locator(`//button[@id='apply-button']`).Click()
		require.NoError(t, err)

		// Once run is applied expect alert to be scrolled into view confirming
		// another run has been triggered.
		triggeredRunAlert := page.Locator("//div[@id='triggered-run-alerts']").GetByRole("alert")
		err = expect.Locator(triggeredRunAlert).ToBeInViewport()
		require.NoError(t, err)

		triggeredRunAlertRe := regexp.MustCompile(`Triggered run-.* in connected workspace`)
		err = expect.Locator(triggeredRunAlert).ToContainText(triggeredRunAlertRe)
		require.NoError(t, err)

		// Click on alert to take user to triggered run's page.
		err = triggeredRunAlert.Locator("a").Click()
		require.NoError(t, err)

		// Check that the triggered run page states that it was triggered by another
		// run.
		err = expect.Locator(page.Locator(`#triggering-run`)).ToHaveText("Triggered by " + triggeringRun.ID.String())
		require.NoError(t, err)

		// Enable auto-apply
		//
		// Navigate to triggers page
		_, err = page.Goto(daemon.URL(paths.EditTriggersWorkspace(triggered.ID)))
		require.NoError(t, err)

		err = page.Locator(`//input[@id='auto-apply']`).Click()
		require.NoError(t, err)

		// Should redirect to same page and show alert.
		err = expect.Locator(page.GetByRole("alert")).ToHaveText("updated auto apply setting: true")
		require.NoError(t, err)

		// Auto-apply should be enabled.
		err = expect.Locator(page.Locator(`//input[@id='auto-apply']`)).ToBeChecked()
		require.NoError(t, err)

		// Auto-apply should be enabled on workspace.
		triggered = daemon.getWorkspace(t, ctx, triggered.ID)
		require.True(t, triggered.AutoApplyRunTrigger)

		// Delete connection via UI.
		//
		// Navigate to triggers page
		_, err = page.Goto(daemon.URL(paths.EditTriggersWorkspace(triggered.ID)))
		require.NoError(t, err)

		// Now delete the connection to the triggering workspace.
		err = page.Locator(`//table/tbody/tr/td[2]//button`).Click()
		require.NoError(t, err)

		// Should redirect to same page and show alert.
		err = expect.Locator(page.GetByRole("alert")).ToContainText("deleted trigger")
		require.NoError(t, err)
	})
}
