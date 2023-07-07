package auth

import (
	"context"
	"testing"

	"github.com/go-logr/logr"
	"github.com/google/uuid"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/sql"
	"github.com/stretchr/testify/require"
)

func NewTestTeam(t *testing.T, organization string) *Team {
	team, err := newTeam(organization, CreateTeamOptions{
		Name: internal.String(uuid.NewString()),
	})
	require.NoError(t, err)
	return team
}

func CreateTestTeam(t *testing.T, db *sql.DB, organization *organization.Organization) *Team {
	userDB := newDB(db, logr.Discard())
	return createTestTeam(t, userDB, organization.Name)
}

func createTestTeam(t *testing.T, db *pgdb, organization string) *Team {
	ctx := context.Background()

	team := NewTestTeam(t, organization)
	err := db.createTeam(ctx, team)
	require.NoError(t, err)

	t.Cleanup(func() {
		db.deleteTeam(ctx, team.ID)
	})
	return team
}

func newTestOwners(t *testing.T, organization string) *Team {
	team, err := newTeam(organization, CreateTeamOptions{
		Name: internal.String("owners"),
	})
	require.NoError(t, err)
	return team
}
