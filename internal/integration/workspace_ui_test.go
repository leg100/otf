package integration

import (
	"context"
	"strings"
	"testing"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/input"
	"github.com/chromedp/chromedp"
	"github.com/chromedp/chromedp/kb"
	"github.com/stretchr/testify/assert"
)

// TestIntegration_WorkspaceUI demonstrates management of workspaces via the UI.
func TestIntegration_WorkspaceUI(t *testing.T) {
	t.Parallel()

	daemon, org, ctx := setup(t, nil)

	var workspaceItems []*cdp.Node
	browser.Run(t, ctx, chromedp.Tasks{
		createWorkspace(t, daemon.Hostname(), org.Name, "workspace-1"),
		createWorkspace(t, daemon.Hostname(), org.Name, "workspace-12"),
		createWorkspace(t, daemon.Hostname(), org.Name, "workspace-2"),
		chromedp.Navigate(workspacesURL(daemon.Hostname(), org.Name)),
		chromedp.WaitReady(`body`),
		// search for 'workspace-1' which should produce two results
		chromedp.Focus(`input[type="search"]`, chromedp.NodeVisible),
		input.InsertText("workspace-1"),
		chromedp.Submit(`input[type="search"]`),
		chromedp.WaitReady(`body`),
		chromedp.Nodes(`//*[@class="item"]`, &workspaceItems, chromedp.BySearch),
		chromedp.ActionFunc(func(c context.Context) error {
			assert.Equal(t, 2, len(workspaceItems))
			return nil
		}),
		// and workspace-2 should not be visible
		chromedp.WaitNotPresent(`//*[@id="item-workspace-workspace-2"]`, chromedp.BySearch),
		// clear search term
		chromedp.SendKeys(`input[type="search"]`, strings.Repeat(kb.Delete, len("workspace-1")), chromedp.BySearch),
		// now workspace-2 should be visible (updated via ajax)
		chromedp.WaitVisible(`//*[@id="item-workspace-workspace-2"]`, chromedp.BySearch),
	})
}
