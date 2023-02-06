package agenttoken

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
