// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package types

import "github.com/leg100/otf/internal/resource"

// VariableSet represents a Terraform Enterprise variable set.
type VariableSet struct {
	ID          resource.ID `jsonapi:"primary,varsets"`
	Name        string      `jsonapi:"attribute" json:"name"`
	Description string      `jsonapi:"attribute" json:"description"`
	Global      bool        `jsonapi:"attribute" json:"global"`

	// Relations
	Organization *Organization          `jsonapi:"relationship" json:"organization"`
	Workspaces   []*Workspace           `jsonapi:"relationship" json:"workspaces,omitempty"`
	Variables    []*VariableSetVariable `jsonapi:"relationship" json:"vars,omitempty"`

	// Projects     []*Project             `jsonapi:"relationship" json:"projects,omitempty"`
}

type VariableSetVariable struct {
	*Variable

	// Relations
	VariableSet *VariableSet `jsonapi:"relationship" json:"varset"`
}

// VariableSetCreateOptions represents the options for creating a new variable set within in a organization.
type VariableSetCreateOptions struct {
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

// VariableSetUpdateOptions represents the options for updating a variable set.
type VariableSetUpdateOptions struct {
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

// VariableSetVariableCreatOptions represents the options for creating a new variable within a variable set
type VariableSetVariableCreateOptions struct {
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

// VariableSetVariableUpdateOptions represents the options for updating a variable.
type VariableSetVariableUpdateOptions struct {
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
