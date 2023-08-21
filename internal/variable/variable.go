// Package variable manages terraform workspace variables
package variable

import (
	"errors"
	"fmt"
	"os"
	"path"
	"strings"

	"log/slog"
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
