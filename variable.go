package otf

import (
	"errors"
	"fmt"
	"strings"
)

// VariableCategory is the category of variable
type VariableCategory string

// VariableCategoryPtr returns a pointer to the given category type.
func VariableCategoryPtr(v VariableCategory) *VariableCategory {
	return &v
}

const (
	// https://developer.hashicorp.com/terraform/cloud-docs/workspaces/variables/managing-variables#character-limits
	VariableDescriptionMaxChars = 512
	VariableKeyMaxChars         = 128
	VariableValueMaxKB          = 256 // 256*1024 bytes

	CategoryTerraform VariableCategory = "terraform"
	CategoryEnv       VariableCategory = "env"
)

var (
	ErrVariableDescriptionMaxExceeded = fmt.Errorf("maximum variable description size (%d chars) exceeded", VariableDescriptionMaxChars)
	ErrVariableKeyMaxExceeded         = fmt.Errorf("maximum variable key size (%d chars) exceeded", VariableKeyMaxChars)
	ErrVariableValueMaxExceeded       = fmt.Errorf("maximum variable value size of %d KB exceeded", VariableValueMaxKB)
)

type (
	Variable struct {
		ID          string
		Key         string
		Value       string
		Description string
		Category    VariableCategory
		Sensitive   bool
		HCL         bool
		WorkspaceID string
	}
	CreateVariableOptions struct {
		Key         *string
		Value       *string
		Description *string
		Category    *VariableCategory
		Sensitive   *bool
		HCL         *bool
	}
	UpdateVariableOptions struct {
		Key         *string
		Value       *string
		Description *string
		Category    *VariableCategory
		Sensitive   *bool
		HCL         *bool
	}
)

func NewVariable(workspaceID string, opts CreateVariableOptions) (*Variable, error) {
	v := Variable{
		ID:          NewID("var"),
		WorkspaceID: workspaceID,
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

func (v *Variable) MarshalLog() any {
	log := struct {
		ID          string `json:"id"`
		Key         string `json:"key"`
		Value       string `json:"value"`
		Sensitive   bool   `json:"sensitive"`
		WorkspaceID string `json:"workspace_id"`
	}{
		ID:          v.ID,
		Key:         v.Key,
		Value:       v.Value,
		Sensitive:   v.Sensitive,
		WorkspaceID: v.WorkspaceID,
	}
	if v.Sensitive {
		log.Value = "*****"
	}
	return log
}

func (v *Variable) Update(opts UpdateVariableOptions) error {
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
		if v.Sensitive {
			return errors.New("changing the category of a sensitive variable is not allowed")
		}
		if err := v.setCategory(*opts.Category); err != nil {
			return err
		}
	}
	if opts.HCL != nil {
		if v.Sensitive {
			return errors.New("toggling HCL mode on a sensitive variable is not allowed")
		}
		v.HCL = *opts.HCL
	}
	if opts.Sensitive != nil {
		if err := v.setSensitive(*opts.Sensitive); err != nil {
			return err
		}
	}
	return nil
}

func (v *Variable) setKey(key string) error {
	if len(key) > VariableKeyMaxChars {
		return ErrVariableKeyMaxExceeded
	}
	v.Key = strings.TrimSpace(key)
	return nil
}

func (v *Variable) setValue(value string) error {
	if len([]byte(value)) > (VariableValueMaxKB * 1024) {
		return ErrVariableValueMaxExceeded
	}
	v.Value = strings.TrimSpace(value)
	return nil
}

func (v *Variable) setDescription(desc string) error {
	if len(desc) > VariableDescriptionMaxChars {
		return ErrVariableDescriptionMaxExceeded
	}
	v.Description = desc
	return nil
}

func (v *Variable) setCategory(cat VariableCategory) error {
	if cat != CategoryEnv && cat != CategoryTerraform {
		return errors.New("invalid variable category")
	}

	v.Category = cat
	return nil
}

func (v *Variable) setSensitive(sensitive bool) error {
	if v.Sensitive && !sensitive {
		return errors.New("cannot change a sensitive variable to a non-sensitive variable")
	}
	v.Sensitive = sensitive
	return nil
}
