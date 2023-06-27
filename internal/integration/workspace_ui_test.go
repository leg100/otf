package integration

import (
	"strings"
	"testing"

	"github.com/chromedp/cdproto/input"
	"github.com/chromedp/chromedp"
	"github.com/chromedp/chromedp/kb"
)

// TestIntegration_WorkspaceUI demonstrates management of workspaces via the UI.
func TestIntegration_WorkspaceUI(t *testing.T) {
	integrationTest(t)

	daemon, org, ctx := setup(t, nil)

	browser.Run(t, ctx, chromedp.Tasks{
		createWorkspace(t, daemon.Hostname(), org.Name, "workspace-1"),
		createWorkspace(t, daemon.Hostname(), org.Name, "workspace-12"),
		createWorkspace(t, daemon.Hostname(), org.Name, "workspace-2"),
		chromedp.Navigate(workspacesURL(daemon.Hostname(), org.Name)),
		// search for 'workspace-1' which should produce two results
		chromedp.Focus(`input[type="search"]`, chromedp.NodeVisible, chromedp.ByQuery),
		input.InsertText("workspace-1"),
		chromedp.Submit(`input[type="search"]`, chromedp.ByQuery),
		chromedp.WaitVisible(`//*[@class="item"]`, chromedp.AtLeast(2)),
		// and workspace-2 should not be visible
		chromedp.WaitNotPresent(`//*[@id="item-workspace-workspace-2"]`),
		// clear search term
		chromedp.SendKeys(`input[type="search"]`, strings.Repeat(kb.Delete, len("workspace-1")), chromedp.ByQuery),
		// now workspace-2 should be visible (updated via ajax)
		chromedp.WaitVisible(`//*[@id="item-workspace-workspace-2"]`),
	})
}
