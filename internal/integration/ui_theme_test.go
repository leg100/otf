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

		// select dark theme
		err = page.Locator(`//*[@id='theme-chooser']`).Click()
		require.NoError(t, err)
		err = page.Locator(`//*[@id='theme-chooser']//button[@data-set-theme='dark']`).Click()
		require.NoError(t, err)

		// confirm light theme has been persisted to dark storage
		storage, err := page.Context().StorageState()
		require.NoError(t, err)
		for _, origin := range storage.Origins {
			for _, entry := range origin.LocalStorage {
				if entry.Name == "theme" {
					require.Equal(t, "dark", entry.Value)
				}
			}
		}
	})
}
