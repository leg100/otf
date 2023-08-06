package integration

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/auth"
	"github.com/leg100/otf/internal/workspace"
	"github.com/stretchr/testify/require"
)

func integrationTest(t *testing.T) {
	// An integration test can take a while to run so it be run in parallel to
	// other integration tests
	t.Parallel()

	// Skip long-running integration tests if user has passed -short flag
	if testing.Short() {
		t.Skip()
	}
}

func runURL(hostname, runID string) string {
	return "https://" + hostname + "/app/runs/" + runID
}

func runsURL(hostname, workspaceID string) string {
	return "https://" + hostname + "/app/workspaces/" + workspaceID + "/runs"
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
  cloud {
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

func userFromContext(t *testing.T, ctx context.Context) *auth.User {
	user, err := auth.UserFromContext(ctx)
	require.NoError(t, err)
	return user
}

func downloadTerraform(t *testing.T, ctx context.Context, version *string) string {
	t.Helper()

	if version == nil {
		version = internal.String(workspace.DefaultTerraformVersion)
	}
	tfpath, err := tfDownloader.Download(ctx, *version, io.Discard)
	require.NoError(t, err)
	return tfpath
}
