package integration

import (
	"testing"

	"github.com/chromedp/cdproto/input"
	"github.com/chromedp/chromedp"
)

// TestIntegration_VCSProviderUI demonstrates management of vcs providers via
// the UI.
func TestIntegration_VCSProviderUI(t *testing.T) {
	integrationTest(t)

	daemon, org, ctx := setup(t, nil)

	browser.Run(t, ctx, chromedp.Tasks{
		// go to org
		chromedp.Navigate(organizationURL(daemon.Hostname(), org.Name)),
		screenshot(t, "organization_main_menu"),
		// go to vcs providers
		chromedp.Click("#vcs_providers > a", chromedp.ByQuery),
		screenshot(t, "vcs_providers_list"),
		// click 'New Github VCS Provider' button
		chromedp.Click(`//button[text()='New Github VCS Provider']`),
		screenshot(t, "new_github_vcs_provider_form"),
		// enter fake github token and name
		chromedp.Focus("textarea#token", chromedp.NodeVisible, chromedp.ByQuery),
		input.InsertText("fake-github-personal-token"),
		chromedp.Focus("input#name", chromedp.ByQuery, chromedp.NodeVisible),
		input.InsertText("my-token"),
		// submit form to create provider
		chromedp.Submit("textarea#token", chromedp.ByQuery),
		matchText(t, "//div[@role='alert']", "created provider: my-token"),
		// edit provider
		chromedp.Click(`//a[@id='edit-vcs-provider-link']`), waitLoaded,
		// change name
		chromedp.Focus("input#name", chromedp.ByQuery, chromedp.NodeVisible),
		chromedp.Clear("input#name", chromedp.ByQuery),
		input.InsertText("my-updated-name"),
		chromedp.Click(`//button[text()='Update']`),
		matchText(t, "//div[@role='alert']", "updated provider: my-updated-name"),
		// change token
		chromedp.Click(`//a[@id='edit-vcs-provider-link']`), waitLoaded,
		chromedp.Focus("input#token", chromedp.ByQuery, chromedp.NodeVisible),
		input.InsertText("my-updated-token"),
		chromedp.Click(`//button[text()='Update']`),
		matchText(t, "//div[@role='alert']", "updated provider: my-updated-name"),
		// delete token
		chromedp.Click(`//a[@id='edit-vcs-provider-link']`), waitLoaded,
		chromedp.Click(`//button[@id='delete-vcs-provider-button']`),
		matchText(t, "//div[@role='alert']", "deleted provider: my-updated-name"),
	})
}
