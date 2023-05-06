package agent

import (
	"os"
	"path"
)

type terraform interface {
	TerraformPath(version string) string
}

type terraformPathFinder struct{}

// TerraformPath returns the path to a given version of the terraform binary
func (*terraformPathFinder) TerraformPath(version string) string {
	return path.Join(os.TempDir(), "otf-terraform-bins", version, "terraform")
}
