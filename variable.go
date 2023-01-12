package otf

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgtype"
)

// VariableCategory is the category of variable
type VariableCategory string

const (
	CategoryTerraform VariableCategory = "terraform"
	CategoryEnv       VariableCategory = "env"

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
func VariableCategoryPtr(v VariableCategory) *VariableCategory {
	return &v
}

type Variable struct {
	id          string
	key         string
	value       string
	description string
	category    VariableCategory
	sensitive   bool
	hcl         bool
	workspaceID string
}

func (v *Variable) ID() string                 { return v.id }
func (v *Variable) Key() string                { return v.key }
func (v *Variable) Value() string              { return v.value }
func (v *Variable) Description() string        { return v.description }
func (v *Variable) Category() VariableCategory { return v.category }
func (v *Variable) Sensitive() bool            { return v.sensitive }
func (v *Variable) HCL() bool                  { return v.hcl }
func (v *Variable) WorkspaceID() string        { return v.workspaceID }

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

func (v *Variable) Update(opts UpdateVariableOptions) error {
	if opts.Key != nil {
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
	if opts.Sensitive != nil {
		v.sensitive = *opts.Sensitive
	}
	if opts.HCL != nil {
		v.hcl = *opts.HCL
	}
	if opts.Category != nil {
		if err := v.setCategory(*opts.Category); err != nil {
			return err
		}
	}

	return nil
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

func (v *Variable) setCategory(cat VariableCategory) error {
	if cat != CategoryEnv && cat != CategoryTerraform {
		return errors.New("invalid variable category")
	}
	v.category = cat
	return nil
}

type VariableService interface {
	CreateVariable(ctx context.Context, workspaceID string, opts CreateVariableOptions) (*Variable, error)
	ListVariables(ctx context.Context, workspaceID string) ([]*Variable, error)
	GetVariable(ctx context.Context, variableID string) (*Variable, error)
	UpdateVariable(ctx context.Context, variableID string, opts UpdateVariableOptions) (*Variable, error)
	DeleteVariable(ctx context.Context, variableID string) (*Variable, error)
}

type VariableStore interface {
	CreateVariable(ctx context.Context, variable *Variable) error
	ListVariables(ctx context.Context, workspaceID string) ([]*Variable, error)
	GetVariable(ctx context.Context, variableID string) (*Variable, error)
	UpdateVariable(ctx context.Context, variableID string, updateFn func(*Variable) error) (*Variable, error)
	DeleteVariable(ctx context.Context, variableID string) (*Variable, error)
}

type (
	CreateVariableOptions struct {
		Key         *string `schema:"key,required"`
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
		id:          NewID("var"),
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
		if err := validateVariableDescription(*opts.Description); err != nil {
			return nil, err
		}
		v.description = *opts.Description
	}
	if opts.Sensitive != nil {
		v.sensitive = *opts.Sensitive
	}
	if opts.HCL != nil {
		v.hcl = *opts.HCL
	}

	return &v, nil
}

type VariableRow struct {
	VariableID  pgtype.Text `json:"variable_id"`
	Key         pgtype.Text `json:"key"`
	Value       pgtype.Text `json:"value"`
	Description pgtype.Text `json:"description"`
	Category    pgtype.Text `json:"category"`
	Sensitive   bool        `json:"sensitive"`
	HCL         bool        `json:"hcl"`
	WorkspaceID pgtype.Text `json:"workspace_id"`
}

func UnmarshalVariableRow(result VariableRow) *Variable {
	return &Variable{
		id:          result.VariableID.String,
		key:         result.Key.String,
		value:       result.Value.String,
		description: result.Description.String,
		category:    VariableCategory(result.Category.String),
		sensitive:   result.Sensitive,
		hcl:         result.HCL,
		workspaceID: result.WorkspaceID.String,
	}
}

func validateVariableDescription(description string) error {
	if len(description) > 500 {
		return errors.New("invalid variable category")
	}
	return nil
}
