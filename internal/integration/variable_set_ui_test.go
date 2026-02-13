package integration

import (
	"fmt"
	"testing"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/variable"
	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/require"
)

// TestIntegration_VariableSetUI tests adding a variable set in the browser.
func TestIntegration_VariableSetUI_New(t *testing.T) {
	integrationTest(t)

	svc, org, ctx := setup(t)

	// Create global variable set in browser
	browser.New(t, ctx, func(page playwright.Page) {
		// go to org
		_, err := page.Goto(organizationURL(svc.System.Hostname(), org.Name))
		require.NoError(t, err)

		// go to variable sets
		err = page.Locator(`#menu-item-variable-sets > a`).Click()
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
	})
}

// TestIntegration_VariableSetUI_Edit tests editing a variable set via the UI.
func TestIntegration_VariableSetUI_Edit(t *testing.T) {
	integrationTest(t)

	svc, org, ctx := setup(t)

	set, err := svc.Variables.CreateVariableSet(ctx, org.Name, variable.CreateVariableSetOptions{
		Name:   "global-1",
		Global: true,
	})
	require.NoError(t, err)

	// Edit global variable set in browser
	browser.New(t, ctx, func(page playwright.Page) {
		// go to variable set's page
		setURL := "https://" + svc.System.Hostname() + "/app/variable-sets/" + set.ID.String() + "/edit"
		_, err := page.Goto(setURL)
		require.NoError(t, err)

		// edit description
		err = page.Locator("textarea#description").Fill("this is my newly updated global variable set")
		require.NoError(t, err)

		// submit form
		err = page.Locator(`//button[@id='save-variable-set-button']`).Click()
		require.NoError(t, err)

		// confirm variable set updated
		err = expect.Locator(page.GetByRole("alert")).ToHaveText("updated variable set: global-1")
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
	})
}

// TestIntegration_VariableSetUI_Delete tests deleting a variable set variable via the UI.
func TestIntegration_VariableSetUI_VariableDelete(t *testing.T) {
	integrationTest(t)

	svc, org, ctx := setup(t)

	set, err := svc.Variables.CreateVariableSet(ctx, org.Name, variable.CreateVariableSetOptions{
		Name:   "global-1",
		Global: true,
	})
	require.NoError(t, err)

	_, err = svc.Variables.CreateVariableSetVariable(ctx, set.ID, variable.CreateVariableOptions{
		Key:      new("varset-var-1"),
		Value:    new("foo"),
		Category: internal.Ptr(variable.CategoryTerraform),
	})
	require.NoError(t, err)

	// Delete variable set variable set in browser
	browser.New(t, ctx, func(page playwright.Page) {
		// go to variable set's page
		setURL := "https://" + svc.System.Hostname() + "/app/variable-sets/" + set.ID.String() + "/edit"
		_, err := page.Goto(setURL)
		require.NoError(t, err)

		// delete variable
		err = page.Locator(`//tr[@id='item-variable-varset-var-1']//button[@id='delete-button']`).Click()
		require.NoError(t, err)

		// confirm variable set variable deleted
		err = expect.Locator(page.GetByRole("alert")).ToHaveText("deleted variable: varset-var-1")
		require.NoError(t, err)
	})
}

// TestIntegration_VariableSetUI_NewEdit_WorkspaceScoped tests creating a
// workspace-scoped variable set via the UI.
func TestIntegration_VariableSetUI_New_WorkspaceScoped(t *testing.T) {
	integrationTest(t)

	svc, org, ctx := setup(t)
	ws1 := svc.createWorkspace(t, ctx, org)
	ws2 := svc.createWorkspace(t, ctx, org)
	ws3 := svc.createWorkspace(t, ctx, org)

	// Create workspace-scoped variable set in browser, and add a variable.
	browser.New(t, ctx, func(page playwright.Page) {
		// go to org
		_, err := page.Goto(organizationURL(svc.System.Hostname(), org.Name))
		require.NoError(t, err)

		// go to variable sets
		err = page.Locator(`#menu-item-variable-sets > a`).Click()
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
	})
}

// TestIntegration_VariableSetUI_WorkspaceVariables tests the visibility and
// precedence of variable set variables for a workspace via the UI.
func TestIntegration_VariableSetUI_WorkspaceVariables(t *testing.T) {
	integrationTest(t)

	svc, org, ctx := setup(t)

	ws1 := svc.createWorkspace(t, ctx, org)

	// create global set
	globalSet, err := svc.Variables.CreateVariableSet(ctx, org.Name, variable.CreateVariableSetOptions{
		Name:   "global-1",
		Global: true,
	})
	require.NoError(t, err)

	// create variable for global set
	_, err = svc.Variables.CreateVariableSetVariable(ctx, globalSet.ID, variable.CreateVariableOptions{
		Key:      new("foo"),
		Value:    new("bar"),
		Category: internal.Ptr(variable.CategoryTerraform),
	})
	require.NoError(t, err)

	// create workspace-scoped set
	workspaceScopedSet, err := svc.Variables.CreateVariableSet(ctx, org.Name, variable.CreateVariableSetOptions{
		Name:       "workspace-scoped-1",
		Workspaces: []resource.TfeID{ws1.ID},
	})
	require.NoError(t, err)

	// create variable for workspace-scoped set
	_, err = svc.Variables.CreateVariableSetVariable(ctx, workspaceScopedSet.ID, variable.CreateVariableOptions{
		Key:      new("foo"),
		Value:    new("bar"),
		Category: internal.Ptr(variable.CategoryTerraform),
	})
	require.NoError(t, err)

	browser.New(t, ctx, func(page playwright.Page) {
		// go to variables page for workspace ws1
		_, err = page.Goto(workspaceURL(svc.System.Hostname(), org.Name, ws1.Name))
		require.NoError(t, err)

		err = page.Locator(`//li[@id='menu-item-variables']/a`).Click()
		require.NoError(t, err)

		// both sets define a variable named 'foo', but the workspace-scoped
		// set takes precedence over the global set, so the latter's
		// variable should be tagged as 'overridden', and the variable name
		// should be struck-through
		err = expect.Locator(page.Locator(`//*[@id='variables-table']//tbody/tr[1]/td[1]`)).ToHaveText(`global-1`)
		require.NoError(t, err)
		err = expect.Locator(page.Locator(`//*[@id='variables-table']//tbody/tr[1]/td[2]/s`)).ToHaveText(`foo`)
		require.NoError(t, err)
		err = expect.Locator(page.Locator(`//*[@id='variables-table']//tbody/tr[1]/td[2]/span`)).ToHaveText("overwritten")
		require.NoError(t, err)

		// whereas the workspace-scoped set should not be overwritten.
		err = expect.Locator(page.Locator(`//*[@id='variables-table']//tbody/tr[2]/td[1]`)).ToHaveText(`workspace-scoped-1`)
		require.NoError(t, err)
		err = expect.Locator(page.Locator(`//*[@id='variables-table']//tbody/tr[2]/td[2]`)).ToHaveText(`foo`)
		require.NoError(t, err)

		// click through to the variable set's edit page
		err = page.Locator(`//*[@id='variables-table']//tbody/tr[1]/td[1]/a`).Click()
		require.NoError(t, err)

		// expect variable set's edit page to load with title
		err = expect.Locator(page.Locator(`//*[@id="content"]/span`)).ToHaveText(`Edit variable set`)
		require.NoError(t, err)

		// go back to workspace variables page
		_, err = page.GoBack()
		require.NoError(t, err)

		// click through to the variable set variable's edit page
		err = page.Locator(`//*[@id='variables-table']//tr[1]//button[@id='edit-button']`).Click()
		require.NoError(t, err)

		// expect variable set variable's edit page to load with title
		err = expect.Locator(page.Locator(`//*[@id="content"]/span`)).ToHaveText(`Edit variable`)
		require.NoError(t, err)
	})
}
