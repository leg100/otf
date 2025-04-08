package integration

import (
	"fmt"
	"testing"

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/require"
)

// TestIntegration_VariableSetUI tests management of variable sets via the UI.
func TestIntegration_VariableSetUI(t *testing.T) {
	integrationTest(t)

	svc, org, ctx := setup(t, nil)

	// Create global variable set in browser
	browser.New(t, ctx, func(page playwright.Page) {
		// go to org
		_, err := page.Goto(organizationURL(svc.System.Hostname(), org.Name))
		require.NoError(t, err)

		// go to variable sets
		err = page.Locator(`//a[text()='variable sets']`).Click()
		require.NoError(t, err)
		// click new variable set button
		err = page.Locator(`button#new-variable-set-button`).Click()
		require.NoError(t, err)

		// enter name
		err = page.Locator("input#name").Fill("global-1")
		require.NoError(t, err)

		// enter description
		err = page.Locator("textarea#description").Fill("this is my global variable set")
		require.NoError(t, err)

		// global radio button should be set by default
		err = expect.Locator(page.Locator(`input#global:checked`)).ToBeVisible()
		require.NoError(t, err)

		// submit form
		err = page.Locator(`//button[@id='save-variable-set-button']`).Click()
		require.NoError(t, err)

		// confirm variable set added
		err = expect.Locator(page.GetByRole("alert")).ToHaveText("added variable set: global-1")
		require.NoError(t, err)

		// add a variable
		err = page.Locator(`//button[@id='add-variable-button']`).Click()
		require.NoError(t, err)

		// enter key
		err = page.Locator("input#key").Fill("foo")
		require.NoError(t, err)

		// enter value
		err = page.Locator("textarea#value").Fill("bar")
		require.NoError(t, err)

		// select terraform variable category
		err = page.Locator("input#terraform").Click()
		require.NoError(t, err)
		// submit form
		err = page.Locator(`//button[@id='save-variable-button']`).Click()
		require.NoError(t, err)
		// confirm variable added
		err = expect.Locator(page.GetByRole("alert")).ToHaveText("added variable: foo")
		require.NoError(t, err)

		ws1 := svc.createWorkspace(t, ctx, org)
		ws2 := svc.createWorkspace(t, ctx, org)
		ws3 := svc.createWorkspace(t, ctx, org)

		// Create workspace-scoped variable set in browser, and add a variable.
		//
		// go to org
		_, err = page.Goto(organizationURL(svc.System.Hostname(), org.Name))
		require.NoError(t, err)

		// go to variable sets
		err = page.Locator(`//a[text()='variable sets']`).Click()
		require.NoError(t, err)

		// click new variable set button and wait for alpine to load on new
		// variable page
		err = page.Locator(`button#new-variable-set-button`).Click()
		require.NoError(t, err)

		// enter name
		err = page.Locator("input#name").Fill("workspace-scoped-1")
		require.NoError(t, err)

		// enter description
		err = page.Locator("textarea#description").Fill("variable set scoped to specific workspaces")
		require.NoError(t, err)

		// select workspace scope
		err = page.Locator(`input#workspace-scoped`).Click()
		require.NoError(t, err)

		// focus 'select workspace' input text box
		err = page.Locator(`input#workspace-input`).Click()
		require.NoError(t, err)

		// that should reveal dropdown menu of three workspaces
		err = expect.Locator(page.Locator(`//div[@x-ref='panel']`)).ToBeVisible()
		require.NoError(t, err)

		err = expect.Locator(page.Locator(fmt.Sprintf(`//div[@x-ref='panel']/button[text()='%s']`, ws1.Name))).ToBeVisible()
		require.NoError(t, err)

		err = expect.Locator(page.Locator(fmt.Sprintf(`//div[@x-ref='panel']/button[text()='%s']`, ws2.Name))).ToBeVisible()
		require.NoError(t, err)

		err = expect.Locator(page.Locator(fmt.Sprintf(`//div[@x-ref='panel']/button[text()='%s']`, ws3.Name))).ToBeVisible()
		require.NoError(t, err)

		// select ws1
		err = page.Locator(fmt.Sprintf(`//div[@x-ref='panel']/button[text()='%s']`, ws1.Name)).Click()
		require.NoError(t, err)

		// that should add ws1 to a list of workspaces
		err = expect.Locator(page.Locator(fmt.Sprintf(`//div[@id='existing-workspaces']//span[text()='%s']`, ws1.Name))).ToBeVisible()
		require.NoError(t, err)

		// submit form
		err = page.Locator(`//button[@id='save-variable-set-button']`).Click()
		require.NoError(t, err)

		// confirm variable set added
		err = expect.Locator(page.GetByRole("alert")).ToHaveText("added variable set: workspace-scoped-1")
		require.NoError(t, err)

		// list of workspaces should be persisted, and include ws1
		err = expect.Locator(page.Locator(`//div[@id='existing-workspaces']//span`)).ToHaveText(ws1.Name)
		require.NoError(t, err)

		// add a variable
		err = page.Locator(`//button[@id='add-variable-button']`).Click()
		require.NoError(t, err)

		// enter key
		err = page.Locator("input#key").Fill("foo")
		require.NoError(t, err)

		// enter value
		err = page.Locator("textarea#value").Fill("baz")
		require.NoError(t, err)

		// select terraform variable category
		err = page.Locator("input#terraform").Click()
		require.NoError(t, err)

		// submit form
		err = page.Locator(`//button[@id='save-variable-button']`).Click()
		require.NoError(t, err)

		// confirm variable added
		err = expect.Locator(page.GetByRole("alert")).ToHaveText("added variable: foo")
		require.NoError(t, err)

		// go to variables page for workspace ws1
		_, err = page.Goto(workspaceURL(svc.System.Hostname(), org.Name, ws1.Name))
		require.NoError(t, err)

		err = page.Locator(`//a[text()='variables']`).Click()
		require.NoError(t, err)

		// page should list 2 variable sets, one global, one
		// workspace-scoped
		err = expect.Locator(page.Locator(`//span[text()='Variable Sets (2)']`)).ToBeVisible()
		require.NoError(t, err)

		err = expect.Locator(page.Locator(`//*[@id='item-variable-set-global-1']`)).ToBeVisible()
		require.NoError(t, err)

		err = expect.Locator(page.Locator(`//*[@id='item-variable-set-workspace-scoped-1']`)).ToBeVisible()
		require.NoError(t, err)

		// both sets define a variable named 'foo', but the workspace-scoped
		// set takes precedence over the global set, so the latter's
		// variable should be tagged as 'overridden', and the variable name
		// should be struck-through
		err = expect.Locator(page.Locator(`//div[@id='variable-set-global-1']/div[@id='variable-set-variables-table']//td[1]/s[text()='foo']`)).ToBeVisible()
		require.NoError(t, err)

		err = expect.Locator(page.Locator(`//div[@id='variable-set-global-1']/div[@id='variable-set-variables-table']//td[1]/span[text()='OVERWRITTEN']`)).ToBeVisible()
		require.NoError(t, err)

		// whereas the workspace-scoped set should not be overwritten.
		err = expect.Locator(page.Locator(`//*[@id='variable-set-workspace-scoped-1']//td[1][text()='foo']`)).ToBeVisible()
		require.NoError(t, err)
	})
}
