package integration

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/chromedp/chromedp"
	"github.com/stretchr/testify/require"
)

func createTabCtx(t *testing.T) context.Context {
	t.Helper()

	ctx, cancel := chromedp.NewContext(browser)
	t.Cleanup(cancel)

	return ctx
}

func createBrowserCtx(t *testing.T) context.Context {
	t.Helper()

	ctx, cancel := chromedp.NewContext(allocator)
	t.Cleanup(cancel)

	return ctx
}

func runURL(hostname, runID string) string {
	return "https://" + hostname + "/app/runs/" + runID
}

func workspaceURL(hostname, org, name string) string {
	return "https://" + hostname + "/app/organizations/" + org + "/workspaces/" + name
}

func workspacesURL(hostname, org string) string {
	return "https://" + hostname + "/app/organizations/" + org + "/workspaces"
}

func organizationURL(hostname, org string) string {
	return "https://" + hostname + "/app/organizations/" + org
}

// newRootModule creates a terraform root module, returning its directory path
func newRootModule(t *testing.T, hostname, organization, workspace string, additionalConfig ...string) string {
	t.Helper()

	config := fmt.Sprintf(`
terraform {
  backend "remote" {
	hostname = "%s"
	organization = "%s"

	workspaces {
	  name = "%s"
	}
  }
}
resource "null_resource" "e2e" {}
`, hostname, organization, workspace)
	for _, cfg := range additionalConfig {
		config += "\n"
		config += cfg
	}

	return createRootModule(t, config)
}

func createRootModule(t *testing.T, tfconfig string) string {
	root := t.TempDir()
	err := os.WriteFile(filepath.Join(root, "main.tf"), []byte(tfconfig), 0o600)
	require.NoError(t, err)

	return root
}
