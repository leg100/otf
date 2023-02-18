package testutil

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/leg100/otf"
	"github.com/leg100/otf/auth"
	"github.com/stretchr/testify/require"
)

func CreateTeam(t *testing.T, db otf.DB, opts ...auth.NewUserOption) *auth.User {
	ctx := context.Background()
	svc := NewAuthService(t, db)

	team, err := svc.CreateTeam(ctx, uuid.NewString())
	require.NoError(t, err)

	t.Cleanup(func() {
		svc.DeleteTeam(ctx, team.Username())
	})
	return team
}
