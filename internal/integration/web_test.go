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
	t.Parallel()

	svc := setup(t, nil)
	user, ctx := svc.createUserCtx(t, ctx)
	org := svc.createOrganization(t, ctx)
	team, err := svc.CreateTeam(ctx, auth.CreateTeamOptions{
		Organization: org.Name,
		Name:         "devops",
	})
	require.NoError(t, err)
	err = svc.AddTeamMembership(ctx, auth.TeamMembershipOptions{
		TeamID:   team.ID,
		Username: user.Username,
	})
	require.NoError(t, err)

	browser := createBrowserCtx(t)
	err = chromedp.Run(browser, chromedp.Tasks{
		newSession(t, ctx, svc.Hostname(), user.Username, svc.Secret),
		// create workspace
		createWorkspace(t, svc.Hostname(), org.Name, "my-workspace"),
		// assign workspace manager role to devops team
		chromedp.Tasks{
			// go to org
			chromedp.Navigate(organizationURL(svc.Hostname(), org.Name)),
			screenshot(t),
			// list teams
			chromedp.Click("#teams > a", chromedp.NodeVisible, chromedp.ByQuery),
			screenshot(t),
			// select devops team
			chromedp.Click("#item-team-devops a", chromedp.NodeVisible, chromedp.ByQuery),
			screenshot(t),
			// tick checkbox for workspace manager role
			chromedp.Click("#manage_workspaces", chromedp.NodeVisible, chromedp.ByQuery),
			// submit form
			chromedp.Submit("#manage_workspaces", chromedp.NodeVisible, chromedp.ByQuery),
			screenshot(t, "team_permissions_added_workspace_manager"),
			// confirm permissions updated
			matchText(t, ".flash-success", "team permissions updated"),
		},
		// add write permission on workspace to devops team
		addWorkspacePermission(t, svc.Hostname(), org.Name, "my-workspace", "devops", "write"),
		// list users
		chromedp.Tasks{
			// go to org
			chromedp.Navigate(organizationURL(svc.Hostname(), org.Name)),
			screenshot(t),
			// list users
			chromedp.Click("#users > a", chromedp.NodeVisible),
			screenshot(t),
			matchText(t, fmt.Sprintf("#item-user-%s .status", user.Username), user.Username),
		},
		// list team members
		chromedp.Tasks{
			// go to org
			chromedp.Navigate(organizationURL(svc.Hostname(), org.Name)),
			screenshot(t),
			// list teams
			chromedp.Click("#teams > a", chromedp.NodeVisible),
			// select owners team
			chromedp.Click("#item-team-owners a", chromedp.NodeVisible),
			screenshot(t),
			matchText(t, fmt.Sprintf("#item-user-%s .status", user.Username), user.Username),
		},
	})
	require.NoError(t, err)
}
