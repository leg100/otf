package sql

import (
	"context"
	"testing"

	"github.com/leg100/otf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTeam_Create(t *testing.T) {
	ctx := context.Background()
	db := newTestDB(t)
	org := createTestOrganization(t, db)
	team := otf.NewTeam("team-awesome", org)

	defer db.DeleteTeam(ctx, otf.TeamSpec{ID: otf.String(team.ID())})

	err := db.CreateTeam(ctx, team)
	require.NoError(t, err)
}

func TestTeam_Get(t *testing.T) {
	db := newTestDB(t)

	org := createTestOrganization(t, db)
	team := createTestTeam(t, db, org)

	tests := []struct {
		name string
		spec otf.TeamSpec
	}{
		{
			name: "id",
			spec: otf.TeamSpec{ID: otf.String(team.ID())},
		},
		{
			name: "name and organization name",
			spec: otf.TeamSpec{Name: otf.String(team.Name()), OrganizationName: otf.String(team.OrganizationName())},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := db.GetTeam(context.Background(), tt.spec)
			require.NoError(t, err)

			assert.Equal(t, team, got)
		})
	}
}

func TestTeam_List(t *testing.T) {
	db := newTestDB(t)
	org := createTestOrganization(t, db)
	team1 := createTestTeam(t, db, org)
	team2 := createTestTeam(t, db, org)
	team3 := createTestTeam(t, db, org)

	got, err := db.ListTeams(context.Background(), org.Name())
	require.NoError(t, err)

	assert.Contains(t, got, team1)
	assert.Contains(t, got, team2)
	assert.Contains(t, got, team3)
}
