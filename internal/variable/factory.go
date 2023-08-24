package variable

import (
	"crypto/rand"
	"errors"
	"fmt"

	"github.com/leg100/otf/internal"
)

type (
	// factory construct new variables
	factory struct {
		// generates random version IDs
		generateVersion
	}
	generateVersion func() string
)

func versionGenerator() string {
	// tfe appears to use 32 hex-encoded random bytes for its version
	// ID, so OTF does the same
	b := make([]byte, 32)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}

func (f *factory) new(opts CreateVariableOptions) (*Variable, error) {
	v := Variable{
		ID:        internal.NewID("var"),
		VersionID: f.generateVersion(),
	}

	// Required fields
	if opts.Key == nil {
		return nil, errors.New("missing key")
	}
	if err := v.setKey(*opts.Key); err != nil {
		return nil, err
	}
	if opts.Category == nil {
		return nil, errors.New("missing category")
	}
	if err := v.setCategory(*opts.Category); err != nil {
		return nil, err
	}

	// Optional fields
	if opts.Value != nil {
		if err := v.setValue(*opts.Value); err != nil {
			return nil, err
		}
	}
	if opts.Description != nil {
		if err := v.setDescription(*opts.Description); err != nil {
			return nil, err
		}
	}
	if opts.Sensitive != nil {
		v.Sensitive = *opts.Sensitive
	}
	if opts.HCL != nil {
		v.HCL = *opts.HCL
	}

	return &v, nil
}

func (f *factory) newWorkspaceVariable(workspaceID string, opts CreateVariableOptions) (*WorkspaceVariable, error) {
	v, err := f.new(opts)
	if err != nil {
		return nil, err
	}
	return &WorkspaceVariable{
		Variable:    v,
		WorkspaceID: workspaceID,
	}, nil
}

func (f *factory) newSet(organization string, opts CreateVariableSetOptions) (*VariableSet, error) {
	set := VariableSet{
		ID:          internal.NewID("varset"),
		Name:        opts.Name,
		Description: opts.Description,
		Global:      opts.Global,
	}
	return &set, nil
}

func (f *factory) update(v *Variable, opts UpdateVariableOptions) error {
	if opts.Key != nil {
		if v.Sensitive {
			return errors.New("changing the key of a sensitive variable is not allowed")
		}
		if err := v.setKey(*opts.Key); err != nil {
			return err
		}
	}
	if opts.Value != nil {
		if err := v.setValue(*opts.Value); err != nil {
			return err
		}
	}
	if opts.Description != nil {
		if err := v.setDescription(*opts.Description); err != nil {
			return err
		}
	}
	if opts.Category != nil {
		if err := v.setCategory(*opts.Category); err != nil {
			return err
		}
	}
	if opts.HCL != nil {
		if v.Sensitive {
			return errors.New("changing HCL mode on a sensitive variable is not allowed")
		}
		v.HCL = *opts.HCL
	}
	if opts.Sensitive != nil {
		if err := v.setSensitive(*opts.Sensitive); err != nil {
			return err
		}
	}
	// generate new version ID on every update call, even if nothing is actually
	// updated
	v.VersionID = f.generateVersion()
	return nil
}
