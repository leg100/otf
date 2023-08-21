package tokens

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAgentTokenNewCommand(t *testing.T) {
	cmd := newFakeCLI([]byte("secret-token")).agentTokenNewCommand()
	cmd.SetArgs([]string{"testing", "--organization", "automatize"})
	got := bytes.Buffer{}
	cmd.SetOut(&got)
	require.NoError(t, cmd.Execute())
	assert.Regexp(t, `Successfully created agent token: secret-token`, got.String())
}

type fakeCLIService struct {
	at []byte

	TokensService
}

func newFakeCLI(at []byte) *CLI {
	return &CLI{TokensService: &fakeCLIService{at: at}}
}

func (f *fakeCLIService) CreateAgentToken(ctx context.Context, opts CreateAgentTokenOptions) ([]byte, error) {
	return f.at, nil
}
