package integration

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/leg100/otf/internal/pubsub"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/user"
	"github.com/stretchr/testify/require"
)

func integrationTest(t *testing.T) {
	t.Parallel()

	// Skip long-running integration tests if user has passed -short flag
	if testing.Short() {
		t.Skip()
	}
}

func runURL(hostname string, runID resource.ID) string {
	return fmt.Sprintf("https://%s/app/runs/%s", hostname, runID)
}

func runsURL(hostname string, workspaceID resource.ID) string {
	return fmt.Sprintf("https://%s/app/workspaces/%s/runs", hostname, workspaceID)
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

func userFromContext(t *testing.T, ctx context.Context) *user.User {
	user, err := user.UserFromContext(ctx)
	require.NoError(t, err)
	return user
}

// Wait for an event to arrive satisfying the condition within a 10 second timeout.
func Wait[T any](t *testing.T, c <-chan pubsub.Event[T], cond func(pubsub.Event[T]) bool) {
	timeout := time.After(10 * time.Second)
	for {
		select {
		case <-timeout:
			t.Fatal("timed out waiting for event")
		case event := <-c:
			if cond(event) {
				return
			}
		}
	}
}
