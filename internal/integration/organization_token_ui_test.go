package integration

import (
	"regexp"
	"testing"

	"github.com/leg100/otf/internal/ui/paths"
	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/require"
)

// TestIntegration_OrganizationTokenUI demonstrates managing organization tokens via the UI.
func TestIntegration_OrganizationTokenUI(t *testing.T) {
	integrationTest(t)

	daemon, org, ctx := setup(t)

	browser.New(t, ctx, func(page playwright.Page) {
		// go to organization
		_, err := page.Goto(daemon.URL(paths.Organization(org.Name)))
		require.NoError(t, err)

		// go to organization settings
		err = page.Locator(`#menu-item-settings > a`).Click()
		require.NoError(t, err)

		// go to organization token page
		err = page.Locator(`#menu-item-token > a`).Click()
		require.NoError(t, err)

		screenshot(t, page, "org_token_new")

		// create new token
		err = page.Locator(`//button[text()='Create organization token']`).Click()
		require.NoError(t, err)
		screenshot(t, page, "org_token_created")

		// check for JWT in flash msg
		err = expect.Locator(page.GetByRole("alert")).ToHaveText(regexp.MustCompile(`Created token:\s+[\w-]+\.[\w-]+\.[\w-]+`))
		require.NoError(t, err)

		// check token exists and begins with `ot-`
		err = expect.Locator(page.Locator(`//*[@id='item-token']//span[@x-ref='content']`)).ToHaveText(regexp.MustCompile(`^ot-`))
		require.NoError(t, err)

		// capture token for comparison
		token, err := page.Locator(`//*[@id='item-token']//span[@x-ref='content']`).TextContent()
		require.NoError(t, err)

		// regenerate token
		err = page.Locator(`//button[text()='Regenerate']`).Click()
		require.NoError(t, err)

		// check regenerated token does not match original token
		err = expect.Locator(page.Locator(`//*[@id='item-token']//span[@x-ref='content']`)).Not().ToHaveText(token)
		require.NoError(t, err)

		// check regenerated token begins with `ot-`
		err = expect.Locator(page.Locator(`//*[@id='item-token']//span[@x-ref='content']`)).ToHaveText(regexp.MustCompile(`^ot-`))
		require.NoError(t, err)

		// delete token
		err = page.Locator(`//button[@id="delete-button"]`).Click()
		require.NoError(t, err)

		// flash msg declaring token is deleted
		err = expect.Locator(page.GetByRole("alert")).ToHaveText(`Deleted organization token`)
		require.NoError(t, err)
	})
}
