package otf

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func NewTestWorkspace(t *testing.T, org *Organization) *Workspace {
	ws, err := NewWorkspace(org, WorkspaceCreateOptions{
		Name: uuid.NewString(),
	})
	require.NoError(t, err)
	return ws
}
