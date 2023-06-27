package integration

import (
	"testing"

	"github.com/chromedp/chromedp"
	"github.com/stretchr/testify/require"
)

// TestAutoApply tests auto-apply functionality, using the UI to enable
// auto-apply on a workspace first before invoking 'terraform apply'.
func TestAutoApply(t *testing.T) {
	integrationTest(t)

	svc, org, ctx := setup(t, nil)

	// create workspace and enable auto-apply
	browser.Run(t, ctx, chromedp.Tasks{
		createWorkspace(t, svc.Hostname(), org.Name, t.Name()),
		chromedp.Tasks{
			// go to workspace
			chromedp.Navigate(workspaceURL(svc.Hostname(), org.Name, t.Name())),
			screenshot(t),
			// go to workspace settings
			chromedp.Click(`//a[text()='settings']`),
			screenshot(t),
			// enable auto-apply
			chromedp.Click("input#auto_apply", chromedp.ByQuery),
			screenshot(t),
			// submit form
			chromedp.Click(`//button[text()='Save changes']`),
			screenshot(t),
			// confirm workspace updated
			matchText(t, ".flash-success", "updated workspace", chromedp.ByQuery),
		},
	})

	// create terraform config
	configPath := newRootModule(t, svc.Hostname(), org.Name, t.Name())
	svc.tfcli(t, ctx, "init", configPath)
	// terraform apply - note we are not passing the -auto-approve flag yet we
	// expect it to auto-apply because the workspace is set to auto-apply.
	out := svc.tfcli(t, ctx, "apply", configPath)
	require.Contains(t, string(out), "Apply complete! Resources: 1 added, 0 changed, 0 destroyed.")
}
