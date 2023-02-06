package variable

import (
	"github.com/leg100/otf"
	"github.com/leg100/otf/http/jsonapi"
)

// jsonapiVariable is a variable suitable for marshaling into jsonapi
type jsonapiVariable struct {
	ID          string `jsonapi:"primary,vars"`
	Key         string `jsonapi:"attr,key"`
	Value       string `jsonapi:"attr,value"`
	Description string `jsonapi:"attr,description"`
	Category    string `jsonapi:"attr,category"`
	HCL         bool   `jsonapi:"attr,hcl"`
	Sensitive   bool   `jsonapi:"attr,sensitive"`

	// Relations
	Workspace *jsonapi.Workspace `jsonapi:"relation,configurable"`
}

func (j *jsonapiVariable) toVariable() *Variable {
	return &Variable{
		id:          j.ID,
		key:         j.Key,
		value:       j.Value,
		description: j.Description,
		category:    otf.VariableCategory(j.Category),
		sensitive:   j.Sensitive,
		hcl:         j.HCL,
		workspaceID: j.Workspace.ID,
	}
}

// jsonapiList represents a list of variables.
type jsonapiList struct {
	*jsonapi.Pagination
	Items []*jsonapiVariable
}

// VariableCreateOptions represents the options for creating a new variable.
type jsonapiVariableCreateOptions struct {
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
type jsonapiVariableUpdateOptions struct {
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
