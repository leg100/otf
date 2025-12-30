// Package variable manages terraform workspace variables
package variable

import (
	"crypto/rand"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path"
	"slices"
	"strings"

	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/run"
	"golang.org/x/exp/maps"
)

const (
	// https://developer.hashicorp.com/terraform/cloud-docs/workspaces/variables/managing-variables#character-limits
	VariableDescriptionMaxChars = 512
	VariableKeyMaxChars         = 128
	VariableValueMaxKB          = 256 // 256*1024 bytes
)

var (
	ErrVariableDescriptionMaxExceeded = fmt.Errorf("maximum variable description size (%d chars) exceeded", VariableDescriptionMaxChars)
	ErrVariableKeyMaxExceeded         = fmt.Errorf("maximum variable key size (%d chars) exceeded", VariableKeyMaxChars)
	ErrVariableValueMaxExceeded       = fmt.Errorf("maximum variable value size of %d KB exceeded", VariableValueMaxKB)
	ErrVariableConflict               = errors.New("variable conflicts with another variable with the same name and type")
)

type (
	Variable struct {
		ID          resource.TfeID   `jsonapi:"primary,variables" db:"variable_id"`
		Key         string           `jsonapi:"attribute" json:"key"`
		Value       string           `jsonapi:"attribute" json:"value"`
		Description string           `jsonapi:"attribute" json:"description"`
		Category    VariableCategory `jsonapi:"attribute" json:"category"`
		Sensitive   bool             `jsonapi:"attribute" json:"sensitive"`
		HCL         bool             `jsonapi:"attribute" json:"hcl"`

		// OTF doesn't use this internally but the go-tfe integration tests
		// expect it to be a random value that changes on every update.
		VersionID string
	}

	WorkspaceVariable struct {
		*Variable
		WorkspaceID resource.TfeID
	}

	CreateVariableOptions struct {
		Key         *string
		Value       *string
		Description *string
		Category    *VariableCategory
		Sensitive   *bool
		HCL         *bool

		generateVersion
	}

	UpdateVariableOptions struct {
		Key         *string
		Value       *string
		Description *string
		Category    *VariableCategory
		Sensitive   *bool
		HCL         *bool

		generateVersion
	}

	// generates random version IDs
	generateVersion func() string
)

func newVariable(collection []*Variable, opts CreateVariableOptions) (*Variable, error) {
	v := Variable{
		ID: resource.NewTfeID(resource.VariableKind),
	}
	if opts.generateVersion == nil {
		opts.generateVersion = versionGenerator
	}
	v.VersionID = opts.generateVersion()

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
	if err := v.checkConflicts(collection); err != nil {
		return nil, err
	}
	return &v, nil
}

func versionGenerator() string {
	// tfe appears to use 32 hex-encoded random bytes for its version
	// ID, so OTF does the same
	b := make([]byte, 32)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}

func (v *Variable) LogValue() slog.Value {
	attrs := []slog.Attr{
		slog.String("id", v.ID.String()),
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

func (v *Variable) update(collection []*Variable, opts UpdateVariableOptions) error {
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
		changed := v.HCL != *opts.HCL
		if changed && v.Sensitive {
			return errors.New("changing HCL mode on a sensitive variable is not allowed")
		}
		v.HCL = *opts.HCL
	}
	if opts.Sensitive != nil {
		if err := v.setSensitive(*opts.Sensitive); err != nil {
			return err
		}
	}
	// check for conflicts with other variables in collection
	if err := v.checkConflicts(collection); err != nil {
		return err
	}
	// generate new version ID on every update call, even if nothing is actually
	// updated
	if opts.generateVersion == nil {
		opts.generateVersion = versionGenerator
	}
	v.VersionID = opts.generateVersion()
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

// checkConflicts checks for conflicts with the given variable. i.e. they share
// same key and category.
func (v *Variable) checkConflicts(collection []*Variable) error {
	for _, v2 := range collection {
		if v.ID == v2.ID {
			// cannot conflict with self
			continue
		}
		if v.Key == v2.Key && v.Category == v2.Category {
			return ErrVariableConflict
		}
	}
	return nil
}

// Matches determines whether variable is contained in vars, i.e. shares the
// same ID.
func (v *Variable) Matches(vars []*Variable) bool {
	for _, v2 := range vars {
		if v.ID == v2.ID {
			return true
		}
	}
	return false
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
			} else if strings.Contains(v.Value, "\n") {
				delimiter := "EOT"
				for strings.Contains(v.Value, delimiter) {
					delimiter = delimiter + "T"
				}

				b.WriteString("<<" + delimiter + "\n")
				b.WriteString(v.Value)
				b.WriteString("\n" + delimiter)
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

// Merge merges variables for a run according to the precedence rules
// documented here:
//
// https://developer.hashicorp.com/terraform/cloud-docs/workspaces/variables#precedence
//
// Note: If run is nil then it is ignored.
func Merge(workspaceSets []*VariableSet, workspaceVariables []*Variable, run *run.Run) []*Variable {
	// terraform variables keyed by variable key
	tfVars := make(map[string]*Variable)
	// environment variables keyed by variable key
	envVars := make(map[string]*Variable)

	// global sets have lowest precedence
	for _, s := range workspaceSets {
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
	slices.SortFunc(workspaceSets, func(a, b *VariableSet) int {
		if a.Name < b.Name {
			return -1
		} else if a.Name > b.Name {
			return 1
		} else {
			return 0
		}
	})
	// reverse order sets (Z->A), so that sets later in the slice take precedence.
	slices.Reverse(workspaceSets)
	for _, s := range workspaceSets {
		if s.Global {
			continue
		}
		for _, v := range s.Variables {
			switch v.Category {
			case CategoryTerraform:
				tfVars[v.Key] = v
			case CategoryEnv:
				envVars[v.Key] = v
			}
		}
	}

	// workspace variables have higher precedence than sets, so override
	// anything from sets
	for _, v := range workspaceVariables {
		switch v.Category {
		case CategoryTerraform:
			tfVars[v.Key] = v
		case CategoryEnv:
			envVars[v.Key] = v
		}
	}

	// run variables have highest precedence
	if run != nil {
		for _, v := range run.Variables {
			tfVars[v.Key] = &Variable{Key: v.Key, Value: v.Value, Category: CategoryTerraform, HCL: true}
		}
	}

	return append(maps.Values(tfVars), maps.Values(envVars)...)
}
