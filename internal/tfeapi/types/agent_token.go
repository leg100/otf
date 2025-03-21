// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package types

import (
	"time"

	"github.com/leg100/otf/internal/resource"
)

// AgentToken represents a TFE agent token.
type AgentToken struct {
	ID          resource.TfeID `jsonapi:"primary,authentication-tokens"`
	CreatedAt   time.Time      `jsonapi:"attribute" json:"created-at"`
	Description string         `jsonapi:"attribute" json:"description"`
	LastUsedAt  time.Time      `jsonapi:"attribute" json:"last-used-at"`
	Token       string         `jsonapi:"attribute" json:"token"`
}

// AgentTokenCreateOptions represents the options for creating a new otf agent token.
type AgentTokenCreateOptions struct {
	// Type is a public field utilized by JSON:API to set the resource type via
	// the field tag.  It is not a user-defined value and does not need to be
	// set.  https://jsonapi.org/format/#crud-creating
	Type string `jsonapi:"primary,agent-tokens"`

	// Description is a meaningful description of the purpose of the agent
	// token.
	Description string `jsonapi:"attribute" json:"description"`
}
