package integration

import (
	"testing"

	"github.com/leg100/otf/internal/run"
	"github.com/stretchr/testify/require"
)

// TestStartRunUI tests starting a run via the Web UI before confirming and
// applying the run.
func TestStartRunUI(t *testing.T) {
	integrationTest(t)

	svc, _, ctx := setup(t, nil)

	ws := svc.createWorkspace(t, ctx, nil)
	_ = svc.createAndUploadConfigurationVersion(t, ctx, ws, nil)

	// now we have a config version, start a run with the plan-and-apply
	// operation
	page := browser.New(t, ctx)

	startRunTasks(t, page, svc.System.Hostname(), ws.Organization, ws.Name, run.PlanAndApplyOperation, true)

	// now destroy resources with the destroy-all operation
	// go to workspace page
	_, err := page.Goto(workspaceURL(svc.System.Hostname(), ws.Organization, ws.Name))
	require.NoError(t, err)
	//screenshot(t, "workspace_page"),
	// navigate to workspace settings
	err = page.Locator(`//a[text()='settings']`).Click()
	require.NoError(t, err)
	//screenshot(t),

	// click 'queue destroy plan' button
	err = page.Locator(`//button[@id='queue-destroy-plan-button']`).Click()
	require.NoError(t, err)
	//screenshot(t),

	planWithOptionalApply(t, page, svc.System.Hostname(), ws.Organization, ws.Name, true)
}
