package jsonapi

// Variable represents a Terraform Enterprise variable.
type Variable struct {
	ID          string `jsonapi:"primary,vars"`
	Key         string `jsonapi:"attr,key"`
	Value       string `jsonapi:"attr,value"`
	Description string `jsonapi:"attr,description"`
	Category    string `jsonapi:"attr,category"`
	HCL         bool   `jsonapi:"attr,hcl"`
	Sensitive   bool   `jsonapi:"attr,sensitive"`

	// Relations
	Workspace *Workspace `jsonapi:"relation,configurable"`
}

// VariableList represents a list of variables.
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
	Key *string `jsonapi:"attr,key"`

	// The value of the variable.
	Value *string `jsonapi:"attr,value,omitempty"`

	// The description of the variable.
	Description *string `jsonapi:"attr,description,omitempty"`

	// Whether this is a Terraform or environment variable.
	Category *string `jsonapi:"attr,category"`

	// Whether to evaluate the value of the variable as a string of HCL code.
	HCL *bool `jsonapi:"attr,hcl,omitempty"`

	// Whether the value is sensitive.
	Sensitive *bool `jsonapi:"attr,sensitive,omitempty"`
}

// VariableUpdateOptions represents the options for updating a variable.
type VariableUpdateOptions struct {
	// Type is a public field utilized by JSON:API to
	// set the resource type via the field tag.
	// It is not a user-defined value and does not need to be set.
	// https://jsonapi.org/format/#crud-creating
	Type string `jsonapi:"primary,vars"`

	// The name of the variable.
	Key *string `jsonapi:"attr,key,omitempty"`

	// The value of the variable.
	Value *string `jsonapi:"attr,value,omitempty"`

	// The description of the variable.
	Description *string `jsonapi:"attr,description,omitempty"`

	// Whether this is a Terraform or environment variable.
	Category *string `jsonapi:"attr,category"`

	// Whether to evaluate the value of the variable as a string of HCL code.
	HCL *bool `jsonapi:"attr,hcl,omitempty"`

	// Whether the value is sensitive.
	Sensitive *bool `jsonapi:"attr,sensitive,omitempty"`
}
