package agent

import (
	"os"
	"path"
)

var defaultTerraformBinDir = path.Join(os.TempDir(), "otf-terraform-bins")

type (
	TerraformPathFinder struct {
		dest string
	}
)

func newTerraformPathFinder(dest string) *TerraformPathFinder {
	if dest == "" {
		dest = defaultTerraformBinDir
	}
	return &TerraformPathFinder{
		dest: dest,
	}
}

func (t *TerraformPathFinder) TerraformPath(version string) string {
	return path.Join(t.dest, version, "terraform")
}
