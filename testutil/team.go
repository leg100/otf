package testutil

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/leg100/otf"
	"github.com/leg100/otf/auth"
	"github.com/leg100/otf/organization"
	"github.com/stretchr/testify/require"
)

func CreateTeam(t *testing.T, db otf.DB, org *organization.Organization) *auth.Team {
	ctx := context.Background()
	svc := NewAuthService(t, db)

	team, err := svc.CreateTeam(ctx, otf.CreateTeamOptions{
		Name:         uuid.NewString(),
		Organization: org.Name(),
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		svc.DeleteTeam(ctx, team.ID())
	})
	return team
}
