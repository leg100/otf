package otf

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserSyncMemberships(t *testing.T) {
	ctx := context.Background()

	org1 := NewTestOrganization(t)
	org2 := NewTestOrganization(t)
	org3 := NewTestOrganization(t)

	team1 := NewTeam("team-1", org1)
	team2 := NewTeam("team-2", org2)
	team3 := NewTeam("team-2", org3)

	user := NewUser("test-user",
		WithOrganizationMemberships(org1.Name(), org2.Name()),
		WithTeamMemberships(team1, team2))

	wantOrgMemberships := []string{org2.Name(), org3.Name()}
	wantTeamMemberships := []*Team{team2, team3}

	store := &fakeUserStore{}
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

type fakeUserStore struct {
	// IDs of orgs and teams added and removed
	addedOrgs, removedOrgs, addedTeams, removedTeams []string
	UserStore
}

func (f *fakeUserStore) AddOrganizationMembership(ctx context.Context, userID, orgID string) error {
	f.addedOrgs = append(f.addedOrgs, orgID)
	return nil
}

func (f *fakeUserStore) RemoveOrganizationMembership(ctx context.Context, userID, orgID string) error {
	f.removedOrgs = append(f.removedOrgs, orgID)
	return nil
}

func (f *fakeUserStore) AddTeamMembership(ctx context.Context, userID, orgID string) error {
	f.addedTeams = append(f.addedTeams, orgID)
	return nil
}

func (f *fakeUserStore) RemoveTeamMembership(ctx context.Context, userID, orgID string) error {
	f.removedTeams = append(f.removedTeams, orgID)
	return nil
}
