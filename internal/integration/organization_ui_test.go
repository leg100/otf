package integration

import (
	"testing"

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/require"
)

// TestIntegration_OrganizationUI demonstrates management of organizations via the UI.
func TestIntegration_OrganizationUI(t *testing.T) {
	integrationTest(t)

	daemon, _, ctx := setup(t, skipDefaultOrganization())

	// test creating/updating/deleting
	browser.New(t, ctx, func(page playwright.Page) {

		// go to the list of organizations
		_, err := page.Goto("https://" + daemon.System.Hostname() + "/app/organizations")
		require.NoError(t, err)

		// add an org
		err = page.Locator("#new-organization-button").Click()
		require.NoError(t, err)

		err = page.Locator("input#name").Fill("acme-corp")
		require.NoError(t, err)
		screenshot(t, page, "new_org_enter_name")

		err = page.Locator("input#name").Press("Enter")
		require.NoError(t, err)

		err = expect.Locator(page.GetByRole("alert")).ToHaveText(
			"created organization: acme-corp",
		)
		require.NoError(t, err)
		screenshot(t, page, "new_org_created")

		// go to the list of organizations
		_, err = page.Goto("https://" + daemon.System.Hostname() + "/app/organizations")
		require.NoError(t, err)

		// there should be one organization listed
		err = expect.Locator(page.Locator(`//table/tbody/tr`)).ToHaveCount(1)
		require.NoError(t, err)

		// go to organization
		err = page.Locator(`//tr[@id='org-item-acme-corp']/td[1]/a`).Click()
		require.NoError(t, err)

		// go to organization settings
		err = page.Locator("#menu-item-settings > a").Click()
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
	})
}
