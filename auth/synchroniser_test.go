package auth

import (
	"context"
	"testing"

	"github.com/leg100/otf/organization"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserSyncMemberships(t *testing.T) {
	ctx := context.Background()

	org1 := organization.NewTestOrganization(t)
	org2 := organization.NewTestOrganization(t)
	org3 := organization.NewTestOrganization(t)

	team1 := newTeam(createTeamOptions{"team-1", org1.Name()})
	team2 := newTeam(createTeamOptions{"team-2", org2.Name()})
	team3 := newTeam(createTeamOptions{"team-2", org3.Name()})

	user := NewTestUser(t,
		WithOrganizations(org1.Name(), org2.Name()),
		WithTeams(team1, team2))

	wantOrgMemberships := []string{org2.Name(), org3.Name()}
	wantTeamMemberships := []*Team{team2, team3}

	store := &fakeUserApp{}
	err := user.SyncMemberships(ctx, store, wantOrgMemberships, wantTeamMemberships)
	require.NoError(t, err)

	assert.Equal(t, wantOrgMemberships, user.organizations)
	assert.Equal(t, wantTeamMemberships, user.teams)

	// expect membership to have been added to org3
	if assert.Equal(t, 1, len(store.addedOrgs)) {
		assert.Equal(t, org3.Name(), store.addedOrgs[0])
	}
	// expect membership to have been removed from org1
	if assert.Equal(t, 1, len(store.removedOrgs)) {
		assert.Equal(t, org1.Name(), store.removedOrgs[0])
	}
	// expect membership to have been added to team3
	if assert.Equal(t, 1, len(store.addedTeams)) {
		assert.Equal(t, team3.ID(), store.addedTeams[0])
	}
	// expect membership to have been removed from team1
	if assert.Equal(t, 1, len(store.removedTeams)) {
		assert.Equal(t, team1.ID(), store.removedTeams[0])
	}
}
