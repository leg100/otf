package integration

import (
	"fmt"
	"testing"

	"github.com/chromedp/cdproto/input"
	"github.com/chromedp/chromedp"
)

// TestIntegration_VariableSetUI tests management of variable sets via the UI.
func TestIntegration_VariableSetUI(t *testing.T) {
	integrationTest(t)

	svc, org, ctx := setup(t, nil)

	// Create global variable set in browser
	page := browser.New(t, ctx)
		chromedp.Tasks{
			// go to org
			_, err = page.Goto(organizationURL(svc.System.Hostname(), org.Name))
require.NoError(t, err)
			// go to variable sets
			err := page.Locator(`//a[text()='variable sets']`).Click()
require.NoError(t, err) waitLoaded,
			// click new variable set button
			err := page.Locator(`button#new-variable-set-button`).Click()
require.NoError(t, err) waitLoaded,
			// enter name
			chromedp.Focus("input#name", chromedp.NodeVisible, chromedp.ByQuery),
			input.InsertText("global-1"),
			// enter description
			chromedp.Focus("textarea#description", chromedp.NodeVisible, chromedp.ByQuery),
			input.InsertText("this is my global variable set"),
			// global radio button should be set by default
			chromedp.WaitVisible(`input#global:checked`, chromedp.ByQuery),
			// submit form
			err := page.Locator(`//button[@id='save-variable-set-button']`).Click()
require.NoError(t, err) waitLoaded,
			// confirm variable set added
			matchText(t, "//div[@role='alert']", "added variable set: global-1"),
			// add a variable
			err := page.Locator(`//button[@id='add-variable-button']`).Click()
require.NoError(t, err) waitLoaded,
			// enter key
			chromedp.Focus("input#key", chromedp.NodeVisible, chromedp.ByQuery),
			input.InsertText("foo"),
			// enter value
			chromedp.Focus("textarea#value", chromedp.NodeVisible, chromedp.ByQuery),
			input.InsertText("bar"),
			// select terraform variable category
			err := page.Locator("input#terraform").Click()
require.NoError(t, err)
			// submit form
			err := page.Locator(`//button[@id='save-variable-button']`).Click()
require.NoError(t, err) waitLoaded,
			// confirm variable added
			matchText(t, "//div[@role='alert']", "added variable: foo"),
		},
	})

	ws1 := svc.createWorkspace(t, ctx, org)
	ws2 := svc.createWorkspace(t, ctx, org)
	ws3 := svc.createWorkspace(t, ctx, org)

	// Create workspace-scoped variable set in browser, and add a variable.
	page := browser.New(t, ctx)
		chromedp.Tasks{
			// go to org
			_, err = page.Goto(organizationURL(svc.System.Hostname(), org.Name))
require.NoError(t, err)
			// go to variable sets
			err := page.Locator(`//a[text()='variable sets']`).Click()
require.NoError(t, err)
			// click new variable set button and wait for alpine to load on new
			// variable page
			err := page.Locator(`button#new-variable-set-button`).Click()
require.NoError(t, err)
			waitLoaded,
			// enter name
			chromedp.Focus("input#name", chromedp.NodeVisible, chromedp.ByQuery),
			input.InsertText("workspace-scoped-1"),
			// enter description
			chromedp.Focus("textarea#description", chromedp.NodeVisible, chromedp.ByQuery),
			input.InsertText("variable set scoped to specific workspaces"),
			// select workspace scope
			err := page.Locator(`input#workspace-scoped`).Click()
require.NoError(t, err)
			// focus 'select workspace' input text box
			err := page.Locator(`input#workspace-input`).Click()
require.NoError(t, err)
			// that should reveal dropdown menu of three workspaces
			chromedp.WaitVisible(`//div[@x-ref='panel']`),
			chromedp.WaitVisible(fmt.Sprintf(`//div[@x-ref='panel']/button[text()='%s']`, ws1.Name)),
			chromedp.WaitVisible(fmt.Sprintf(`//div[@x-ref='panel']/button[text()='%s']`, ws2.Name)),
			chromedp.WaitVisible(fmt.Sprintf(`//div[@x-ref='panel']/button[text()='%s']`, ws3.Name)),
			// select ws1
			err := page.Locator(fmt.Sprintf(`//div[@x-ref='panel']/button[text()='%s']`, ws1.Name)).Click()
			require.NoError(t, err)
			// that should add ws1 to a list of workspaces
			chromedp.WaitVisible(fmt.Sprintf(`//div[@id='existing-workspaces']//span[text()='%s']`, ws1.Name)),
			// submit form
			err := page.Locator(`//button[@id='save-variable-set-button']`).Click()
require.NoError(t, err)
			// confirm variable set added
			matchText(t, "//div[@role='alert']", "added variable set: workspace-scoped-1"),
			// list of workspaces should be persisted, and include ws1
			chromedp.WaitVisible(fmt.Sprintf(`//div[@id='existing-workspaces']//span[text()='%s']`, ws1.Name)),
			// add a variable
			err := page.Locator(`//button[@id='add-variable-button']`).Click()
require.NoError(t, err)
			// enter key
			chromedp.Focus("input#key", chromedp.NodeVisible, chromedp.ByQuery),
			input.InsertText("foo"),
			// enter value
			chromedp.Focus("textarea#value", chromedp.NodeVisible, chromedp.ByQuery),
			input.InsertText("baz"),
			// select terraform variable category
			err := page.Locator("input#terraform").Click()
require.NoError(t, err)
			// submit form
			err := page.Locator(`//button[@id='save-variable-button']`).Click()
require.NoError(t, err)
			// confirm variable added
			matchText(t, "//div[@role='alert']", "added variable: foo"),
			// go to variables page for workspace ws1
			_, err = page.Goto(workspaceURL(svc.System.Hostname(), org.Name, ws1.Name))
require.NoError(t, err)
			err := page.Locator(`//a[text()='variables']`).Click()
require.NoError(t, err)
			// page should list 2 variable sets, one global, one
			// workspace-scoped
			chromedp.WaitVisible(`//span[text()='Variable Sets (2)']`),
			chromedp.WaitVisible(`//div[@id='item-variable-set-global-1']`),
			chromedp.WaitVisible(`//div[@id='item-variable-set-workspace-scoped-1']`),
			// both sets define a variable named 'foo', but the workspace-scoped
			// set takes precedence over the global set, so the latter's
			// variable should be tagged as 'overridden', and the variable name
			// should be struck-through
			chromedp.WaitVisible(`//div[@id='variable-set-global-1']//td[1]/s/a[text()='foo']`),
			chromedp.WaitVisible(`//div[@id='variable-set-global-1']//td[1]/span[text()='OVERWRITTEN']`),
			// whereas the workspace-scoped set should not be overwritten.
			chromedp.WaitVisible(`//div[@id='variable-set-workspace-scoped-1']//td[1]/a[text()='foo']`),
		},
	})
}
