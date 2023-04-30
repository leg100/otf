// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package types

type (

	// OrganizationTagsList represents a list of organization tags
	OrganizationTagsList struct {
		*Pagination
		Items []*OrganizationTag
	}

	// OrganizationTag represents a Terraform Enterprise Organization tag
	OrganizationTag struct {
		ID string `jsonapi:"primary,tags"`
		// Optional:
		Name string `jsonapi:"attr,name,omitempty"`

		// Optional: Number of workspaces that have this tag
		InstanceCount int `jsonapi:"attr,instance-count,omitempty"`

		// The org this tag belongs to
		Organization *Organization `jsonapi:"relation,organization"`
	}

	// Tag is owned by an organization and applied to workspaces. Used for grouping and search.
	Tag struct {
		ID   string `jsonapi:"primary,tags"`
		Name string `jsonapi:"attr,name,omitempty"`
	}

	// OrganizationTagsDeleteOptions represents the request body for deleting a tag in an organization
	OrganizationTagsDeleteOptions struct {
		IDs []string // Required
	}

	// AddWorkspacesToTagOptions represents the request body to add a workspace to a tag
	AddWorkspacesToTagOptions struct {
		WorkspaceIDs []string // Required
	}

	// Tags is a list of tags to be added to a workspace, or removed from a
	// workspace.
	Tags struct {
		Tags []*Tag
	}
)
