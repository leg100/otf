package integration

import (
	"testing"

	"github.com/chromedp/chromedp"
	"github.com/stretchr/testify/require"
)

// TestAutoApply demonstrates enabling auto-apply via the UI.
func TestAutoApply(t *testing.T) {
	integrationTest(t)

	svc, org, ctx := setup(t, nil)

	// create workspace and enable auto-apply
	browser.Run(t, ctx, chromedp.Tasks{
		createWorkspace(t, svc.System.Hostname(), org.Name, t.Name()),
		chromedp.Tasks{
			// go to workspace
			chromedp.Navigate(workspaceURL(svc.System.Hostname(), org.Name, t.Name())),
			screenshot(t),
			// go to workspace settings
			chromedp.Click(`//a[text()='settings']`),
			screenshot(t),
			// enable auto-apply
			chromedp.Click(`//input[@name='auto_apply' and @value='true']`),
			screenshot(t),
			// submit form
			chromedp.Click(`//button[text()='Save changes']`),
			screenshot(t),
			// confirm workspace updated
			matchText(t, "//div[@role='alert']", "updated workspace"),
		},
	})
	// check UI has correctly updated the workspace resource
	ws, err := svc.Workspaces.GetWorkspaceByName(ctx, org.Name, t.Name())
	require.NoError(t, err)
	require.Equal(t, true, ws.AutoApply)
}
