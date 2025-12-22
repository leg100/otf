package module

import (
	"bytes"
	"fmt"
	"os"
	"path"

	"github.com/hashicorp/terraform-config-inspect/tfconfig"
	"github.com/leg100/otf/internal"
	"github.com/pkg/errors"
)

// TerraformModule is a module of terraform configuration
type TerraformModule struct {
	*tfconfig.Module

	readme []byte
}

func UnmarshalTerraformModule(tarball []byte) (*TerraformModule, error) {
	dir, err := os.MkdirTemp("", "")
	if err != nil {
		return nil, errors.Wrap(err, "creating temporary directory")
	}
	if err := internal.Unpack(bytes.NewReader(tarball), dir); err != nil {
		return nil, errors.Wrap(err, "extracting tarball")
	}

	// parse module to check that it is valid
	mod, diags := tfconfig.LoadModule(dir)
	if diags.HasErrors() {
		return nil, fmt.Errorf("parsing HCL: %s", diags.Error())
	}

	tfmod := &TerraformModule{Module: mod}

	// retrieve readme if there is one
	if readme, err := os.ReadFile(path.Join(dir, "README.md")); err == nil {
		tfmod.readme = readme
	}

	// valid module
	return tfmod, nil
}

// GetReadme returns the module's readme content
func (t *TerraformModule) GetReadme() []byte {
	return t.readme
}
