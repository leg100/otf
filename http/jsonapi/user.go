package jsonapi

type (
	// User represents an OTF user.
	User struct {
		ID               string     `jsonapi:"primary,users"`
		AvatarURL        string     `jsonapi:"attr,avatar-url"`
		Email            string     `jsonapi:"attr,email"`
		IsServiceAccount bool       `jsonapi:"attr,is-service-account"`
		TwoFactor        *TwoFactor `jsonapi:"attr,two-factor"`
		UnconfirmedEmail string     `jsonapi:"attr,unconfirmed-email"`
		Username         string     `jsonapi:"attr,username"`
		V2Only           bool       `jsonapi:"attr,v2-only"`

		// Relations
		// AuthenticationTokens *AuthenticationTokens `jsonapi:"relation,authentication-tokens"`
	}

	// TwoFactor represents the organization permissions.
	TwoFactor struct {
		Enabled  bool `jsonapi:"attr,enabled"`
		Verified bool `jsonapi:"attr,verified"`
	}

	// CreateUserOptions represents the options for creating a
	// user.
	CreateUserOptions struct {
		// Type is a public field utilized by JSON:API to
		// set the resource type via the field tag.
		// It is not a user-defined value and does not need to be set.
		// https://jsonapi.org/format/#crud-creating
		Type string `jsonapi:"primary,users"`

		Username *string `jsonapi:"attr,username"`
	}
)
