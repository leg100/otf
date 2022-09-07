package dto

// AgentToken represents an otf agent token
type AgentToken struct {
	ID               string `jsonapi:"primary,agent_tokens"`
	OrganizationName string `jsonapi:"attr,organization_name"`
}
