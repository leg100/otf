package auth

import "context"

type fakeUserApp struct {
	// IDs of orgs and teams added and removed
	addedOrgs, removedOrgs, addedTeams, removedTeams []string

	userService
}

func (f *fakeUserApp) AddOrganizationMembership(ctx context.Context, userID, orgID string) error {
	f.addedOrgs = append(f.addedOrgs, orgID)
	return nil
}

func (f *fakeUserApp) RemoveOrganizationMembership(ctx context.Context, userID, orgID string) error {
	f.removedOrgs = append(f.removedOrgs, orgID)
	return nil
}

func (f *fakeUserApp) AddTeamMembership(ctx context.Context, userID, orgID string) error {
	f.addedTeams = append(f.addedTeams, orgID)
	return nil
}

func (f *fakeUserApp) RemoveTeamMembership(ctx context.Context, userID, orgID string) error {
	f.removedTeams = append(f.removedTeams, orgID)
	return nil
}
