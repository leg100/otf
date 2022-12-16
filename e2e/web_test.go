package e2e

import (
	"fmt"
	"path"
	"testing"

	"github.com/chromedp/chromedp"
	"github.com/leg100/otf"
	"github.com/stretchr/testify/require"
)

// TestWeb is a random walkthrough of the Web UI
func TestWeb(t *testing.T) {
	addBuildsToPath(t)

	org := otf.NewTestOrganization(t)
	owners := otf.NewTeam("owners", org)
	devops := otf.NewTeam("devops", org)
	user := otf.NewTestUser(t, otf.WithOrganizationMemberships(org), otf.WithTeamMemberships(owners, devops))

	daemon := &daemon{}
	daemon.withGithubUser(user)
	hostname := daemon.start(t)
	url := "https://" + hostname
	workspaceName := "test-web"

	// create browser
	ctx, cancel := chromedp.NewContext(allocator)
	defer cancel()

	err := chromedp.Run(ctx, chromedp.Tasks{
		// login
		githubLoginTasks(t, hostname, user.Username()),
		// create workspace
		createWorkspaceTasks(t, hostname, org.Name(), workspaceName),
		// assign workspace manager role to devops team
		chromedp.Tasks{
			// go to org
			chromedp.Navigate(path.Join(url, "organizations", org.Name())),
			// list teams
			chromedp.Click("#teams > a", chromedp.NodeVisible, chromedp.ByQuery),
			// select devops team
			chromedp.Click("#item-team-devops a", chromedp.NodeVisible, chromedp.ByQuery),
			screenshot(t),
			// tick checkbox for workspace manager role
			chromedp.Click("#manage_workspaces", chromedp.NodeVisible, chromedp.ByQuery),
			// submit form
			chromedp.Submit("#manage_workspaces", chromedp.NodeVisible, chromedp.ByQuery),
			screenshot(t),
			// confirm permissions updated
			matchText(t, ".flash-success", "team permissions updated"),
		},
		// add write permission on workspace to devops team
		addWorkspacePermissionTasks(t, url, org.Name(), workspaceName, devops.Name(), "write"),
		// list users
		chromedp.Tasks{
			// go to org
			chromedp.Navigate(path.Join(url, "organizations", org.Name())),
			screenshot(t),
			// list users
			chromedp.Click("#users > a", chromedp.NodeVisible),
			screenshot(t),
			matchText(t, fmt.Sprintf("#item-user-%s .status", user.Username()), user.Username()),
		},
		// list team members
		chromedp.Tasks{
			// go to org
			chromedp.Navigate(path.Join(url, "organizations", org.Name())),
			screenshot(t),
			// list teams
			chromedp.Click("#teams > a", chromedp.NodeVisible),
			// select owners team
			chromedp.Click("#item-team-owners a", chromedp.NodeVisible),
			screenshot(t),
			matchText(t, fmt.Sprintf("#item-user-%s .status", user.Username()), user.Username()),
		},
	})
	require.NoError(t, err)
}
