package integration

import (
	"os"
	"path/filepath"
	"regexp"
	"testing"
	"time"

	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/run"
	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/require"
)

// createWorkspace creates a workspace via the UI
func createWorkspace(t *testing.T, page playwright.Page, hostname string, org organization.Name, name string) {
	t.Helper()

	_, err := page.Goto(organizationURL(hostname, org))
	require.NoError(t, err)

	err = page.Locator("#menu-item-workspaces > a").Click()
	require.NoError(t, err)

	err = page.Locator("#new-workspace-button").Click()
	require.NoError(t, err)

	err = page.Locator("input#name").Fill(name)
	require.NoError(t, err)

	err = page.Locator("#create-workspace-button").Click()
	require.NoError(t, err)

	err = expect.Locator(page.GetByRole("alert")).ToHaveText("created workspace: " + name)
	require.NoError(t, err)
}

// screenshot takes a screenshot of a browser and saves it to disk
func screenshot(t *testing.T, page playwright.Page, fname string) {
	t.Helper()

	// disable screenshots if headless mode is disabled - screenshots are
	// most likely unnecessary if the developer is using headless mode to
	// view the browser; and, depending on the developer's monitor, the
	// viewport in the screenshots is different to that when headless mode
	// is enabled, but we want the viewport to be consistent because
	// screenshots are also used in the documentation!
	if headless, ok := os.LookupEnv("OTF_E2E_HEADLESS"); ok {
		if headless == "false" {
			return
		}
	}

	// disable screenshots unless an environment variable is defined
	if _, ok := os.LookupEnv("OTF_DOC_SCREENSHOTS"); !ok {
		return
	}

	path := filepath.Join("..", "..", "docs", "docs", "images", fname+".png")
	err := os.MkdirAll(filepath.Dir(path), 0o755)
	require.NoError(t, err)

	_, err = page.Screenshot(playwright.PageScreenshotOptions{
		Path: &path,
	})
	require.NoError(t, err)
}

// addWorkspacePermission adds a workspace permission via the UI, assigning
// a role to a team.
func addWorkspacePermission(t *testing.T, page playwright.Page, hostname string, org organization.Name, workspaceName string, teamID resource.TfeID, role string) {
	t.Helper()

	// go to workspace
	_, err := page.Goto(workspaceURL(hostname, org, workspaceName))
	require.NoError(t, err)

	// go to workspace settings
	err = page.Locator(`//li[@id='menu-item-settings']/a`).Click()
	require.NoError(t, err)

	// confirm builtin admin role for owners team
	err = expect.Locator(page.Locator("#permissions-owners td:first-child")).ToHaveText("owners")
	require.NoError(t, err)

	err = expect.Locator(page.Locator("#permissions-owners td:last-child")).ToHaveText("admin")
	require.NoError(t, err)

	// assign role to team
	selectValues := []string{string(role)}
	_, err = page.Locator(`//select[@id="permissions-add-select-role"]`).SelectOption(playwright.SelectOptionValues{
		Values: &selectValues,
	})
	require.NoError(t, err)

	selectValues = []string{teamID.String()}
	_, err = page.Locator(`//select[@id="permissions-add-select-team"]`).SelectOption(playwright.SelectOptionValues{
		Values: &selectValues,
	})
	require.NoError(t, err)

	// scroll to bottom so that permissions are visible in screenshot
	err = page.Locator("#permissions-add-button").ScrollIntoViewIfNeeded()
	require.NoError(t, err)
	screenshot(t, page, "workspace_permissions")

	err = page.Locator("#permissions-add-button").Click()
	require.NoError(t, err)

	err = expect.Locator(page.GetByRole("alert")).ToHaveText("updated workspace permissions")
	require.NoError(t, err)
}

// startRunTasks starts a run via the UI
func startRunTasks(t *testing.T, page playwright.Page, hostname string, organization organization.Name, workspaceName string, op run.Operation, apply bool) {
	t.Helper()

	// go to workspace page
	_, err := page.Goto(workspaceURL(hostname, organization, workspaceName))
	require.NoError(t, err)

	screenshot(t, page, "connected_workspace_main_page")

	// select operation for run
	selectValues := []string{string(op)}
	_, err = page.Locator(`//select[@id="start-run-operation"]`).SelectOption(playwright.SelectOptionValues{
		Values: &selectValues,
	})
	require.NoError(t, err)

	// wait for page to transition then take screenshot of run page.
	page.WaitForLoadState()
	screenshot(t, page, "run_page_started")

	planWithOptionalApply(t, page, hostname, apply)
}

func planWithOptionalApply(t *testing.T, page playwright.Page, hostname string, apply bool) {
	t.Helper()

	// confirm plan begins and ends
	err := expect.Locator(page.Locator(`//*[@id='tailed-plan-logs']`)).ToContainText("Initializing the backend")
	require.NoError(t, err)

	err = expect.Locator(page.Locator(`//span[@id='plan-status']`)).ToHaveText("finished")
	require.NoError(t, err)

	// wait for run to enter planned or planned and finished state
	err = expect.Locator(page.Locator(`//span[contains(concat(' ', normalize-space(@class), ' '), ' run-status ')]/a`)).ToHaveText(regexp.MustCompile(`planned|planned and finished`))
	require.NoError(t, err)

	// run widget should show plan summary
	err = expect.Locator(page.Locator(`//div[@id='resource-summary']/span[1]`)).ToHaveText(regexp.MustCompile(`\+[0-9]+`))
	require.NoError(t, err)

	err = expect.Locator(page.Locator(`//div[@id='resource-summary']/span[2]`)).ToHaveText(regexp.MustCompile(`\~[0-9]+`))
	require.NoError(t, err)

	err = expect.Locator(page.Locator(`//div[@id='resource-summary']/span[3]`)).ToHaveText(regexp.MustCompile(`\-[0-9]+`))
	require.NoError(t, err)

	screenshot(t, page, "run_page_planned_state")

	if !apply {
		// not applying, nothing more to be done.
		return
	}

	// run widget should show discard button if run is applyable
	err = expect.Locator(page.Locator(`//button[@id='run-discard-button']`)).ToBeVisible()
	require.NoError(t, err)

	// click 'apply' button
	opts := playwright.PageGetByRoleOptions{Name: "apply"}
	err = page.GetByRole("button", opts).First().Click()
	require.NoError(t, err)

	// confirm apply begins and ends
	expect.Locator(page.Locator(`//*[@id='tailed-apply-logs']`)).ToContainText("Initializing the backend")
	err = expect.Locator(page.Locator(`//span[@id='apply-status']`)).ToHaveText("finished")
	require.NoError(t, err)

	// confirm run ends in applied state
	err = expect.Locator(page.Locator(`//span[contains(concat(' ', normalize-space(@class), ' '), ' run-status ')]/a`)).ToHaveText("applied")
	require.NoError(t, err)

	// run widget should show apply summary
	err = expect.Locator(page.Locator(`//div[@id='resource-summary']/span[1]`)).ToHaveText(regexp.MustCompile(`\+[0-9]+`))
	require.NoError(t, err)

	err = expect.Locator(page.Locator(`//div[@id='resource-summary']/span[2]`)).ToHaveText(regexp.MustCompile(`\~[0-9]+`))
	require.NoError(t, err)

	err = expect.Locator(page.Locator(`//div[@id='resource-summary']/span[3]`)).ToHaveText(regexp.MustCompile(`\-[0-9]+`))
	require.NoError(t, err)

	// because run was triggered from the UI, the UI icon should be visible.
	err = expect.Locator(page.Locator(`//*[@id='run-trigger-ui']`)).ToBeVisible()
	require.NoError(t, err)

	// run should show elapsed time
	err = expect.Locator(page.Locator(`//div[@id='elapsed-time']/span`)).ToHaveText(regexp.MustCompile(`\d+(s|ms)`))
	require.NoError(t, err)

	// plan should show running time
	err = expect.Locator(page.Locator(`//span[@id='running-time-plan']`)).ToHaveText(regexp.MustCompile(`\d+(s|ms)`))
	require.NoError(t, err)

	// apply should show running time
	err = expect.Locator(page.Locator(`//span[@id='running-time-apply']`)).ToHaveText(regexp.MustCompile(`\d+(s|ms)`))
	require.NoError(t, err)
}

func connectWorkspaceTasks(t *testing.T, page playwright.Page, hostname string, org organization.Name, name, provider string) {
	t.Helper()

	// go to workspace
	_, err := page.Goto(workspaceURL(hostname, org, name))
	require.NoError(t, err)
	screenshot(t, page, "workspace_main_page")

	// navigate to workspace settings
	err = page.Locator(`//li[@id='menu-item-settings']/a`).Click()
	require.NoError(t, err)
	screenshot(t, page, "workspace_settings")

	// click connect button
	err = page.Locator(`//button[@id='list-workspace-vcs-providers-button']`).Click()
	require.NoError(t, err)
	screenshot(t, page, "workspace_vcs_providers_list")

	// select provider
	err = page.Locator(`//tr[@id='item-vcsprovider-` + provider + `']//button[text()='Select']`).Click()
	require.NoError(t, err)
	screenshot(t, page, "workspace_vcs_repo_list")

	// connect to first repo in list (there should only be one)
	err = page.Locator(`//tbody//tr[1]//button[text()='Connect']`).Click()
	require.NoError(t, err)

	// confirm connected
	err = expect.Locator(page.GetByRole("alert")).ToHaveText("connected workspace to repo")
	require.NoError(t, err)
}

func disconnectWorkspaceTasks(t *testing.T, page playwright.Page, hostname string, org organization.Name, name string) {
	t.Helper()

	// go to workspace
	_, err := page.Goto(workspaceURL(hostname, org, name))
	require.NoError(t, err)

	// navigate to workspace settings
	err = page.Locator(`//li[@id='menu-item-settings']/a`).Click()
	require.NoError(t, err)

	// click disconnect button
	err = page.Locator(`//button[@id='disconnect-workspace-repo-button']`).Click()
	require.NoError(t, err)

	// confirm disconnected
	err = expect.Locator(page.GetByRole("alert")).ToHaveText("disconnected workspace from repo")
	require.NoError(t, err)
}

func reloadUntilEnabled(t *testing.T, page playwright.Page, sel string) {
	t.Helper()

	for range 10 {
		visible, err := page.Locator(sel).IsEnabled()
		require.NoError(t, err)

		if visible {
			return
		}

		time.Sleep(time.Second)

		_, err = page.Reload()
		require.NoError(t, err)
	}
	t.Fatalf("timed out waiting for dom node %s to be visible", sel)
}
