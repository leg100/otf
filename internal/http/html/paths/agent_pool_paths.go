// Code generated by "go generate"; DO NOT EDIT.

package paths

import (
	"fmt"

	"github.com/leg100/otf/internal/resource"
)

func AgentPools(organization resource.ID) string {
	return fmt.Sprintf("/app/organizations/%s/agent-pools", organization)
}

func CreateAgentPool(organization resource.ID) string {
	return fmt.Sprintf("/app/organizations/%s/agent-pools/create", organization)
}

func NewAgentPool(organization resource.ID) string {
	return fmt.Sprintf("/app/organizations/%s/agent-pools/new", organization)
}

func AgentPool(agentPool resource.ID) string {
	return fmt.Sprintf("/app/agent-pools/%s", agentPool)
}

func EditAgentPool(agentPool resource.ID) string {
	return fmt.Sprintf("/app/agent-pools/%s/edit", agentPool)
}

func UpdateAgentPool(agentPool resource.ID) string {
	return fmt.Sprintf("/app/agent-pools/%s/update", agentPool)
}

func DeleteAgentPool(agentPool resource.ID) string {
	return fmt.Sprintf("/app/agent-pools/%s/delete", agentPool)
}
