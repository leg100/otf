package integration

import (
	"testing"

	"github.com/leg100/otf/internal/run"
	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/require"
)

// TestStartRunUI tests starting a run via the Web UI before confirming and
// applying the run.
func TestStartRunUI(t *testing.T) {
	integrationTest(t)

	for _, tt := range engineTestSpecs() {
		t.Run(tt.name, func(t *testing.T) {
			svc, _, ctx := setup(t, withDefaultEngine(tt.Engine))

			ws := svc.createWorkspace(t, ctx, nil)
			_ = svc.createAndUploadConfigurationVersion(t, ctx, ws, nil)

			// now we have a config version, start a run with the plan-and-apply
			// operation
			browser.New(t, ctx, func(page playwright.Page) {
				startRunTasks(t, page, svc.System.Hostname(), ws.Organization, ws.Name, run.PlanAndApplyOperation, true)

				// now destroy resources with the destroy-all operation
				// go to workspace page
				_, err := page.Goto(workspaceURL(svc.System.Hostname(), ws.Organization, ws.Name))
				require.NoError(t, err)
				screenshot(t, page, "workspace_page")
				// navigate to workspace settings
				err = page.Locator(`//li[@id='menu-item-settings']/a`).Click()
				require.NoError(t, err)

				// click 'queue destroy plan' button
				err = page.Locator(`//button[@id='queue-destroy-plan-button']`).Click()
				require.NoError(t, err)

				planWithOptionalApply(t, page, svc.System.Hostname(), true)
			})
		})
	}
}
