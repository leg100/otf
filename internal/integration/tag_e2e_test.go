package integration

import (
	"fmt"
	"testing"
	"time"

	goexpect "github.com/google/goexpect"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/workspace"
	"github.com/playwright-community/playwright-go"
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
`, daemon.System.Hostname(), org.Name))

	// run terraform init
	_, token := daemon.createToken(t, ctx, nil)
	e, tferr, err := goexpect.SpawnWithArgs(
		[]string{tfpath, "-chdir=" + root, "init", "-no-color"},
		time.Minute,
		goexpect.PartialMatch(true),
		goexpect.SetEnv(internal.SafeAppend(sharedEnvs, internal.CredentialEnv(daemon.System.Hostname(), token))),
	)
	require.NoError(t, err)
	defer e.Close()

	// create tagged workspace when prompted
	e.ExpectBatch([]goexpect.Batcher{
		&goexpect.BExp{R: "Enter a value: "},
		&goexpect.BSnd{S: "tagged\n"},
		&goexpect.BExp{R: "Terraform Cloud has been successfully initialized!"},
	}, time.Second*5)
	// Terraform should return with exit code 0
	require.NoError(t, <-tferr, e.String)

	// confirm tagged workspace has been created
	got, err := daemon.Workspaces.List(ctx, workspace.ListOptions{
		Organization: &org.Name,
		Tags:         []string{"foo", "bar"},
	})
	require.NoError(t, err)
	require.Equal(t, 1, len(got.Items))
	ws := got.Items[0]
	require.Equal(t, 2, len(ws.Tags))
	require.Contains(t, ws.Tags, "foo")
	require.Contains(t, ws.Tags, "bar")

	// test UI management of tags
	browser.New(t, ctx, func(page playwright.Page) {
		_, err = page.Goto(workspaceURL(daemon.System.Hostname(), org.Name, "tagged"))
		require.NoError(t, err)
		// confirm workspace page lists both tags
		err = expect.Locator(page.Locator(`//*[@id='tags']//span[contains(text(),'foo')]`)).ToBeVisible()
		require.NoError(t, err)
		err = expect.Locator(page.Locator(`//*[@id='tags']//span[contains(text(),'bar')]`)).ToBeVisible()
		require.NoError(t, err)
		// remove bar tag
		err = page.Locator(`//button[@id='button-remove-tag-bar']`).Click()
		require.NoError(t, err)
		err = expect.Locator(page.GetByRole("alert")).ToHaveText("removed tag: bar")
		require.NoError(t, err)

		// add new tag
		err = page.Locator(`//input[@x-ref='input-search']`).Fill("baz")
		require.NoError(t, err)

		err = page.Locator(`//input[@x-ref='input-search']`).Press("Enter")
		require.NoError(t, err)

		err = expect.Locator(page.GetByRole("alert")).ToHaveText("created tag: baz")

		require.NoError(t, err)
		// go to workspace listing
		err = page.Locator(`//span[@id='content-header-title']//a[text()='workspaces']`).Click()
		require.NoError(t, err)
		// show tags
		err = page.Locator(`//input[@name='tag_filter_visible']`).Click()
		require.NoError(t, err)
		// filter by tag foo
		err = page.Locator(`//input[@id='filter-tag-foo']`).Click()
		require.NoError(t, err)
		// filter by tag baz
		err = page.Locator(`//input[@id='filter-tag-baz']`).Click()
		require.NoError(t, err)
		// confirm workspace listing contains tagged workspace
		err = expect.Locator(page.Locator(`//div[@id='page-results']//tr[@id='item-workspace-tagged']`)).ToBeVisible()
		require.NoError(t, err)
	})

	// should be tags 'foo' and 'baz'
	tags, err := daemon.Workspaces.ListTags(ctx, org.Name, workspace.ListTagsOptions{})
	require.NoError(t, err)
	assert.Len(t, tags.Items, 2)

	// demonstrate deleting the workspace also deletes the tags from the system
	_, err = daemon.Workspaces.Delete(ctx, ws.ID)
	require.NoError(t, err)

	// should be no tags
	tags, err = daemon.Workspaces.ListTags(ctx, org.Name, workspace.ListTagsOptions{})
	require.NoError(t, err)
	assert.Len(t, tags.Items, 0)
}
