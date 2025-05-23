package integration

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestCloudBlock tests support for terraform's 'cloud' block:
//
// https://developer.hashicorp.com/terraform/cli/cloud/settings#the-cloud-block
func TestCloudBlock(t *testing.T) {
	integrationTest(t)

	svc, org, ctx := setup(t)

	// create terraform root module with a cloud configuration block.
	config := fmt.Sprintf(`terraform {
  cloud {
	hostname = "%s"
	organization = "%s"

	workspaces {
	  name = "test-workspace"
	}
  }
}
resource "null_resource" "e2e" {}
`, svc.System.Hostname(), org.Name)
	root := t.TempDir()
	err := os.WriteFile(filepath.Join(root, "main.tf"), []byte(config), 0o600)
	require.NoError(t, err)

	// a successful 'terraform init' sufficiently demonstrates support for the
	// 'cloud' block.
	out := svc.engineCLI(t, ctx, "", "init", root)
	require.Contains(t, out, "Terraform Cloud has been successfully initialized!")
}
