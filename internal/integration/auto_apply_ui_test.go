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
	page := browser.New(t, ctx)
		createWorkspace(t, svc.System.Hostname(), org.Name, t.Name()),
		chromedp.Tasks{
			// go to workspace
			_, err = page.Goto(workspaceURL(svc.System.Hostname(), org.Name, t.Name()))
require.NoError(t, err)
			//screenshot(t),
			// go to workspace settings
			err := page.Locator(`//a[text()='settings']`).Click()
require.NoError(t, err)
			//screenshot(t),
			// enable auto-apply
			err := page.Locator(`//input[@name='auto_apply' and @value='true']`).Click()
require.NoError(t, err)
			//screenshot(t),
			// submit form
			err := page.Locator(`//button[text()='Save changes']`).Click()
require.NoError(t, err)
			//screenshot(t),
			// confirm workspace updated
			matchText(t, "//div[@role='alert']", "updated workspace"),
		},
	})
	// check UI has correctly updated the workspace resource
	ws, err := svc.Workspaces.GetByName(ctx, org.Name, t.Name())
	require.NoError(t, err)
	require.Equal(t, true, ws.AutoApply)
}
