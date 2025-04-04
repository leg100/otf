// Code generated by "go generate"; DO NOT EDIT.

package paths

import "fmt"

func AgentTokens(agentPool fmt.Stringer) string {
	return fmt.Sprintf("/app/agent-pools/%s/agent-tokens", agentPool)
}

func CreateAgentToken(agentPool fmt.Stringer) string {
	return fmt.Sprintf("/app/agent-pools/%s/agent-tokens/create", agentPool)
}

func NewAgentToken(agentPool fmt.Stringer) string {
	return fmt.Sprintf("/app/agent-pools/%s/agent-tokens/new", agentPool)
}

func AgentToken(agentToken fmt.Stringer) string {
	return fmt.Sprintf("/app/agent-tokens/%s", agentToken)
}

func EditAgentToken(agentToken fmt.Stringer) string {
	return fmt.Sprintf("/app/agent-tokens/%s/edit", agentToken)
}

func UpdateAgentToken(agentToken fmt.Stringer) string {
	return fmt.Sprintf("/app/agent-tokens/%s/update", agentToken)
}

func DeleteAgentToken(agentToken fmt.Stringer) string {
	return fmt.Sprintf("/app/agent-tokens/%s/delete", agentToken)
}
