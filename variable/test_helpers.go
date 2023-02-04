package variable

import (
	"testing"

	"github.com/leg100/otf"
	"github.com/stretchr/testify/require"
)

func NewTestVariable(t *testing.T, ws *otf.Workspace, opts otf.CreateVariableOptions) *Variable {
	v, err := NewVariable(ws.ID(), opts)
	require.NoError(t, err)
	return v
}
