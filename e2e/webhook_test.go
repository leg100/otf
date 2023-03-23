package e2e

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/google/uuid"
	"github.com/leg100/otf/cloud"
	"github.com/stretchr/testify/require"
)

// TestWebhook tests webhook functionality. Two workspaces are created and
// connected to a repository. This should result in the creation of a webhook
// on github. Both workspaces are then disconnected. This should result in the
// deletion of the webhook from github.
//
// Functionality specific to VCS events triggering webhooks and in turn
// triggering workspace runs and publishing module versions in tested in other
// E2E tests.
func TestWebhook(t *testing.T) {
	addBuildsToPath(t)

	org := uuid.NewString()
	user := cloud.User{
		Name: uuid.NewString(),
		Teams: []cloud.Team{
			{
				Name:         "owners",
				Organization: org,
			},
		},
		Organizations: []string{org},
	}

	repo := cloud.NewTestRepo()
	tarball, err := os.ReadFile("../testdata/github.tar.gz")
	require.NoError(t, err)

	// create an otf daemon with a fake github backend, ready to sign in a user,
	// serve up a repo and its contents via tarball. And register a callback to
	// test receipt of commit statuses
	daemon := &daemon{}
	daemon.withGithubUser(&user)
	daemon.withGithubRepo(repo)
	daemon.withGithubTarball(tarball)

	hostname := daemon.start(t)
	url := "https://" + hostname

	// create browser
	ctx, cancel := chromedp.NewContext(allocator)
	defer cancel()

	// create timeout
	ctx, cancel = context.WithTimeout(ctx, time.Minute)
	defer cancel()

	// create and connect first workspace
	err = chromedp.Run(ctx, chromedp.Tasks{
		// need to login first and create a vcs provider
		githubLoginTasks(t, hostname, user.Name),
		createGithubVCSProviderTasks(t, url, org, "github"),

		createWorkspaceTasks(t, hostname, org, "workspace-1"),
		connectWorkspaceTasks(t, url, org, "workspace-1"),
	})
	require.NoError(t, err)

	// webhook should now have been registered with github
	require.True(t, daemon.githubServer.HasWebhook())

	// create and connect second workspace
	err = chromedp.Run(ctx, chromedp.Tasks{
		createWorkspaceTasks(t, hostname, org, "workspace-2"),
		connectWorkspaceTasks(t, url, org, "workspace-2"),
	})
	require.NoError(t, err)

	// second workspace re-uses same webhook on github
	require.True(t, daemon.githubServer.HasWebhook())

	// disconnect second workspace
	err = chromedp.Run(ctx, disconnectWorkspaceTasks(t, url, org, "workspace-2"))
	require.NoError(t, err)

	// first workspace is still connected, so webhook should still be configured
	// on github
	require.True(t, daemon.githubServer.HasWebhook())

	// disconnect first workspace
	err = chromedp.Run(ctx, disconnectWorkspaceTasks(t, url, org, "workspace-1"))
	require.NoError(t, err)

	// No more workspaces are connected to repo, so webhook should have been
	// deleted
	require.False(t, daemon.githubServer.HasWebhook())
}
