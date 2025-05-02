package variable

import (
	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/tfeapi/types"
	"github.com/leg100/otf/internal/workspace"
)

// TFEWorkspaceVariable is a workspace variable.
type TFEWorkspaceVariable struct {
	*TFEVariable

	// Relations
	Workspace *workspace.TFEWorkspace `jsonapi:"relationship" json:"configurable"`
}

// TFEVariable is a workspace variable.
type TFEVariable struct {
	ID          resource.TfeID `jsonapi:"primary,vars"`
	Key         string         `jsonapi:"attribute" json:"key"`
	Value       string         `jsonapi:"attribute" json:"value"`
	Description string         `jsonapi:"attribute" json:"description"`
	Category    string         `jsonapi:"attribute" json:"category"`
	HCL         bool           `jsonapi:"attribute" json:"hcl"`
	Sensitive   bool           `jsonapi:"attribute" json:"sensitive"`
	VersionID   string         `jsonapi:"attribute" json:"version-id"`
}

// TFEVariableList is a list of workspace variables
type TFEVariableList struct {
	*types.Pagination
	Items []*Variable
}

// TFEVariableCreateOptions represents the options for creating a new variable.
type TFEVariableCreateOptions struct {
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

// TFEVariableUpdateOptions represents the options for updating a variable.
type TFEVariableUpdateOptions struct {
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

// TFEVariableSet represents a Terraform Enterprise variable set.
type TFEVariableSet struct {
	ID          resource.TfeID `jsonapi:"primary,varsets"`
	Name        string         `jsonapi:"attribute" json:"name"`
	Description string         `jsonapi:"attribute" json:"description"`
	Global      bool           `jsonapi:"attribute" json:"global"`

	// Relations
	Organization *organization.TFEOrganization `jsonapi:"relationship" json:"organization"`
	Workspaces   []*workspace.TFEWorkspace     `jsonapi:"relationship" json:"workspaces,omitempty"`
	Variables    []*TFEVariableSetVariable     `jsonapi:"relationship" json:"vars,omitempty"`

	// Projects     []*Project             `jsonapi:"relationship" json:"projects,omitempty"`
}

type TFEVariableSetVariable struct {
	*TFEVariable

	// Relations
	VariableSet *TFEVariableSet `jsonapi:"relationship" json:"varset"`
}

// TFEVariableSetCreateOptions represents the options for creating a new variable set within in a organization.
type TFEVariableSetCreateOptions struct {
	// Type is a public field utilized by JSON:API to
	// set the resource type via the field tag.
	// It is not a user-defined value and does not need to be set.
	// https://jsonapi.org/format/#crud-creating
	Type string `jsonapi:"primary,varsets"`

	// The name of the variable set.
	// Affects variable precedence when there are conflicts between Variable Sets
	// https://developer.hashicorp.com/terraform/cloud-docs/api-docs/variable-sets#apply-variable-set-to-workspaces
	Name string `jsonapi:"attribute" json:"name"`

	// A description to provide context for the variable set.
	Description string `jsonapi:"attribute" json:"description,omitempty"`

	// If true the variable set is considered in all runs in the organization.
	Global bool `jsonapi:"attribute" json:"global,omitempty"`
}

// TFEVariableSetUpdateOptions represents the options for updating a variable set.
type TFEVariableSetUpdateOptions struct {
	// Type is a public field utilized by JSON:API to
	// set the resource type via the field tag.
	// It is not a user-defined value and does not need to be set.
	// https://jsonapi.org/format/#crud-creating
	Type string `jsonapi:"primary,varsets"`

	// The name of the variable set.
	// Affects variable precedence when there are conflicts between Variable Sets
	// https://developer.hashicorp.com/terraform/cloud-docs/api-docs/variable-sets#apply-variable-set-to-workspaces
	Name *string `jsonapi:"attribute" json:"name,omitempty"`

	// A description to provide context for the variable set.
	Description *string `jsonapi:"attribute" json:"description,omitempty"`

	// If true the variable set is considered in all runs in the organization.
	Global *bool `jsonapi:"attribute" json:"global,omitempty"`
}

// TFEVariableSetVariableCreatOptions represents the options for creating a new
// variable within a variable set
type TFEVariableSetVariableCreateOptions struct {
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

// TFEVariableSetVariableUpdateOptions represents the options for updating a variable.
type TFEVariableSetVariableUpdateOptions struct {
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

	// Whether to evaluate the value of the variable as a string of HCL code.
	HCL *bool `jsonapi:"attribute" json:"hcl,omitempty"`

	// Whether the value is sensitive.
	Sensitive *bool `jsonapi:"attribute" json:"sensitive,omitempty"`
}

type TFEWorkspace struct {
	ID resource.TfeID `jsonapi:"primary,workspaces"`
}
