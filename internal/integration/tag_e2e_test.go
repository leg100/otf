package integration

import (
	"fmt"
	"testing"
	"time"

	"github.com/chromedp/cdproto/input"
	"github.com/chromedp/chromedp"
	expect "github.com/google/goexpect"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/workspace"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestIntegration_TagsE2E demonstrates end-to-end usage of workspace tags.
func TestIntegration_TagsE2E(t *testing.T) {
	integrationTest(t)

	daemon, org, ctx := setup(t, nil)

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
	_, token := daemon.createToken(t, ctx, nil)
	e, tferr, err := expect.SpawnWithArgs(
		[]string{"terraform", "-chdir=" + root, "init", "-no-color"},
		time.Minute,
		expect.PartialMatch(true),
		expect.SetEnv(internal.SafeAppend(sharedEnvs, internal.CredentialEnv(daemon.Hostname(), token))),
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
		Organization: internal.String(org.Name),
		Tags:         []string{"foo", "bar"},
	})
	require.NoError(t, err)
	require.Equal(t, 1, len(got.Items))
	ws := got.Items[0]
	require.Equal(t, 2, len(ws.Tags))
	require.Contains(t, ws.Tags, "foo")
	require.Contains(t, ws.Tags, "bar")

	// test UI management of tags
	browser.Run(t, ctx, chromedp.Tasks{
		chromedp.Navigate(workspaceURL(daemon.Hostname(), org.Name, "tagged")),
		// confirm workspace page lists both tags
		chromedp.WaitVisible(`//*[@id='tags']//span[contains(text(),'foo')]`),
		chromedp.WaitVisible(`//*[@id='tags']//span[contains(text(),'bar')]`),
		// remove bar tag
		chromedp.Click(`//button[@id='button-remove-tag-bar']`),
		screenshot(t),
		matchText(t, "//div[@role='alert']", "removed tag: bar"),
		// add new tag
		chromedp.Focus(`//input[@x-ref='input-search']`),
		input.InsertText("baz"),
		chromedp.Submit(`//input[@x-ref='input-search']`),
		screenshot(t),
		matchText(t, "//div[@role='alert']", "created tag: baz"),
		// go to workspace listing
		chromedp.Click(`//span[@id='content-header-title']//a[text()='workspaces']`),
		screenshot(t),
		// filter by tag foo
		chromedp.Click(`//label[@for='workspace-tag-filter-foo']`),
		screenshot(t),
		// filter by tag bar
		chromedp.Click(`//label[@for='workspace-tag-filter-baz']`),
		screenshot(t),
		// confirm workspace listing contains tagged workspace
		chromedp.WaitVisible(`//div[@id='content-list']//a[text()='tagged']`),
	})

	// should be tags 'foo' and 'baz'
	tags, err := daemon.ListTags(ctx, org.Name, workspace.ListTagsOptions{})
	require.NoError(t, err)
	assert.Len(t, tags.Items, 2)

	// demonstrate deleting the workspace also deletes the tags from the system
	_, err = daemon.DeleteWorkspace(ctx, ws.ID)
	require.NoError(t, err)

	// should be no tags
	tags, err = daemon.ListTags(ctx, org.Name, workspace.ListTagsOptions{})
	require.NoError(t, err)
	assert.Len(t, tags.Items, 0)
}
