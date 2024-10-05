package integration

import (
	"strings"
	"testing"

	"github.com/chromedp/chromedp"
	"github.com/stretchr/testify/assert"
)

// TestIntegration_OrganizationTokenUI demonstrates managing organization tokens via the UI.
func TestIntegration_OrganizationTokenUI(t *testing.T) {
	integrationTest(t)

	svc, org, ctx := setup(t, nil)

	var (
		createdTokenID     string
		regeneratedTokenID string
	)
	page := browser.New(t, ctx)
		// go to organization
		_, err = page.Goto(organizationURL(svc.System.Hostname(), org.Name))
require.NoError(t, err)
		// go to organization token page
		err := page.Locator(`//span[@id='organization_tokens']/a`).Click()
require.NoError(t, err)
		//screenshot(t, "org_token_new"),
		// create new token
		err := page.Locator(`//button[text()='Create organization token']`).Click()
require.NoError(t, err)
		//screenshot(t, "org_token_created"),
		// check for JWT in flash msg
		matchRegex(t, "//div[@role='alert']", `Created token:\s+[\w-]+\.[\w-]+\.[\w-]+`),
		// token widget should be visible
		chromedp.WaitVisible(`//div[@class='widget']//span[text()='Token']`),
		chromedp.Text(`//div[@class='widget']//span[@class='identifier']`, &createdTokenID),
		// regenerate token
		err := page.Locator(`//button[text()='regenerate']`).Click()
require.NoError(t, err)
		chromedp.Text(`//div[@class='widget']//span[@class='identifier']`, &regeneratedTokenID),
		// delete token
		err := page.Locator(`//button[text()='delete']`).Click()
require.NoError(t, err)
		// flash msg declaring token is deleted
		matchText(t, "//div[@role='alert']", `Deleted organization token`),
	})
	assert.True(t, strings.HasPrefix(createdTokenID, "ot-"))
	assert.True(t, strings.HasPrefix(regeneratedTokenID, "ot-"))
	assert.NotEqual(t, createdTokenID, regeneratedTokenID)
}
