package e2e

import (
	"fmt"
	"testing"

	"github.com/chromedp/chromedp"
	"github.com/google/uuid"
	"github.com/leg100/otf/cloud"
	"github.com/stretchr/testify/require"
)

// TestWeb is a random walkthrough of the Web UI
func TestWeb(t *testing.T) {
	org, workspace := setup(t)

	user := cloud.User{
		Name: uuid.NewString(),
		Teams: []cloud.Team{
			{
				Name:         "owners",
				Organization: org,
			},
			{
				Name:         "devops",
				Organization: org,
			},
		},
		Organizations: []string{org},
	}

	daemon := &daemon{}
	daemon.withGithubUser(&user)
	hostname := daemon.start(t)

	// create browser
	ctx, cancel := chromedp.NewContext(allocator)
	defer cancel()

	err := chromedp.Run(ctx, chromedp.Tasks{
		// login
		githubLoginTasks(t, hostname, user.Name),
		// create workspace
		createWorkspaceTasks(t, hostname, org, workspace),
		// assign workspace manager role to devops team
		chromedp.Tasks{
			// go to org
			chromedp.Navigate(organizationPath(hostname, org)),
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
		addWorkspacePermissionTasks(t, hostname, org, workspace, "devops", "write"),
		// list users
		chromedp.Tasks{
			// go to org
			chromedp.Navigate(organizationPath(hostname, org)),
			screenshot(t),
			// list users
			chromedp.Click("#users > a", chromedp.NodeVisible),
			screenshot(t),
			matchText(t, fmt.Sprintf("#item-user-%s .status", user.Name), user.Name),
		},
		// list team members
		chromedp.Tasks{
			// go to org
			chromedp.Navigate(organizationPath(hostname, org)),
			screenshot(t),
			// list teams
			chromedp.Click("#teams > a", chromedp.NodeVisible),
			// select owners team
			chromedp.Click("#item-team-owners a", chromedp.NodeVisible),
			screenshot(t),
			matchText(t, fmt.Sprintf("#item-user-%s .status", user.Name), user.Name),
		},
	})
	require.NoError(t, err)
}
