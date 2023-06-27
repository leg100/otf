package integration

import (
	"fmt"
	"testing"

	"github.com/chromedp/chromedp"
	"github.com/leg100/otf/internal/auth"
	"github.com/stretchr/testify/require"
)

// TestWeb is a random walkthrough of the Web UI
func TestWeb(t *testing.T) {
	integrationTest(t)

	daemon, org, ctx := setup(t, nil)
	user := userFromContext(t, ctx)

	team, err := daemon.CreateTeam(ctx, auth.CreateTeamOptions{
		Organization: org.Name,
		Name:         "devops",
	})
	require.NoError(t, err)
	err = daemon.AddTeamMembership(ctx, auth.TeamMembershipOptions{
		TeamID:   team.ID,
		Username: user.Username,
	})
	require.NoError(t, err)

	browser.Run(t, ctx, chromedp.Tasks{
		// create workspace
		createWorkspace(t, daemon.Hostname(), org.Name, "my-workspace"),
		// assign workspace manager role to devops team
		chromedp.Tasks{
			// go to org
			chromedp.Navigate(organizationURL(daemon.Hostname(), org.Name)),
			screenshot(t),
			// list teams
			chromedp.Click("#teams > a", chromedp.ByQuery),
			screenshot(t),
			// select devops team
			chromedp.Click("#item-team-devops a", chromedp.ByQuery),
			screenshot(t),
			// tick checkbox for workspace manager role
			chromedp.Click("#manage_workspaces", chromedp.ByQuery),
			// submit form
			chromedp.Submit("#manage_workspaces", chromedp.NodeVisible, chromedp.ByQuery),
			screenshot(t, "team_permissions_added_workspace_manager"),
			// confirm permissions updated
			matchText(t, ".flash-success", "team permissions updated", chromedp.ByQuery),
		},
		// add write permission on workspace to devops team
		addWorkspacePermission(t, daemon.Hostname(), org.Name, "my-workspace", "devops", "write"),
		// list users
		chromedp.Tasks{
			// go to org
			chromedp.Navigate(organizationURL(daemon.Hostname(), org.Name)),
			screenshot(t),
			// list users
			chromedp.Click("#users > a", chromedp.ByQuery),
			screenshot(t),
			matchText(t, fmt.Sprintf("#item-user-%s .status", user.Username), user.Username, chromedp.ByQuery),
		},
		// list team members
		chromedp.Tasks{
			// go to org
			chromedp.Navigate(organizationURL(daemon.Hostname(), org.Name)),
			screenshot(t),
			// list teams
			chromedp.Click("#teams > a", chromedp.ByQuery),
			// select owners team
			chromedp.Click("#item-team-owners a", chromedp.ByQuery),
			screenshot(t),
			matchText(t, fmt.Sprintf("#item-user-%s .status", user.Username), user.Username, chromedp.ByQuery),
		},
	})
}
