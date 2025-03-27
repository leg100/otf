// Code generated by "go generate"; DO NOT EDIT.

package paths

import (
	"fmt"

	"github.com/leg100/otf/internal/resource"
)

func AgentTokens(agentPool resource.ID) string {
	return fmt.Sprintf("/app/agent-pools/%s/agent-tokens", agentPool)
}

func CreateAgentToken(agentPool resource.ID) string {
	return fmt.Sprintf("/app/agent-pools/%s/agent-tokens/create", agentPool)
}

func NewAgentToken(agentPool resource.ID) string {
	return fmt.Sprintf("/app/agent-pools/%s/agent-tokens/new", agentPool)
}

func AgentToken(agentToken resource.ID) string {
	return fmt.Sprintf("/app/agent-tokens/%s", agentToken)
}

func EditAgentToken(agentToken resource.ID) string {
	return fmt.Sprintf("/app/agent-tokens/%s/edit", agentToken)
}

func UpdateAgentToken(agentToken resource.ID) string {
	return fmt.Sprintf("/app/agent-tokens/%s/update", agentToken)
}

func DeleteAgentToken(agentToken resource.ID) string {
	return fmt.Sprintf("/app/agent-tokens/%s/delete", agentToken)
}
