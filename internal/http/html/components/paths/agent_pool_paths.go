// Code generated by "go generate"; DO NOT EDIT.

package paths

import (
	"fmt"

	"github.com/a-h/templ"
)

func AgentPools(organization fmt.Stringer) templ.SafeURL {
	return templ.URL(fmt.Sprintf("/app/organizations/%s/agent-pools", organization))
}

func CreateAgentPool(organization fmt.Stringer) templ.SafeURL {
	return templ.URL(fmt.Sprintf("/app/organizations/%s/agent-pools/create", organization))
}

func NewAgentPool(organization fmt.Stringer) templ.SafeURL {
	return templ.URL(fmt.Sprintf("/app/organizations/%s/agent-pools/new", organization))
}

func AgentPool(agentPool fmt.Stringer) templ.SafeURL {
	return templ.URL(fmt.Sprintf("/app/agent-pools/%s", agentPool))
}

func EditAgentPool(agentPool fmt.Stringer) templ.SafeURL {
	return templ.URL(fmt.Sprintf("/app/agent-pools/%s/edit", agentPool))
}

func UpdateAgentPool(agentPool fmt.Stringer) templ.SafeURL {
	return templ.URL(fmt.Sprintf("/app/agent-pools/%s/update", agentPool))
}

func DeleteAgentPool(agentPool fmt.Stringer) templ.SafeURL {
	return templ.URL(fmt.Sprintf("/app/agent-pools/%s/delete", agentPool))
}
