package ots

import tfe "github.com/leg100/go-tfe"

// AgentPool represents a Terraform Cloud agent pool.
type AgentPool struct {
	ID   string `jsonapi:"primary,agent-pools"`
	Name string `jsonapi:"attr,name"`

	// Relations
	Organization *tfe.Organization `jsonapi:"relation,organization"`
}
