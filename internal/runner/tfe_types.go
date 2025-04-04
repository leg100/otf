// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package runner

import (
	"time"

	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/tfeapi/types"
	"github.com/leg100/otf/internal/workspace"
)

// TFEAgentPool represents a Terraform Cloud agent pool.
type TFEAgentPool struct {
	ID                 resource.TfeID `jsonapi:"primary,agent-pools"`
	Name               string         `jsonapi:"attribute" json:"name"`
	AgentCount         int            `jsonapi:"attribute" json:"agent-count"`
	OrganizationScoped bool           `jsonapi:"attribute" json:"organization-scoped"`

	// Relations
	Organization      *organization.TFEOrganization `jsonapi:"relationship" json:"organization"`
	Workspaces        []*workspace.TFEWorkspace     `jsonapi:"relationship" json:"workspaces"`
	AllowedWorkspaces []*workspace.TFEWorkspace     `jsonapi:"relationship" json:"allowed-workspaces"`
}

// TFEAgentPoolCreateOptions represents the options for creating an agent pool.
type TFEAgentPoolCreateOptions struct {
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

// TFEAgentPoolUpdateOptions represents the options for updating an agent pool.
type TFEAgentPoolUpdateOptions struct {
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

// TFEAgentPoolListOptions represents the options for listing agent pools.
type TFEAgentPoolListOptions struct {
	types.ListOptions
	// Optional: A list of relations to include. See available resources
	// https://developer.hashicorp.com/terraform/cloud-docs/api-docs/agents#available-related-resources
	// Include []AgentPoolIncludeOpt `url:"include,omitempty"`

	// Optional: A search query string used to filter agent pool. Agent pools are searchable by name
	Query *string `schema:"q,omitempty"`

	// Optional: String (workspace name) used to filter the results.
	AllowedWorkspacesName *string `schema:"filter[allowed_workspaces][name],omitempty"`
}

// TFEAgentToken represents a TFE agent token.
type TFEAgentToken struct {
	ID          resource.TfeID `jsonapi:"primary,authentication-tokens"`
	CreatedAt   time.Time      `jsonapi:"attribute" json:"created-at"`
	Description string         `jsonapi:"attribute" json:"description"`
	LastUsedAt  time.Time      `jsonapi:"attribute" json:"last-used-at"`
	Token       string         `jsonapi:"attribute" json:"token"`
}

// TFEAgentTokenCreateOptions represents the options for creating a new otf agent token.
type TFEAgentTokenCreateOptions struct {
	// Type is a public field utilized by JSON:API to set the resource type via
	// the field tag.  It is not a user-defined value and does not need to be
	// set.  https://jsonapi.org/format/#crud-creating
	Type string `jsonapi:"primary,agent-tokens"`

	// Description is a meaningful description of the purpose of the agent
	// token.
	Description string `jsonapi:"attribute" json:"description"`
}
