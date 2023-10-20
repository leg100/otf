// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package types

import "time"

// TeamToken represents a Terraform Enterprise team token.
type TeamToken struct {
	ID          string     `jsonapi:"primary,authentication-tokens"`
	CreatedAt   time.Time  `jsonapi:"attribute" json:"created-at"`
	Description string     `jsonapi:"attribute" json:"description"`
	LastUsedAt  time.Time  `jsonapi:"attribute" json:"last-used-at"`
	Token       string     `jsonapi:"attribute" json:"token"`
	ExpiredAt   *time.Time `jsonapi:"attribute" json:"expired-at"`
}

// TeamTokenCreateOptions contains the options for creating a team token.
type TeamTokenCreateOptions struct {
	// Optional: The token's expiration date.
	// This feature is available in TFE release v202305-1 and later
	ExpiredAt *time.Time `jsonapi:"attribute" json:"expired-at,omitempty"`
}
