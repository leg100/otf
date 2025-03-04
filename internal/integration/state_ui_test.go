package integration

import (
	"regexp"
	"testing"

	"github.com/leg100/otf/internal/runstatus"
	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/require"
)

// TestIntegration_StateUI demonstrates the displaying of terraform state via
// the UI
func TestIntegration_StateUI(t *testing.T) {
	integrationTest(t)

	daemon, org, ctx := setup(t, nil)
	ws := daemon.createWorkspace(t, ctx, org)
	cv := daemon.createAndUploadConfigurationVersion(t, ctx, ws, nil)

	// create run and wait for it to complete
	r := daemon.createRun(t, ctx, ws, cv, nil)
	planned := daemon.waitRunStatus(t, r.ID, runstatus.Planned)
	err := daemon.Runs.Apply(ctx, planned.ID)
	require.NoError(t, err)
	daemon.waitRunStatus(t, r.ID, runstatus.Applied)

	browser.New(t, ctx, func(page playwright.Page) {
		_, err := page.Goto(workspaceURL(daemon.System.Hostname(), org.Name, ws.Name))
		require.NoError(t, err)

		err = expect.Locator(page.Locator(`//input[@id='resources-label']`)).ToHaveAttribute(`aria-label`, regexp.MustCompile(`Resources \(1\)`))
		require.NoError(t, err)

		err = expect.Locator(page.Locator(`//input[@id='outputs-label']`)).ToHaveAttribute(`aria-label`, regexp.MustCompile(`Outputs \(0\)`))
		require.NoError(t, err)

		err = expect.Locator(page.Locator(`//table[@id='resources-table']/tbody/tr/td[1]`)).ToHaveText(`test`)
		require.NoError(t, err)

		err = expect.Locator(page.Locator(`//table[@id='resources-table']/tbody/tr/td[2]`)).ToHaveText(`hashicorp/null`)
		require.NoError(t, err)

		err = expect.Locator(page.Locator(`//table[@id='resources-table']/tbody/tr/td[3]`)).ToHaveText(`null_resource`)
		require.NoError(t, err)

		err = expect.Locator(page.Locator(`//table[@id='resources-table']/tbody/tr/td[4]`)).ToHaveText(`root`)
		require.NoError(t, err)
	})
}
