package integration

import (
	"fmt"
	"testing"
	"time"

	"github.com/chromedp/cdproto/input"
	"github.com/chromedp/chromedp"
	expect "github.com/google/goexpect"
	"github.com/leg100/otf"
	"github.com/leg100/otf/workspace"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestIntegration_TagsE2E demonstrates end-to-end usage of workspace tags.
func TestIntegration_TagsE2E(t *testing.T) {
	t.Parallel()

	daemon := setup(t, nil)
	user, ctx := daemon.createUserCtx(t, ctx)
	org := daemon.createOrganization(t, ctx)

	// create a root module with a cloud backend configured to use workspaces
	// with foo and bar tags.
	root := createRootModule(t, fmt.Sprintf(`
terraform {
  cloud {
	hostname = "%s"
	organization = "%s"

	workspaces {
		tags = ["foo", "bar"]
	}
  }
}
resource "null_resource" "tags_e2e" {}
`, daemon.Hostname(), org.Name))

	// run terraform init
	_, token := daemon.createToken(t, ctx, user)
	e, tferr, err := expect.SpawnWithArgs(
		[]string{"terraform", "-chdir=" + root, "init", "-no-color"},
		time.Minute,
		expect.PartialMatch(true),
		expect.SetEnv(
			append(envs, otf.CredentialEnv(daemon.Hostname(), token)),
		),
	)
	require.NoError(t, err)
	defer e.Close()

	// create tagged workspace when prompted
	e.ExpectBatch([]expect.Batcher{
		&expect.BExp{R: "Enter a value: "},
		&expect.BSnd{S: "tagged\n"},
		&expect.BExp{R: "Terraform Cloud has been successfully initialized!"},
	}, time.Second*5)
	// Terraform should return with exit code 0
	require.NoError(t, <-tferr, e.String)

	// confirm tagged workspace has been created
	got, err := daemon.ListWorkspaces(ctx, workspace.ListOptions{
		Organization: otf.String(org.Name),
		Tags:         []string{"foo", "bar"},
	})
	require.NoError(t, err)
	require.Equal(t, 1, len(got.Items))
	if assert.Equal(t, 2, len(got.Items[0].Tags)) {
		assert.Contains(t, got.Items[0].Tags, "foo")
		assert.Contains(t, got.Items[0].Tags, "bar")
	}

	// test UI management of tags
	browser := createBrowserCtx(t)
	err = chromedp.Run(browser, chromedp.Tasks{
		newSession(t, ctx, daemon.Hostname(), user.Username, daemon.Secret),
		chromedp.Navigate(workspaceURL(daemon.Hostname(), org.Name, "tagged")),
		// confirm workspace page lists both tags
		chromedp.WaitVisible(`//*[@class='workspace-tag'][contains(text(),'foo')]`),
		chromedp.WaitVisible(`//*[@class='workspace-tag'][contains(text(),'bar')]`),
		// go to workspace settings
		chromedp.Click(`//a[text()='settings']`, chromedp.NodeVisible),
		screenshot(t),
		// remove bar tag
		chromedp.Click(`//button[@id='button-remove-tag-bar']`, chromedp.NodeVisible),
		screenshot(t),
		matchText(t, ".flash-success", "removed tag: bar"),
		// add new tag
		chromedp.Focus("input#new-tag-name", chromedp.NodeVisible),
		input.InsertText("baz"),
		chromedp.Click(`//button[text()='Add new tag']`, chromedp.NodeVisible),
		screenshot(t),
		matchText(t, ".flash-success", "created tag: baz"),
	})
	require.NoError(t, err)
}
