package agent

import (
	"os"
	"path"
)

type Terraform interface {
	TerraformPath(version string) string
}

type TerraformPathFinder struct{}

// TerraformPath returns the path to a given version of the terraform binary
func (*TerraformPathFinder) TerraformPath(version string) string {
	return path.Join(os.TempDir(), "otf-terraform-bins", version, "terraform")
}
