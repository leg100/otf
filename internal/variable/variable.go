// Package variable manages terraform workspace variables
package variable

import (
	"errors"
	"fmt"
	"os"
	"path"
	"slices"
	"strings"

	"log/slog"

	"github.com/leg100/otf/internal/run"
	"golang.org/x/exp/maps"
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
	ErrVariableConflict               = errors.New("variable conflicts with a variable with the same name and type")
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

		// OTF doesn't use this internally but the go-tfe integration tests
		// expect it to be a random value that changes on every update.
		VersionID string
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

func (v *Variable) LogValue() slog.Value {
	attrs := []slog.Attr{
		slog.String("id", v.ID),
		slog.String("key", v.Key),
		slog.Bool("sensitive", v.Sensitive),
	}
	if v.Sensitive {
		attrs = append(attrs, slog.String("value", "*****"))
	} else {
		attrs = append(attrs, slog.String("value", v.Value))
	}
	return slog.GroupValue(attrs...)
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

func (v *Variable) conflicts(v2 *Variable) error {
	if v.ID == v2.ID {
		// cannot conflict with self
		return nil
	}
	if v.Key == v2.Key && v.Category == v2.Category {
		return ErrVariableConflict
	}
	return nil
}

// WriteTerraformVars writes workspace variables to a file named
// terraform.tfvars located in the given path. If the file already exists it'll
// be appended to.
func WriteTerraformVars(dir string, vars []*Variable) error {
	path := path.Join(dir, "terraform.tfvars")
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0o600)
	if err != nil {
		return err
	}
	defer f.Close()

	var b strings.Builder
	// lazily start with a new line in case user uploaded terraform.tfvars with
	// content already
	b.WriteRune('\n')
	for _, v := range vars {
		if v.Category == CategoryTerraform {
			b.WriteString(v.Key)
			b.WriteString(" = ")
			if v.HCL {
				b.WriteString(v.Value)
			} else {
				b.WriteString(`"`)
				b.WriteString(v.Value)
				b.WriteString(`"`)
			}
			b.WriteRune('\n')
		}
	}
	f.WriteString(b.String())

	return nil
}

// mergeVariables merges variables for a run according to the precedence rules
// documented here:
//
// https://developer.hashicorp.com/terraform/cloud-docs/workspaces/variables#precedence
func mergeVariables(run *run.Run, workspaceVariables []*WorkspaceVariable, sets []*VariableSet) []*Variable {
	// terraform variables keyed by variable key
	tfVars := make(map[string]*Variable)
	// environment variables keyed by variable key
	envVars := make(map[string]*Variable)

	// global sets have lowest precedence
	for _, s := range sets {
		if s.Global {
			for _, v := range s.Variables {
				switch v.Category {
				case CategoryTerraform:
					tfVars[v.Key] = v
				case CategoryEnv:
					envVars[v.Key] = v
				}
			}
		}
	}

	// workspace-scoped sets have next lowest precedence; lexical order of the
	// set name determines precedence if a variable with the same key is found
	// in more than one set, e.g. variable foo in set named A takes precedence
	// over variable foo in set named B.
	//
	// sort sets by lexical order, A->Z
	slices.SortFunc(sets, func(a, b *VariableSet) int {
		if a.Name < b.Name {
			return -1
		} else if a.Name > b.Name {
			return 1
		} else {
			return 0
		}
	})
	// reverse order sets (Z->A), so that sets later in the slice take precedence.
	slices.Reverse(sets)
	for _, s := range sets {
		if slices.Contains(s.Workspaces, run.WorkspaceID) {
			for _, v := range s.Variables {
				switch v.Category {
				case CategoryTerraform:
					tfVars[v.Key] = v
				case CategoryEnv:
					envVars[v.Key] = v
				}
			}
		}
	}

	// workspace variables have higher precedence than sets, so override
	// anything from sets
	for _, v := range workspaceVariables {
		switch v.Category {
		case CategoryTerraform:
			tfVars[v.Key] = v.Variable
		case CategoryEnv:
			envVars[v.Key] = v.Variable
		}
	}

	// run variables have highest precedence
	for _, v := range run.Variables {
		tfVars[v.Key] = &Variable{Key: v.Key, Value: v.Value, Category: CategoryTerraform, HCL: true}
	}

	return append(maps.Values(tfVars), maps.Values(envVars)...)
}
