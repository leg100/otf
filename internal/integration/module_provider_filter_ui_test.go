package integration

import (
	"testing"

	"github.com/leg100/otf/internal/module"
	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/require"
)

func TestModuleProviderFilterUI(t *testing.T) {
	integrationTest(t)
	svc, org, ctx := setup(t)

	// create umpteen modules
	mods := []struct {
		name     string
		provider string
	}{
		{"vpc", "aws"},
		{"vpc", "gcp"},
		{"vpc", "azure"},
		{"k8s", "aws"},
		{"k8s", "gcp"},
		{"k8s", "azure"},
	}
	for _, mod := range mods {
		_, err := svc.Modules.CreateModule(ctx, module.CreateOptions{
			Organization: org.Name,
			Name:         mod.name,
			Provider:     mod.provider,
		})
		require.NoError(t, err)
	}

	// check provider filtering capabilities on UI
	browser.New(t, ctx, func(page playwright.Page) {
		// go to org
		_, err := page.Goto(organizationURL(svc.System.Hostname(), org.Name))
		require.NoError(t, err)

		// go to modules
		err = page.Locator("#menu-item-modules > a").Click()
		require.NoError(t, err)

		// expect 6 modules
		err = expect.Locator(page.Locator(`#page-info`)).ToHaveText("1-6 of 6")
		require.NoError(t, err)

		// reveal provider filter
		err = page.Locator(`//input[@name='provider_filter_visible']`).Click()
		require.NoError(t, err)

		// filter by provider 'gcp'
		err = page.Locator(`//label[@for='filter-item-gcp']`).Click()
		require.NoError(t, err)

		// expect 2 modules
		err = expect.Locator(page.Locator(`#page-info`)).ToHaveText("1-2 of 2")
		require.NoError(t, err)

		err = expect.Locator(page.Locator(`//*[@id="mod-item-k8s"]/td[2]`)).ToHaveText(`gcp`)
		require.NoError(t, err)

		err = expect.Locator(page.Locator(`//*[@id="mod-item-vpc"]/td[2]`)).ToHaveText(`gcp`)
		require.NoError(t, err)

		// filter by provider 'aws' as well
		err = page.Locator(`//label[@for='filter-item-aws']`).Click()
		require.NoError(t, err)

		// expect 4 modules
		err = expect.Locator(page.Locator(`#page-info`)).ToHaveText("1-4 of 4")
		require.NoError(t, err)
	})
}
