package integration

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestLockFile tests various scenarios relating to the dependency lock file
// (.terraform.lock.hcl). Users may upload one, they may not, they may upload
// one with hashes for a different OS (e.g. Mac) and OTF is running on linux and
// has to then generates hashes for linux, etc. It is notorious for causing
// difficulties for users and it's no different for OTF.
func TestLockFile(t *testing.T) {
	integrationTest(t)

	svc, org, ctx := setup(t, nil)

	// in a browser, create workspace
	browser.Run(t, ctx, createWorkspace(t, svc.Hostname(), org.Name, "my-test-workspace"))

	// create root module with only a variable and no resources - this should
	// result in *no* lock file being created.
	root := t.TempDir()
	config := []byte(fmt.Sprintf(`
terraform {
  cloud {
	hostname = "%s"
	organization = "%s"

	workspaces {
	  name = "%s"
	}
  }
}

variable "foo" {
	default = "bar"
}
`, svc.Hostname(), org.Name, "my-test-workspace"))
	err := os.WriteFile(filepath.Join(root, "main.tf"), []byte(config), 0o600)
	require.NoError(t, err)

	// verify terraform init and plan run without error
	svc.tfcli(t, ctx, "init", root)
	svc.tfcli(t, ctx, "plan", root)
}
