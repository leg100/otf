package agent

import (
	"os"
	"path"
)

var defaultTerraformBinDir = path.Join(os.TempDir(), "otf-terraform-bins")

type (
	terraformPathFinder struct {
		dest string
	}
)

func newTerraformPathFinder(dest string) terraformPathFinder {
	if dest == "" {
		dest = defaultTerraformBinDir
	}
	return terraformPathFinder{
		dest: dest,
	}
}

func (t terraformPathFinder) TerraformPath(version string) string {
	return path.Join(t.dest, version, "terraform")
}
