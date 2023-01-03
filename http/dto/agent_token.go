package dto

// AgentToken represents an otf agent token
type AgentToken struct {
	ID string `jsonapi:"primary,agent_tokens"`
	// Only set upon creation and never thereafter
	Token            *string `jsonapi:"attr,token,omitempty"`
	OrganizationName string  `jsonapi:"attr,organization_name"`
}

// AgentTokenCreateOptions represents the options for creating a new otf agent token.
type AgentTokenCreateOptions struct {
	// Type is a public field utilized by JSON:API to set the resource type via
	// the field tag.  It is not a user-defined value and does not need to be
	// set.  https://jsonapi.org/format/#crud-creating
	Type string `jsonapi:"primary,runs"`

	// Description is a meaningful description of the purpose of the agent
	// token.
	Description string `jsonapi:"attr,description"`

	// OrganizationName is the name of the organization in which to create the
	// token.
	OrganizationName string `jsonapi:"attr,organization_name"`
}
