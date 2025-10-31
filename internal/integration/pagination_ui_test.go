package integration

import (
	"testing"

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/require"
)

// TestIntegration_PaginationUI tests the pagination functionality on the UI.
func TestIntegration_PaginationUI(t *testing.T) {
	integrationTest(t)

	daemon, _, ctx := setup(t, skipDefaultOrganization())

	// create 101 orgs
	for range 101 {
		daemon.createOrganization(t, ctx)
	}

	browser.New(t, ctx, func(page playwright.Page) {
		// go to the list of organizations
		_, err := page.Goto("https://" + daemon.System.Hostname() + "/app/organizations")
		require.NoError(t, err)

		// should be 20 orgs listed on page one
		err = expect.Locator(page.Locator(`//table/tbody/tr`)).ToHaveCount(20)
		require.NoError(t, err)

		// expect accurate page info
		err = expect.Locator(page.Locator(`#page-info`)).ToHaveText("1-20 of 101")
		require.NoError(t, err)

		// set page size instead to 100
		oneHundredPerPage := []string{"100"}
		_, err = page.Locator(`#page-size-selector`).SelectOption(playwright.SelectOptionValues{
			Values: &oneHundredPerPage,
		})
		require.NoError(t, err)

		// should now be 100 orgs listed on page one
		err = expect.Locator(page.Locator(`//table//tbody/tr`)).ToHaveCount(100)
		require.NoError(t, err)

		// expect accurate page info
		err = expect.Locator(page.Locator(`#page-info`)).ToHaveText("1-100 of 101")
		require.NoError(t, err)

		// go to next page
		err = page.Locator("#next-page-link").Click()
		require.NoError(t, err)

		// should now be 1 org listed on page one
		err = expect.Locator(page.Locator(`//table//tbody/tr`)).ToHaveCount(1)
		require.NoError(t, err)

		// expect accurate page info
		err = expect.Locator(page.Locator(`#page-info`)).ToHaveText("101-101 of 101")
		require.NoError(t, err)
	})
}
