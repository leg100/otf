package types

import "github.com/leg100/otf/internal/resource"

type (
	// User represents an OTF user.
	User struct {
		ID               resource.ID `jsonapi:"primary,users"`
		AvatarURL        string      `jsonapi:"attribute" json:"avatar-url"`
		Email            string      `jsonapi:"attribute" json:"email"`
		IsServiceAccount bool        `jsonapi:"attribute" json:"is-service-account"`
		TwoFactor        *TwoFactor  `jsonapi:"attribute" json:"two-factor"`
		UnconfirmedEmail string      `jsonapi:"attribute" json:"unconfirmed-email"`
		Username         string      `jsonapi:"attribute" json:"username"`
		V2Only           bool        `jsonapi:"attribute" json:"v2-only"`

		// Relations
		// AuthenticationTokens *AuthenticationTokens `jsonapi:"relation,authentication-tokens"`
	}

	// TwoFactor represents the organization permissions.
	TwoFactor struct {
		Enabled  bool `jsonapi:"attribute" json:"enabled"`
		Verified bool `jsonapi:"attribute" json:"verified"`
	}

	// CreateUserOptions represents the options for creating a
	// user.
	CreateUserOptions struct {
		// Type is a public field utilized by JSON:API to
		// set the resource type via the field tag.
		// It is not a user-defined value and does not need to be set.
		// https://jsonapi.org/format/#crud-creating
		Type string `jsonapi:"primary,users"`

		Username *string `jsonapi:"attribute" json:"username"`
	}
)
