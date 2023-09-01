package integration

import (
	"testing"

	"github.com/chromedp/cdproto/input"
	"github.com/chromedp/chromedp"
)

// TestIntegration_VariableSetUI tests management of variable sets via the UI.
func TestIntegration_VariableSetUI(t *testing.T) {
	integrationTest(t)

	svc, org, ctx := setup(t, nil)

	// Create global variable set in browser
	browser.Run(t, ctx, chromedp.Tasks{
		chromedp.Tasks{
			// go to org
			chromedp.Navigate(organizationURL(svc.Hostname(), org.Name)),
			// go to variable sets
			chromedp.Click(`//a[text()='variable sets']`),
			// click new variable set button
			chromedp.Click(`button#new-variable-set-button`, chromedp.ByQuery),
			// enter name
			chromedp.Focus("input#name", chromedp.NodeVisible, chromedp.ByQuery),
			input.InsertText("foo"),
			// enter description
			chromedp.Focus("textarea#description", chromedp.NodeVisible, chromedp.ByQuery),
			input.InsertText("this is my special foo variable"),
			// global radio button should be set by default
			chromedp.WaitVisible(`input#global:checked`, chromedp.ByQuery),
			// submit form
			chromedp.Click(`//button[@id='save-variable-set-button']`),
			// confirm variable set added
			matchText(t, "//div[@role='alert']", "added variable set: foo"),
		},
	})
}
