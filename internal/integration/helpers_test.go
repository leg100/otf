package integration

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/leg100/otf/internal/engine"
	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/pubsub"
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

// newRootModule creates a terraform root module, returning its directory path
func newRootModule(t *testing.T, hostname string, organization organization.Name, workspace string, additionalConfig ...string) string {
	t.Helper()

	var config strings.Builder
	config.WriteString(fmt.Sprintf(`
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
`, hostname, organization, workspace))
	for _, cfg := range additionalConfig {
		config.WriteString("\n")
		config.WriteString(cfg)
	}

	return createRootModule(t, config.String())
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

// wait for an event to arrive satisfying the condition within a timeout.
func wait[T any](t *testing.T, c <-chan pubsub.Event[T], cond func(pubsub.Event[T]) bool) {
	t.Helper()

	timeout := time.After(5 * time.Second)
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

type engineTestSpec struct {
	*engine.Engine
	name string
	path string
}

func engineTestSpecs() []engineTestSpec {
	return []engineTestSpec{
		{name: "Terraform", path: terraformPath, Engine: engine.Terraform},
		{name: "OpenTofu", path: tofuPath, Engine: engine.Tofu},
	}
}
