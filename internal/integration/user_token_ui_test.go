package integration

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// TestIntegration_UserTokenUI demonstrates managing user tokens via the UI.
func TestIntegration_UserTokenUI(t *testing.T) {
	integrationTest(t)

	svc, _, ctx := setup(t, nil)
	page := browser.New(t, ctx)
	// go to profile
	_, err := page.Goto("https://" + svc.System.Hostname() + "/app/profile")
	require.NoError(t, err)

	// go to user tokens
	err = page.Locator(`//div[@id='user-tokens-link']/a`).Click()
	require.NoError(t, err)

	////screenshot(t, "user_tokens"),
	// go to new token
	err = page.Locator(`//button[@id='new-user-token-button']`).Click()
	require.NoError(t, err)

	// enter description for new token and submit
	err = page.Locator("textarea#description").Fill("my new token")
	require.NoError(t, err)
	//screenshot(t, "user_token_enter_description"),

	err = page.Locator(`//button[text()='Create token']`).Click()
	require.NoError(t, err)

	//screenshot(t, "user_token_created"),
	err = expect.Locator(page.Locator("//div[@role='alert']")).ToHaveText(`Created token:\s+[\w-]+\.[\w-]+\.[\w-]+`)

	// delete the token
	err = page.Locator(`//button[text()='delete']`).Click()
	require.NoError(t, err)

	//screenshot(t),
	err = expect.Locator(page.Locator("//div[@role='alert']")).ToHaveText("Deleted token")
	require.NoError(t, err)
}
