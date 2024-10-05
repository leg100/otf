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
	page := browser.New(t, ctx)
		// go to the list of organizations
		_, err = page.Goto("https://" + daemon.System.Hostname() + "/app/organizations")
require.NoError(t, err)
		// add an org
		err := page.Locator("#new-organization-button").Click()
require.NoError(t, err)
		chromedp.Focus("input#name", chromedp.NodeVisible, chromedp.ByQuery),
		input.InsertText("acme-corp"),
		//screenshot(t, "new_org_enter_name"),
		chromedp.Submit("input#name", chromedp.ByQuery),
		//screenshot(t, "new_org_created"),
		matchText(t, "//div[@role='alert']", "created organization: acme-corp"),
		// go to organization settings
		err := page.Locator("#settings > a").Click()
require.NoError(t, err)
		//screenshot(t),
		// change organization name
		chromedp.Focus("input#name", chromedp.NodeVisible, chromedp.ByQuery),
		chromedp.Clear("input#name", chromedp.ByQuery),
		input.InsertText("super-duper-org"),
		//screenshot(t),
		err := page.Locator(`//button[text()='Update organization name']`).Click()
require.NoError(t, err)
		//screenshot(t),
		matchText(t, "//div[@role='alert']", "updated organization"),
		// delete the organization
		err := page.Locator(`//button[@id='delete-organization-button']`).Click()
require.NoError(t, err)
		//screenshot(t),
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
	page := browser.New(t, ctx)
		// go to the list of organizations
		_, err = page.Goto("https://" + daemon.System.Hostname() + "/app/organizations")
require.NoError(t, err)
		// should be 100 orgs listed on page one
		chromedp.Nodes(`.widget`, &pageOneWidgets, chromedp.NodeVisible),
		// go to page two
		err := page.Locator(`#next-page-link`).Click()
require.NoError(t, err)
		// should be one org listed
		chromedp.Nodes(`.widget`, &pageTwoWidgets, chromedp.NodeVisible),
	})
	assert.Equal(t, 100, len(pageOneWidgets))
	assert.Equal(t, 1, len(pageTwoWidgets))
}
