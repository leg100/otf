package integration

import (
	"testing"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/input"
	"github.com/chromedp/chromedp"
	"github.com/stretchr/testify/assert"
)

// TestIntegration_OrganizationUI demonstrates management of organizations via the UI.
func TestIntegration_OrganizationUI(t *testing.T) {
	integrationTest(t)

	daemon, _, ctx := setup(t, &config{skipDefaultOrganization: true})

	// test creating/updating/deleting
	browser.Run(t, ctx, chromedp.Tasks{
		// go to the list of organizations
		chromedp.Navigate("https://" + daemon.System.Hostname() + "/app/organizations"),
		// add an org
		chromedp.Click("#new-organization-button", chromedp.ByQuery),
		chromedp.Focus("input#name", chromedp.NodeVisible, chromedp.ByQuery),
		input.InsertText("acme-corp"),
		screenshot(t, "new_org_enter_name"),
		chromedp.Submit("input#name", chromedp.ByQuery),
		screenshot(t, "new_org_created"),
		matchText(t, "//div[@role='alert']", "created organization: acme-corp"),
		// go to organization settings
		chromedp.Click("#settings > a", chromedp.ByQuery),
		screenshot(t),
		// change organization name
		chromedp.Focus("input#name", chromedp.NodeVisible, chromedp.ByQuery),
		chromedp.Clear("input#name", chromedp.ByQuery),
		input.InsertText("super-duper-org"),
		screenshot(t),
		chromedp.Click(`//button[text()='Update organization name']`),
		screenshot(t),
		matchText(t, "//div[@role='alert']", "updated organization"),
		// delete the organization
		chromedp.Click(`//button[@id='delete-organization-button']`),
		screenshot(t),
		matchText(t, "//div[@role='alert']", "deleted organization: super-duper-org"),
	})

	// test listing orgs...first create 101 orgs
	for i := 0; i < 101; i++ {
		daemon.createOrganization(t, ctx)
	}
	var (
		pageOneWidgets []*cdp.Node
		pageTwoWidgets []*cdp.Node
	)
	// open browser
	browser.Run(t, ctx, chromedp.Tasks{
		// go to the list of organizations
		chromedp.Navigate("https://" + daemon.System.Hostname() + "/app/organizations"),
		// should be 100 orgs listed on page one
		chromedp.Nodes(`.widget`, &pageOneWidgets, chromedp.NodeVisible),
		// go to page two
		chromedp.Click(`#next-page-link`, chromedp.ByQuery),
		// should be one org listed
		chromedp.Nodes(`.widget`, &pageTwoWidgets, chromedp.NodeVisible),
	})
	assert.Equal(t, 100, len(pageOneWidgets))
	assert.Equal(t, 1, len(pageTwoWidgets))
}
