package otf

import (
	"bytes"
	"fmt"
	"os"

	"github.com/hashicorp/terraform-config-inspect/tfconfig"
	"github.com/pkg/errors"
)

func UnmarshalTerraformModule(tarball []byte) (*tfconfig.Module, error) {
	dir, err := os.MkdirTemp("", "")
	if err != nil {
		return nil, errors.Wrap(err, "creating temporary directory")
	}
	if err := Unpack(bytes.NewReader(tarball), dir); err != nil {
		return nil, errors.Wrap(err, "extracting tarball")
	}

	// parse module to check that it is valid
	mod, diags := tfconfig.LoadModule(dir)
	if diags.HasErrors() {
		return nil, fmt.Errorf("parsing HCL: %s", diags.Error())
	}
	// valid module
	return mod, nil
}
