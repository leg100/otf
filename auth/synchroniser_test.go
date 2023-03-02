package auth

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/leg100/otf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserSyncOrganizations(t *testing.T) {
	ctx := context.Background()
	svc := &fakeSynchroniserService{}

	user := otf.NewUser(uuid.NewString(), otf.WithOrganizations("org-1", "org-2"))

	s := &synchroniser{
		service: svc,
	}
	want := []string{"org-2", "org-3"}
	err := s.syncOrganizations(ctx, user, want)
	require.NoError(t, err)

	// expect membership to have been added to org3
	if assert.Equal(t, 1, len(svc.addedOrgs)) {
		assert.Equal(t, "org-3", svc.addedOrgs[0])
	}
	// expect membership to have been removed from org1
	if assert.Equal(t, 1, len(svc.removedOrgs)) {
		assert.Equal(t, "org-1", svc.removedOrgs[0])
	}
}

func TestUserSyncTeams(t *testing.T) {
	ctx := context.Background()
	svc := &fakeSynchroniserService{}

	team1 := NewTestTeam(t, "org-1")
	team2 := NewTestTeam(t, "org-1")
	team3 := NewTestTeam(t, "org-1")

	user := otf.NewUser(uuid.NewString(), otf.WithTeams(team1, team2))

	s := &synchroniser{
		service: svc,
	}
	want := []*otf.Team{team2, team3}
	err := s.syncTeams(ctx, user, want)
	require.NoError(t, err)

	// expect membership to have been added to team3
	if assert.Equal(t, 1, len(svc.addedTeams)) {
		assert.Equal(t, team3.ID, svc.addedTeams[0])
	}
	// expect membership to have been removed from team1
	if assert.Equal(t, 1, len(svc.removedTeams)) {
		assert.Equal(t, team1.ID, svc.removedTeams[0])
	}
}

type fakeSynchroniserService struct {
	addedOrgs, removedOrgs, addedTeams, removedTeams []string

	service
}

func (f *fakeSynchroniserService) AddOrganizationMembership(ctx context.Context, userID, orgID string) error {
	f.addedOrgs = append(f.addedOrgs, orgID)
	return nil
}

func (f *fakeSynchroniserService) RemoveOrganizationMembership(ctx context.Context, userID, orgID string) error {
	f.removedOrgs = append(f.removedOrgs, orgID)
	return nil
}

func (f *fakeSynchroniserService) AddTeamMembership(ctx context.Context, userID, orgID string) error {
	f.addedTeams = append(f.addedTeams, orgID)
	return nil
}

func (f *fakeSynchroniserService) RemoveTeamMembership(ctx context.Context, userID, orgID string) error {
	f.removedTeams = append(f.removedTeams, orgID)
	return nil
}
