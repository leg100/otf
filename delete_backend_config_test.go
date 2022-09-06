package otf

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDeleteBackendConfig(t *testing.T) {
	config := `
terraform {
  required_providers {
    aws = {
      version = ">= 2.7.0"
      source = "hashicorp/aws"
    }
  }
  backend "remote" {
    hostname = "app.terraform.io"
    organization = "automatize"

    workspaces {
      prefix = "etok-example-"
    }
  }
  required_version = "0.12.5"
}
`

	deleted, got, err := deleteBackendConfig([]byte(config))
	require.NoError(t, err)
	assert.True(t, deleted)

	want := `
terraform {
  required_providers {
    aws = {
      version = ">= 2.7.0"
      source  = "hashicorp/aws"
    }
  }
  required_version = "0.12.5"
}
`

	assert.Equal(t, want, string(got))
}

func TestDeleteBackendConfigFromDirectory(t *testing.T) {
	config := `
terraform {
  required_providers {
    aws = {
      version = ">= 2.7.0"
      source = "hashicorp/aws"
    }
  }
  backend "remote" {
    hostname = "app.terraform.io"
    organization = "automatize"

    workspaces {
      prefix = "etok-example-"
    }
  }
  required_version = "0.12.5"
}
`

	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "config.tf"), []byte(config), 0o644))

	assert.NoError(t, deleteBackendConfigFromDirectory(context.Background(), dir))
}
