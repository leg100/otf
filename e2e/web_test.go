package e2e

import (
	"fmt"
	"path"
	"testing"

	"github.com/chromedp/chromedp"
	"github.com/google/uuid"
	"github.com/leg100/otf/cloud"
	"github.com/stretchr/testify/require"
)

// TestWeb is a random walkthrough of the Web UI
func TestWeb(t *testing.T) {
	addBuildsToPath(t)

	org := uuid.NewString()
	user := cloud.User{
		Name: "cluster-user",
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
	url := "https://" + hostname
	workspaceName := "test-web"

	// create browser
	ctx, cancel := chromedp.NewContext(allocator)
	defer cancel()

	err := chromedp.Run(ctx, chromedp.Tasks{
		// login
		githubLoginTasks(t, hostname, user.Name),
		// create workspace
		createWorkspaceTasks(t, hostname, org, workspaceName),
		// assign workspace manager role to devops team
		chromedp.Tasks{
			// go to org
			chromedp.Navigate(path.Join(url, "organizations", org)),
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
		addWorkspacePermissionTasks(t, url, org, workspaceName, "devops", "write"),
		// list users
		chromedp.Tasks{
			// go to org
			chromedp.Navigate(path.Join(url, "organizations", org)),
			screenshot(t),
			// list users
			chromedp.Click("#users > a", chromedp.NodeVisible),
			screenshot(t),
			matchText(t, fmt.Sprintf("#item-user-%s .status", user.Name), user.Name),
		},
		// list team members
		chromedp.Tasks{
			// go to org
			chromedp.Navigate(path.Join(url, "organizations", org)),
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
