package integration

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestIntegration_OrganizationTokenUI demonstrates managing organization tokens via the UI.
func TestIntegration_OrganizationTokenUI(t *testing.T) {
	integrationTest(t)

	svc, org, ctx := setup(t, nil)

	page := browser.New(t, ctx)
	// go to organization
	_, err := page.Goto(organizationURL(svc.System.Hostname(), org.Name))
	require.NoError(t, err)
	// go to organization token page
	err = page.Locator(`//span[@id='organization_tokens']/a`).Click()
	require.NoError(t, err)
	//screenshot(t, "org_token_new"),
	// create new token
	err = page.Locator(`//button[text()='Create organization token']`).Click()
	require.NoError(t, err)
	//screenshot(t, "org_token_created"),

	// check for JWT in flash msg
	err = expect.Locator(page.Locator("//div[@role='alert']")).ToHaveText(regexp.MustCompile(`Created token:\s+[\w-]+\.[\w-]+\.[\w-]+`))
	require.NoError(t, err)

	// token widget should be visible
	err = expect.Locator(page.Locator(`//div[@class='widget']//span[text()='Token']`)).ToBeVisible()
	require.NoError(t, err)

	// check token begins with `ot-`
	err = expect.Locator(page.Locator(`//div[@class='widget']//span[@class='identifier']`)).ToHaveText(regexp.MustCompile(`^ot-`))
	require.NoError(t, err)

	// capture token for comparison
	token, err := page.Locator(`//div[@class='widget']//span[@class='identifier']`).TextContent()
	require.NoError(t, err)

	// regenerate token
	err = page.Locator(`//button[text()='regenerate']`).Click()
	require.NoError(t, err)

	// check regenerated token does not match original token
	err = expect.Locator(page.Locator(`//div[@class='widget']//span[@class='identifier']`)).Not().ToHaveText(token)
	require.NoError(t, err)

	// check regenerated token begins with `ot-`
	err = expect.Locator(page.Locator(`//div[@class='widget']//span[@class='identifier']`)).ToHaveText(regexp.MustCompile(`^ot-`))
	require.NoError(t, err)

	// delete token
	err = page.Locator(`//button[text()='delete']`).Click()
	require.NoError(t, err)

	// flash msg declaring token is deleted
	err = expect.Locator(page.Locator("//div[@role='alert']")).ToHaveText(`Deleted organization token`)
	require.NoError(t, err)
}
