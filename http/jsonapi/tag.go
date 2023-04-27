// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package jsonapi

// OrganizationTagsList represents a list of organization tags
type OrganizationTagsList struct {
	*Pagination
	Items []*OrganizationTag
}

// OrganizationTag represents a Terraform Enterprise Organization tag
type OrganizationTag struct {
	ID string `jsonapi:"primary,tags"`
	// Optional:
	Name string `jsonapi:"attr,name,omitempty"`

	// Optional: Number of workspaces that have this tag
	InstanceCount int `jsonapi:"attr,instance-count,omitempty"`

	// The org this tag belongs to
	Organization *Organization `jsonapi:"relation,organization"`
}

// Tag is owned by an organization and applied to workspaces. Used for grouping and search.
type Tag struct {
	ID   string `jsonapi:"primary,tags"`
	Name string `jsonapi:"attr,name,omitempty"`
}

// OrganizationTagsDeleteOptions represents the request body for deleting a tag in an organization
type OrganizationTagsDeleteOptions struct {
	IDs []string // Required
}

// AddWorkspacesToTagOptions represents the request body to add a workspace to a tag
type AddWorkspacesToTagOptions struct {
	WorkspaceIDs []string // Required
}

// this represents a single tag ID
type tagID struct {
	ID string `jsonapi:"primary,tags"`
}

// this represents a single workspace ID
type workspaceID struct {
	ID string `jsonapi:"primary,workspaces"`
}
