package workspace

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/leg100/otf"
	"github.com/stretchr/testify/require"
)

func CreateTestWorkspace(t *testing.T, db otf.DB, organization string) Workspace {
	ctx := context.Background()
	wsDB := newdb(db)
	ws, err := NewWorkspace(CreateWorkspaceOptions{
		Name:         otf.String(uuid.NewString()),
		Organization: &organization,
	})
	err = wsDB.CreateWorkspace(ctx, ws)
	require.NoError(t, err)

	t.Cleanup(func() {
		wsDB.DeleteWorkspace(ctx, ws.ID)
	})
	return Workspace{
		ID:   ws.ID,
		Name: ws.Name,
	}
}
