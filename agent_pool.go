package ots

// AgentPool represents a Terraform Cloud agent pool.
type AgentPool struct {
	ID   string `jsonapi:"primary,agent-pools"`
	Name string `jsonapi:"attr,name"`

	// Relations
	Organization *Organization `jsonapi:"relation,organization"`
}
