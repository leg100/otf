package jsonapi

// AgentToken is an otf agent token. The token field is only populated upon
// creation and never thereafter.
type AgentToken struct {
	ID           string  `jsonapi:"primary,agent_tokens"`
	Token        *string `jsonapi:"attr,token,omitempty"`
	Organization string  `jsonapi:"attr,organization_name"`
}

// AgentTokenCreateOptions represents the options for creating a new otf agent token.
type AgentTokenCreateOptions struct {
	// Type is a public field utilized by JSON:API to set the resource type via
	// the field tag.  It is not a user-defined value and does not need to be
	// set.  https://jsonapi.org/format/#crud-creating
	Type string `jsonapi:"primary,agent_tokens"`

	// Description is a meaningful description of the purpose of the agent
	// token.
	Description string `jsonapi:"attr,description"`

	// Organization is the name of the organization in which to create the
	// token.
	Organization string `jsonapi:"attr,organization_name"`
}
