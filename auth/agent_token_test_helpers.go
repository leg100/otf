package auth

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func NewTestAgentToken(t *testing.T, org string) *AgentToken {
	token, _, err := NewAgentToken(NewAgentTokenOptions{
		CreateAgentTokenOptions: CreateAgentTokenOptions{
			Organization: org,
			Description:  "lorem ipsum...",
		},
	})
	require.NoError(t, err)
	return token
}
