// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package runner

import (
	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/tfeapi/types"
	"github.com/leg100/otf/internal/workspace"
)

// AgentPool represents a Terraform Cloud agent pool.
type AgentPool struct {
	ID                 resource.TfeID `jsonapi:"primary,agent-pools"`
	Name               string         `jsonapi:"attribute" json:"name"`
	AgentCount         int            `jsonapi:"attribute" json:"agent-count"`
	OrganizationScoped bool           `jsonapi:"attribute" json:"organization-scoped"`

	// Relations
	Organization      *organization.TFEOrganization `jsonapi:"relationship" json:"organization"`
	Workspaces        []*workspace.TFEWorkspace     `jsonapi:"relationship" json:"workspaces"`
	AllowedWorkspaces []*workspace.TFEWorkspace     `jsonapi:"relationship" json:"allowed-workspaces"`
}

// AgentPoolCreateOptions represents the options for creating an agent pool.
type AgentPoolCreateOptions struct {
	// Type is a public field utilized by JSON:API to
	// set the resource type via the field tag.
	// It is not a user-defined value and does not need to be set.
	// https://jsonapi.org/format/#crud-creating
	Type string `jsonapi:"primary,agent-pools"`

	// Required: A name to identify the agent pool.
	Name *string `jsonapi:"attribute" json:"name"`

	// True if the agent pool is organization scoped, false otherwise.
	OrganizationScoped *bool `jsonapi:"attribute" json:"organization-scoped,omitempty"`

	// List of workspaces that are associated with an agent pool.
	AllowedWorkspaces []*workspace.TFEWorkspace `jsonapi:"relationship" json:"allowed-workspaces,omitempty"`
}

// AgentPoolUpdateOptions represents the options for updating an agent pool.
type AgentPoolUpdateOptions struct {
	// Type is a public field utilized by JSON:API to
	// set the resource type via the field tag.
	// It is not a user-defined value and does not need to be set.
	// https://jsonapi.org/format/#crud-creating
	Type string `jsonapi:"primary,agent-pools"`

	// A new name to identify the agent pool.
	Name *string `jsonapi:"attribute" json:"name,omitempty"`

	// True if the agent pool is organization scoped, false otherwise.
	OrganizationScoped *bool `jsonapi:"attribute" json:"organization-scoped,omitempty"`

	// A new list of workspaces that are associated with an agent pool.
	AllowedWorkspaces []*workspace.TFEWorkspace `jsonapi:"relationship" json:"allowed-workspaces,omitempty"`
}

// AgentPoolListOptions represents the options for listing agent pools.
type AgentPoolListOptions struct {
	types.ListOptions
	// Optional: A list of relations to include. See available resources
	// https://developer.hashicorp.com/terraform/cloud-docs/api-docs/agents#available-related-resources
	// Include []AgentPoolIncludeOpt `url:"include,omitempty"`

	// Optional: A search query string used to filter agent pool. Agent pools are searchable by name
	Query *string `schema:"q,omitempty"`

	// Optional: String (workspace name) used to filter the results.
	AllowedWorkspacesName *string `schema:"filter[allowed_workspaces][name],omitempty"`
}
