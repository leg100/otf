package agenttoken

import (
	"testing"

	"github.com/leg100/otf"
	"github.com/stretchr/testify/require"
)

func NewTestAgentToken(t *testing.T, org otf.Organization) *AgentToken {
	token, err := NewAgentToken(CreateAgentTokenOptions{
		Organization: org.Name(),
		Description:  "lorem ipsum...",
	})
	require.NoError(t, err)
	return token
}
