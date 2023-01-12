package e2e

import (
	"path"
	"testing"

	"github.com/chromedp/cdproto/input"
	"github.com/chromedp/chromedp"
	"github.com/google/uuid"
	"github.com/leg100/otf/cloud"
	"github.com/stretchr/testify/require"
)

// TestVariables tests adding, updating and deleting workspace variables via the
// UI.
func TestVariables(t *testing.T) {
	addBuildsToPath(t)

	org := uuid.NewString()
	user := cloud.User{
		Name:          uuid.NewString(),
		Teams:         []cloud.Team{{"owners", org}},
		Organizations: []string{org},
	}

	daemon := &daemon{}
	daemon.withGithubUser(&user)
	hostname := daemon.start(t)
	url := "https://" + hostname
	workspaceName := t.Name()

	// create browser
	ctx, cancel := chromedp.NewContext(allocator)
	defer cancel()

	// Click OK on any browser javascript dialog boxes that pop up
	okDialog(t, ctx)

	err := chromedp.Run(ctx, chromedp.Tasks{
		// login
		githubLoginTasks(t, hostname, user.Name),
		// create workspace
		createWorkspaceTasks(t, hostname, org, workspaceName),
		// assign workspace manager role to devops team
		chromedp.Tasks{
			// go to workspace
			chromedp.Navigate(path.Join(url, "organizations", org, "workspaces", workspaceName)),
			screenshot(t),
			// go to variables
			chromedp.Click(`//a[text()='variables']`, chromedp.NodeVisible),
			screenshot(t),
			// click add variable button
			chromedp.Click(`//button[text()='Add variable']`, chromedp.NodeVisible),
			screenshot(t),
			// enter key
			chromedp.Focus("input#key", chromedp.NodeVisible),
			input.InsertText("foo"),
			screenshot(t),
			// enter value
			chromedp.Focus("input#value", chromedp.NodeVisible),
			input.InsertText("bar"),
			screenshot(t),
			// select terraform variable category
			chromedp.Click("input#terraform", chromedp.NodeVisible),
			screenshot(t),
			// submit form
			chromedp.Click(`//button[text()='Save variable']`, chromedp.NodeVisible),
			screenshot(t),
			// confirm variable added
			matchText(t, ".flash-success", "added variable: foo"),
			screenshot(t),
			// edit variable
			chromedp.Click(`//a[text()='foo']`, chromedp.NodeVisible),
			screenshot(t),
			// make it a 'sensitive' variable
			chromedp.Click("input#sensitive", chromedp.NodeVisible, chromedp.ByQuery),
			screenshot(t),
			// submit form
			chromedp.Click(`//button[text()='Save variable']`, chromedp.NodeVisible),
			screenshot(t),
			// confirm variable updated
			matchText(t, ".flash-success", "updated variable: foo"),
			screenshot(t),
			// confirm value is hidden (because it is sensitive)
			matchText(t, `//table[@class='variables']/tbody/tr/td[2]`, "hidden"),
			// delete variable
			chromedp.Click(`//button[text()='Delete']`, chromedp.NodeVisible),
			screenshot(t),
			// confirm variable deleted
			matchText(t, ".flash-success", "deleted variable: foo"),
		},
	})
	require.NoError(t, err)
}
