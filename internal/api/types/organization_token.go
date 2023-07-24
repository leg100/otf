// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package types

import "time"

// OrganizationToken represents a Terraform Enterprise organization token.
type OrganizationToken struct {
	ID        string     `jsonapi:"primary,authentication-tokens"`
	CreatedAt time.Time  `jsonapi:"attribute" json:"created-at"`
	Token     string     `jsonapi:"attribute" json:"token"`
	ExpiredAt *time.Time `jsonapi:"attribute" json:"expired-at"`
}

// OrganizationTokenCreateOptions contains the options for creating an organization token.
type OrganizationTokenCreateOptions struct {
	// Optional: The token's expiration date.
	// This feature is available in TFE release v202305-1 and later
	ExpiredAt *time.Time `jsonapi:"attribute" json:"expired-at,omitempty"`
}
