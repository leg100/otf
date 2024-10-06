package integration

import (
	"testing"

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/require"
)

// TestIntegration_OrganizationUI demonstrates management of organizations via the UI.
func TestIntegration_OrganizationUI(t *testing.T) {
	integrationTest(t)

	daemon, _, ctx := setup(t, &config{skipDefaultOrganization: true})

	// test creating/updating/deleting
	page := browser.New(t, ctx)

	// go to the list of organizations
	_, err := page.Goto("https://" + daemon.System.Hostname() + "/app/organizations")
	require.NoError(t, err)

	// add an org
	err = page.Locator("#new-organization-button").Click()
	require.NoError(t, err)

	err = page.Locator("input#name").Fill("acme-corp")
	require.NoError(t, err)
	screenshot(t, page, "new_org_enter_name")
	// screenshot(t, "new_org_created"),

	err = page.Locator("input#name").Press("Enter")
	require.NoError(t, err)

	err = page.GetByRole("alert").Filter(playwright.LocatorFilterOptions{
		HasText: "created organization: acme-corp",
	}).Click()
	require.NoError(t, err)

	// go to organization settings
	err = page.Locator("#settings > a").Click()
	require.NoError(t, err)

	// change organization name
	err = page.Locator("input#name").Clear()
	require.NoError(t, err)
	err = page.Locator("input#name").Fill("super-duper-org")
	require.NoError(t, err)

	err = page.Locator(`//button[text()='Update organization name']`).Click()
	require.NoError(t, err)

	err = expect.Locator(page.GetByRole("alert")).ToHaveText("updated organization")
	require.NoError(t, err)

	// delete the organization
	err = page.Locator(`//button[@id='delete-organization-button']`).Click()
	require.NoError(t, err)

	err = expect.Locator(page.GetByRole("alert")).ToHaveText("deleted organization: super-duper-org")
	require.NoError(t, err)

	// test listing orgs...first create 101 orgs
	for i := 0; i < 101; i++ {
		daemon.createOrganization(t, ctx)
	}

	// go to the list of organizations
	_, err = page.Goto("https://" + daemon.System.Hostname() + "/app/organizations")
	require.NoError(t, err)

	// should be 100 orgs listed on page one
	err = expect.Locator(page.Locator(`.widget`)).ToHaveCount(100)
	require.NoError(t, err)

	// go to page two
	err = page.Locator(`#next-page-link`).Click()
	require.NoError(t, err)

	// should be one org listed
	expect.Locator(page.Locator(`.widget`)).ToHaveCount(1)
}
