package integration

import (
	"fmt"
	"testing"

	"github.com/leg100/otf/internal/ui/paths"
	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/require"
)

// TestIntegration_OrganizationUI demonstrates management of organizations via the UI.
func TestIntegration_OrganizationUI(t *testing.T) {
	integrationTest(t)

	daemon, _, ctx := setup(t, skipDefaultOrganization())

	t.Run("create", func(t *testing.T) {
		browser.New(t, ctx, func(page playwright.Page) {
			// go to the list of organizations
			_, err := page.Goto(daemon.URL(paths.Organizations()))
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
		})
	})

	t.Run("list", func(t *testing.T) {
		org1 := daemon.createOrganization(t, ctx)
		org2 := daemon.createOrganization(t, ctx)
		org3 := daemon.createOrganization(t, ctx)

		browser.New(t, ctx, func(page playwright.Page) {
			// go to the list of organizations
			_, err := page.Goto(daemon.URL(paths.Organizations()))
			require.NoError(t, err)

			// check three orgs are listed
			err = expect.Locator(page.Locator("tr#org-item-" + org1.Name.String())).ToBeVisible()
			require.NoError(t, err)
			err = expect.Locator(page.Locator("tr#org-item-" + org2.Name.String())).ToBeVisible()
			require.NoError(t, err)
			err = expect.Locator(page.Locator("tr#org-item-" + org3.Name.String())).ToBeVisible()
			require.NoError(t, err)

		})
	})

	t.Run("settings", func(t *testing.T) {
		org1 := daemon.createOrganization(t, ctx)

		browser.New(t, ctx, func(page playwright.Page) {
			// go to the list of organizations
			_, err := page.Goto(daemon.URL(paths.Organizations()))
			require.NoError(t, err)

			// go to organization
			selector := fmt.Sprintf("//tr[@id='org-item-%s']/td[1]/a", org1.Name)
			err = page.Locator(selector).Click()
			require.NoError(t, err)

			// go to organization's settings
			err = page.Locator("#menu-item-settings > a").Click()
			require.NoError(t, err)

			// change organization name
			err = page.Locator("input#name").Clear()
			require.NoError(t, err)
			err = page.Locator("input#name").Fill("super-duper-org")
			require.NoError(t, err)
			err = page.Locator("input#sentinel-version").Fill("0.40.0")
			require.NoError(t, err)

			err = page.Locator(`//button[text()='Update organization settings']`).Click()
			require.NoError(t, err)

			err = expect.Locator(page.GetByRole("alert")).ToHaveText("updated organization")
			require.NoError(t, err)
			err = expect.Locator(page.Locator("input#sentinel-version")).ToHaveValue("0.40.0")
			require.NoError(t, err)

			// go to advanced settings
			err = page.Locator("#menu-item-advanced > a").Click()
			require.NoError(t, err)

			// delete the organization
			err = page.Locator(`//button[@id='delete-organization-button']`).Click()
			require.NoError(t, err)

			err = expect.Locator(page.GetByRole("alert")).ToHaveText("deleted organization: super-duper-org")
			require.NoError(t, err)
		})
	})
}
