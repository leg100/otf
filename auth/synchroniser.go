package auth

import (
	"context"
	"errors"

	"github.com/go-logr/logr"
	"github.com/leg100/otf"
	"github.com/leg100/otf/cloud"
	"github.com/leg100/otf/organization"
)

// synchroniser synchronises a user account from the cloud to an a user account
// in OTF:
// * user account is created if it doesn't already exist
// * organizations are created if they don't already exist
// * a 'personal' organization is created matching username if it doesn't
// already exist
// * teams are created if they don't already exist
// * organization memberships are added if they don't already exist
// * organization memberships are removed if they exist in OTF but not on the cloud
// * team memberships are added if they don't already exist
// * team memberships are removed if they exist in OTF but not on the cloud
type synchroniser struct {
	logr.Logger
	OrganizationService

	AuthService
}

func (s *synchroniser) sync(ctx context.Context, from cloud.User) (*User, error) {
	// ensure user exists
	user, err := s.GetUser(ctx, UserSpec{Username: otf.String(from.Name)})
	if err == otf.ErrResourceNotFound {
		user, err = s.CreateUser(ctx, from.Name)
		if err != nil {
			return nil, err
		}
	}

	// Create organization for each cloud organization
	var organizations []string
	for _, want := range from.Organizations {
		got, err := s.GetOrganization(ctx, want)
		if err == otf.ErrResourceNotFound {
			got, err = s.CreateOrganization(ctx, organization.OrganizationCreateOptions{
				Name: otf.String(want),
			})
			if err != nil {
				return nil, err
			}
		} else if err != nil {
			return nil, err
		}
		organizations = append(organizations, got.Name)
	}

	// A user also gets their own personal organization matching their username
	personal, err := s.GetOrganization(ctx, user.Username)
	if err == otf.ErrResourceNotFound {
		personal, err = s.CreateOrganization(ctx, organization.OrganizationCreateOptions{
			Name: otf.String(user.Username),
		})
		if err != nil {
			return nil, err
		}
	}
	organizations = append(organizations, personal.Name)

	// Create team for each cloud team
	var teams []*Team
	for _, want := range from.Teams {
		got, err := s.GetTeam(ctx, want.Organization, want.Name)
		if err == otf.ErrResourceNotFound {
			got, err = s.CreateTeam(ctx, NewTeamOptions{
				Name:         want.Name,
				Organization: want.Organization,
			})
			if err != nil {
				return nil, err
			}
		} else if err != nil {
			return nil, err
		}
		teams = append(teams, got)
	}

	// And make them an owner of their personal org
	owners, err := s.GetTeam(ctx, personal.Name, "owners")
	if err == otf.ErrResourceNotFound {
		owners, err = s.CreateTeam(ctx, NewTeamOptions{
			Name:         "owners",
			Organization: personal.Name,
		})
		if err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}
	teams = append(teams, owners)

	// Add/remove memberships
	if err = s.syncOrganizations(ctx, user, organizations); err != nil {
		return nil, err
	}
	if err = s.syncTeams(ctx, user, teams); err != nil {
		return nil, err
	}

	return user, nil
}

// syncOrganizations updates a user's organization memberships to match those in wanted
func (s *synchroniser) syncOrganizations(ctx context.Context, u *User, wanted []string) error {
	// Add org memberships
	for _, want := range wanted {
		if !otf.Contains(u.Organizations, want) {
			if err := s.AddOrganizationMembership(ctx, u.Username, want); err != nil {
				if errors.Is(err, otf.ErrResourceAlreadyExists) {
					// ignore conflicts - sometimes the caller may provide
					// duplicate orgs
					continue
				} else {
					return err
				}
			}
		}
	}

	// Remove org memberships
	for _, got := range u.Organizations {
		if !otf.Contains(wanted, got) {
			if err := s.RemoveOrganizationMembership(ctx, u.Username, got); err != nil {
				return err
			}
		}
	}

	return nil
}

// syncTeams updates a user's team memberships to match those in wanted.
func (s *synchroniser) syncTeams(ctx context.Context, u *User, wanted []*Team) error {
	// Add team memberships
	for _, want := range wanted {
		if !u.IsTeamMember(want.ID) {
			if err := s.AddTeamMembership(ctx, u.Username, want.ID); err != nil {
				if errors.Is(err, otf.ErrResourceAlreadyExists) {
					// ignore conflicts - sometimes the caller may provide
					// duplicate teams
					continue
				} else {
					return err
				}
			}
		}
	}

	// Remove team memberships
	for _, team := range u.Teams {
		if !inTeamList(wanted, team.ID) {
			if err := s.RemoveTeamMembership(ctx, u.Username, team.ID); err != nil {
				return err
			}
		}
	}

	return nil
}

func inTeamList(teams []*Team, teamID string) bool {
	for _, team := range teams {
		if team.ID == teamID {
			return true
		}
	}
	return false
}
