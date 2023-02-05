// Package variable is responsible for workspace variables
package variable

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/leg100/otf"
	"github.com/leg100/otf/http/jsonapi"
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
)

// VariableCategoryPtr returns a pointer to the given category type.
func VariableCategoryPtr(v otf.VariableCategory) *otf.VariableCategory {
	return &v
}

type Variable struct {
	id          string
	key         string
	value       string
	description string
	category    otf.VariableCategory
	sensitive   bool
	hcl         bool
	workspaceID string
}

func NewVariable(workspaceID string, opts otf.CreateVariableOptions) (*Variable, error) {
	v := Variable{
		id:          otf.NewID("var"),
		workspaceID: workspaceID,
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
		v.sensitive = *opts.Sensitive
	}
	if opts.HCL != nil {
		v.hcl = *opts.HCL
	}

	return &v, nil
}

func (v *Variable) ID() string                     { return v.id }
func (v *Variable) Key() string                    { return v.key }
func (v *Variable) Value() string                  { return v.value }
func (v *Variable) Description() string            { return v.description }
func (v *Variable) Category() otf.VariableCategory { return v.category }
func (v *Variable) Sensitive() bool                { return v.sensitive }
func (v *Variable) HCL() bool                      { return v.hcl }
func (v *Variable) WorkspaceID() string            { return v.workspaceID }

func (v *Variable) MarshalLog() any {
	log := struct {
		ID          string `json:"id"`
		Key         string `json:"key"`
		Value       string `json:"value"`
		Sensitive   bool   `json:"sensitive"`
		WorkspaceID string `json:"workspace_id"`
	}{
		ID:          v.id,
		Key:         v.key,
		Value:       v.value,
		Sensitive:   v.sensitive,
		WorkspaceID: v.workspaceID,
	}
	if v.sensitive {
		log.Value = "*****"
	}
	return log
}

func (v *Variable) Update(opts otf.UpdateVariableOptions) error {
	if opts.Key != nil {
		if v.sensitive {
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
		if v.sensitive {
			return errors.New("changing the category of a sensitive variable is not allowed")
		}
		if err := v.setCategory(*opts.Category); err != nil {
			return err
		}
	}
	if opts.HCL != nil {
		if v.sensitive {
			return errors.New("toggling HCL mode on a sensitive variable is not allowed")
		}
		v.hcl = *opts.HCL
	}
	if opts.Sensitive != nil {
		if err := v.setSensitive(*opts.Sensitive); err != nil {
			return err
		}
	}
	return nil
}

func (v *Variable) ToJSONAPI() any {
	to := jsonapiVariable{
		ID:          v.ID(),
		Key:         v.Key(),
		Value:       v.Value(),
		Description: v.Description(),
		Category:    string(v.Category()),
		Sensitive:   v.Sensitive(),
		HCL:         v.HCL(),
		Workspace: &jsonapi.Workspace{
			ID: v.WorkspaceID(),
		},
	}
	if to.Sensitive {
		to.Value = "" // scrub sensitive values
	}
	return &to
}

func (v *Variable) setKey(key string) error {
	if len(key) > VariableKeyMaxChars {
		return ErrVariableKeyMaxExceeded
	}
	v.key = strings.TrimSpace(key)
	return nil
}

func (v *Variable) setValue(value string) error {
	if len([]byte(value)) > (VariableValueMaxKB * 1024) {
		return ErrVariableValueMaxExceeded
	}
	v.value = strings.TrimSpace(value)
	return nil
}

func (v *Variable) setDescription(desc string) error {
	if len(desc) > VariableDescriptionMaxChars {
		return ErrVariableDescriptionMaxExceeded
	}
	v.description = desc
	return nil
}

func (v *Variable) setCategory(cat otf.VariableCategory) error {
	if cat != otf.CategoryEnv && cat != otf.CategoryTerraform {
		return errors.New("invalid variable category")
	}

	v.category = cat
	return nil
}

func (v *Variable) setSensitive(sensitive bool) error {
	if v.sensitive && !sensitive {
		return errors.New("cannot change a sensitive variable to a non-sensitive variable")
	}
	v.sensitive = sensitive
	return nil
}

type service interface {
	create(ctx context.Context, workspaceID string, opts otf.CreateVariableOptions) (*Variable, error)
	list(ctx context.Context, workspaceID string) ([]*Variable, error)
	get(ctx context.Context, variableID string) (*Variable, error)
	update(ctx context.Context, variableID string, opts otf.UpdateVariableOptions) (*Variable, error)
	delete(ctx context.Context, variableID string) (*Variable, error)
}

type VariableStore interface {
	CreateVariable(ctx context.Context, variable *Variable) error
	ListVariables(ctx context.Context, workspaceID string) ([]*Variable, error)
	GetVariable(ctx context.Context, variableID string) (*Variable, error)
	UpdateVariable(ctx context.Context, variableID string, updateFn func(*Variable) error) (*Variable, error)
	DeleteVariable(ctx context.Context, variableID string) (*Variable, error)
}

func UnmarshalVariableJSONAPI(json *jsonapiVariable) *Variable {
	return &Variable{
		id:          json.ID,
		key:         json.Key,
		value:       json.Value,
		description: json.Description,
		category:    otf.VariableCategory(json.Category),
		sensitive:   json.Sensitive,
		hcl:         json.HCL,
		workspaceID: json.Workspace.ID,
	}
}

// UnmarshalVariableListJSONAPI converts a DTO into a workspace list
func UnmarshalVariableListJSONAPI(json *jsonapiList) []*Variable {
	var list []*Variable
	for _, i := range json.Items {
		list = append(list, UnmarshalVariableJSONAPI(i))
	}
	return list
}
