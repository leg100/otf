package integration

import (
	"testing"

	"github.com/leg100/otf/internal/run"
	"github.com/leg100/otf/internal/path"
	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/require"
)

// TestStartRunUI tests starting a run via the Web UI before confirming and
// applying the run.
func TestStartRunUI(t *testing.T) {
	integrationTest(t)

	for _, tt := range engineTestSpecs() {
		t.Run(tt.name, func(t *testing.T) {
			// Test with both terraform and opentofu
			daemon, _, ctx := setup(t, withDefaultEngine(tt.Engine))

			ws := daemon.createWorkspace(t, ctx, nil)
			_ = daemon.createAndUploadConfigurationVersion(t, ctx, ws, nil)

			// now we have a config version, start a run with the plan-and-apply
			// operation
			browser.New(t, ctx, func(page playwright.Page) {
				workspaceURL := daemon.URL(path.Get(ws.ID))
				startRun(t, page, workspaceURL, run.PlanAndApplyOperation, true)

				// now destroy resources with the destroy-all operation
				// go to workspace page
				_, err := page.Goto(workspaceURL)
				require.NoError(t, err)
				screenshot(t, page, "workspace_page")

				// navigate to workspace settings
				err = page.Locator(`//li[@id='menu-item-settings']/a`).Click()
				require.NoError(t, err)

				// navigate to workspace settings
				err = page.Locator(`//li[@id='menu-item-advanced']/a`).Click()
				require.NoError(t, err)

				// click 'queue destroy plan' button
				err = page.Locator(`//button[@id='queue-destroy-plan-button']`).Click()
				require.NoError(t, err)

				planWithOptionalApply(t, page, true)
			})
		})
	}
}
