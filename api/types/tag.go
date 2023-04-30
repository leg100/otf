// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package types

type (
	// OrganizationTag represents a Terraform Enterprise Organization tag
	OrganizationTag struct {
		ID string `jsonapi:"primary,tags"`

		// Optional:
		Name string `jsonapi:"attribute" json:"name,omitempty"`

		// Optional: Number of workspaces that have this tag
		InstanceCount int `jsonapi:"attribute" json:"instance-count,omitempty"`

		// The org this tag belongs to
		Organization *Organization `jsonapi:"relationship" json:"organization"`
	}

	DeleteTagOption struct {
		ID string `jsonapi:"primary,tags"`
	}
)
