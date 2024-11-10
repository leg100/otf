// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package types

import "github.com/leg100/otf/internal/resource"

type (
	// OrganizationTag represents a Terraform Enterprise Organization tag
	OrganizationTag struct {
		ID resource.ID `jsonapi:"primary,tags"`

		// Optional:
		Name string `jsonapi:"attribute" json:"name,omitempty"`

		// Optional: Number of workspaces that have this tag
		InstanceCount int `jsonapi:"attribute" json:"instance-count,omitempty"`

		// The org this tag belongs to
		Organization *Organization `jsonapi:"relationship" json:"organization"`
	}

	// TagSpec is owned by an organization and applied to workspaces. Used for
	// grouping and search. Only one of ID or name must be specified.
	TagSpec struct {
		// Type is a public field utilized by JSON:API to set the resource type via
		// the field tag.  It is not a user-defined value and does not need to be
		// set.  https://jsonapi.org/format/#crud-creating
		Type string `jsonapi:"primary,tags"`

		ID   *resource.ID `jsonapi:"attribute" json:"id"`
		Name string       `jsonapi:"attribute" json:"name,omitempty"`
	}
)
