package otf

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRewriteHCL(t *testing.T) {
	got := `
terraform {
  backend "remote" {
    hostname = "app.terraform.io"
    organization = "automatize"

    workspaces {
      prefix = "etok-example-"
    }
  }
}
`
	want := `
terraform {
}
`

	modulePath := t.TempDir()
	cfgFile := filepath.Join(modulePath, "config.tf")
	require.NoError(t, os.WriteFile(cfgFile, []byte(got), 0o644))

	err := rewriteHCL(modulePath, removeBackendBlock)
	assert.NoError(t, err)

	f, err := os.ReadFile(cfgFile)
	require.NoError(t, err)
	assert.Equal(t, want, string(f))
}
