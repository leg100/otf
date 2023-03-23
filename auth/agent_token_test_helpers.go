package auth

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func NewTestAgentToken(t *testing.T, org string) *AgentToken {
	token, err := NewAgentToken(CreateAgentTokenOptions{
		Organization: org,
		Description:  "lorem ipsum...",
	})
	require.NoError(t, err)
	return token
}
