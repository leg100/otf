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
	browser.Run(t, ctx, chromedp.Tasks{
		// go to organization
		chromedp.Navigate(organizationURL(svc.Hostname(), org.Name)),
		// go to organization token page
		chromedp.Click(`//span[@id='organization_tokens']/a`),
		screenshot(t, "org_token_new"),
		// create new token
		chromedp.Click(`//button[text()='Create organization token']`),
		screenshot(t, "org_token_created"),
		// check for JWT in flash msg
		matchRegex(t, "//div[@role='alert']", `Created token:\s+[\w-]+\.[\w-]+\.[\w-]+`),
		// token widget should be visible
		chromedp.WaitVisible(`//div[@class='widget']//span[text()='Token']`),
		chromedp.Text(`//div[@class='widget']//span[@class='identifier']`, &createdTokenID),
		// regenerate token
		chromedp.Click(`//button[text()='regenerate']`),
		chromedp.Text(`//div[@class='widget']//span[@class='identifier']`, &regeneratedTokenID),
		// delete token
		chromedp.Click(`//button[text()='delete']`),
		// flash msg declaring token is deleted
		matchText(t, "//div[@role='alert']", `Deleted organization token`),
	})
	assert.True(t, strings.HasPrefix(createdTokenID, "ot-"))
	assert.True(t, strings.HasPrefix(regeneratedTokenID, "ot-"))
	assert.NotEqual(t, createdTokenID, regeneratedTokenID)
}
