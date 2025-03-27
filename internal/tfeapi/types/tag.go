// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package types

import (
	"encoding"
	"errors"

	"github.com/leg100/otf/internal/resource"
)

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

	// Tag is owned by an organization and applied to workspaces. Used for
	// grouping and search. Only one of ID or name must be specified.
	Tag struct {
		ID   resource.ID `jsonapi:"primary,tags"`
		Name string      `jsonapi:"attribute" json:"name,omitempty"`
	}
)

// UnmarshalID helps datadog/jsonapi to unmarshal the ID in a serialized tag -
// either the ID or the name is set, and datadog/jsonapi otherwise gets upset
// when ID is unset.
func (t *Tag) UnmarshalID(id string) error {
	if id == "" {
		return nil
	}
	unmarshalable, ok := t.ID.(encoding.TextUnmarshaler)
	if !ok {
		return errors.New("id is not unmarshalable")
	}
	return unmarshalable.UnmarshalText([]byte(id))
}
