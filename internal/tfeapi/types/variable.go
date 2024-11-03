package types

import "github.com/leg100/otf/internal/resource"

// Variable is a workspace variable.
type Variable struct {
	ID          resource.ID `jsonapi:"primary,vars"`
	Key         string      `jsonapi:"attribute" json:"key"`
	Value       string      `jsonapi:"attribute" json:"value"`
	Description string      `jsonapi:"attribute" json:"description"`
	Category    string      `jsonapi:"attribute" json:"category"`
	HCL         bool        `jsonapi:"attribute" json:"hcl"`
	Sensitive   bool        `jsonapi:"attribute" json:"sensitive"`
	VersionID   resource.ID `jsonapi:"attribute" json:"version-id"`
}

// VariableList is a list of workspace variables
type VariableList struct {
	*Pagination
	Items []*Variable
}

// VariableCreateOptions represents the options for creating a new variable.
type VariableCreateOptions struct {
	// Type is a public field utilized by JSON:API to
	// set the resource type via the field tag.
	// It is not a user-defined value and does not need to be set.
	// https://jsonapi.org/format/#crud-creating
	Type string `jsonapi:"primary,vars"`

	// The name of the variable.
	Key *string `jsonapi:"attribute" json:"key"`

	// The value of the variable.
	Value *string `jsonapi:"attribute" json:"value,omitempty"`

	// The description of the variable.
	Description *string `jsonapi:"attribute" json:"description,omitempty"`

	// Whether this is a Terraform or environment variable.
	Category *string `jsonapi:"attribute" json:"category"`

	// Whether to evaluate the value of the variable as a string of HCL code.
	HCL *bool `jsonapi:"attribute" json:"hcl,omitempty"`

	// Whether the value is sensitive.
	Sensitive *bool `jsonapi:"attribute" json:"sensitive,omitempty"`
}

// VariableUpdateOptions represents the options for updating a variable.
type VariableUpdateOptions struct {
	// Type is a public field utilized by JSON:API to
	// set the resource type via the field tag.
	// It is not a user-defined value and does not need to be set.
	// https://jsonapi.org/format/#crud-creating
	Type string `jsonapi:"primary,vars"`

	// The name of the variable.
	Key *string `jsonapi:"attribute" json:"key,omitempty"`

	// The value of the variable.
	Value *string `jsonapi:"attribute" json:"value,omitempty"`

	// The description of the variable.
	Description *string `jsonapi:"attribute" json:"description,omitempty"`

	// Whether this is a Terraform or environment variable.
	Category *string `jsonapi:"attribute" json:"category"`

	// Whether to evaluate the value of the variable as a string of HCL code.
	HCL *bool `jsonapi:"attribute" json:"hcl,omitempty"`

	// Whether the value is sensitive.
	Sensitive *bool `jsonapi:"attribute" json:"sensitive,omitempty"`
}
