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
		key: newTestJWK(t, "something_secret"),
	})
	require.NoError(t, err)
	return token
}
