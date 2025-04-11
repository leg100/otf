package integration

import (
	"testing"

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/require"
)

func TestUITheme(t *testing.T) {
	integrationTest(t)
	daemon, _, ctx := setup(t, nil)

	browser.New(t, ctx, func(page playwright.Page) {
		// go to main page
		_, err := page.Goto("https://" + daemon.System.Hostname())
		require.NoError(t, err)
		// click on profile in menu
		err = page.Locator("#menu-item-profile > a").Click()
		require.NoError(t, err)
		// confirm light theme is the default
		err = expect.Locator(page.Locator(`//select[@id='theme-selector']`)).ToHaveValue("light")
		require.NoError(t, err)
		// change theme to dark
		selectValues := []string{"dark"}
		_, err = page.Locator(`//select[@id='theme-selector']`).SelectOption(playwright.SelectOptionValues{
			Values: &selectValues,
		})
		require.NoError(t, err)
		// reload page
		_, err = page.Reload()
		require.NoError(t, err)
		// confirm theme change has been persisted by confirming dark theme is
		// now the selected option
		err = expect.Locator(page.Locator(`//select[@id='theme-selector']`)).ToHaveValue("dark")
		require.NoError(t, err)
	})
}
