package auth

// User represents a Terraform Enterprise user.
type jsonapiUser struct {
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
type TwoFactor struct {
	Enabled  bool `jsonapi:"attr,enabled"`
	Verified bool `jsonapi:"attr,verified"`
}

// AgentToken represents an otf agent token
type jsonapiAgentToken struct {
	ID string `jsonapi:"primary,agent_tokens"`
	// Only set upon creation and never thereafter
	Token        *string `jsonapi:"attr,token,omitempty"`
	Organization string  `jsonapi:"attr,organization_name"`
}

// AgentTokenCreateOptions represents the options for creating a new otf agent token.
type jsonapiCreateOptions struct {
	// Type is a public field utilized by JSON:API to set the resource type via
	// the field tag.  It is not a user-defined value and does not need to be
	// set.  https://jsonapi.org/format/#crud-creating
	Type string `jsonapi:"primary,runs"`

	// Description is a meaningful description of the purpose of the agent
	// token.
	Description string `jsonapi:"attr,description"`

	// Organization is the name of the organization in which to create the
	// token.
	Organization string `jsonapi:"attr,organization_name"`
}
