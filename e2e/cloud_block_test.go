package e2e

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// TestCloudBlock tests support for terraform's 'cloud' block:
//
// https://developer.hashicorp.com/terraform/cli/cloud/settings#the-cloud-block
func TestCloudBlock(t *testing.T) {
	addBuildsToPath(t)

	token := "abc123" // site-admin's token

	// run daemon configured with site-admin access
	daemon := &daemon{}
	daemon.withFlags("--site-token", token)
	hostname := daemon.start(t)

	// use site-admin for CLI auth (for both 'otf' and 'terraform')
	t.Setenv("TF_TOKEN_"+strings.ReplaceAll(hostname, ".", "_"), token)

	// create org via CLI
	org := uuid.NewString()
	cmd := exec.Command("otf", "organizations", "new", "--address", hostname, org)
	out, err := cmd.CombinedOutput()
	t.Log(string(out))
	require.NoError(t, err)

	config := []byte(fmt.Sprintf(`
terraform {
  cloud {
	hostname = "%s"
	organization = "%s"

	workspaces {
	  name = "test-workspace"
	}
  }
}
resource "null_resource" "e2e" {}
`, hostname, org))

	root := t.TempDir()
	err = os.WriteFile(filepath.Join(root, "main.tf"), config, 0o600)
	require.NoError(t, err)

	// a successful 'terraform init' sufficiently demonstrates support for the
	// 'cloud' block.
	cmd = exec.Command("terraform", "init", "-no-color")
	cmd.Dir = root
	out, err = cmd.CombinedOutput()
	t.Log(string(out))
	require.NoError(t, err)
	require.Contains(t, string(out), "Terraform Cloud has been successfully initialized!")
}
