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
// in otf:
// * user account is created if it doesn't already exist
// * organizations are created if they don't already exist
// * a 'personal' organization is created matching username if it doesn't
// already exist, and they are made an owner of this org
// * teams are created if they don't already exist
// * team memberships are added if they don't already exist
// * team memberships are removed if they exist in otf but not on the cloud
type synchroniser struct {
	logr.Logger
	OrganizationService

	AuthService
}

func (s *synchroniser) sync(ctx context.Context, from cloud.User) (*User, error) {
	// Create user account
	user, err := s.GetUser(ctx, UserSpec{Username: otf.String(from.Name)})
	if err == otf.ErrResourceNotFound {
		user, err = s.CreateUser(ctx, from.Name)
		if err != nil {
			return nil, err
		}
	}

	// fn for idempotently creating an org
	getOrCreateOrg := func(name string) error {
		_, err := s.GetOrganization(ctx, name)
		if err == otf.ErrResourceNotFound {
			_, err = s.CreateOrganization(ctx, organization.OrganizationCreateOptions{
				Name: otf.String(name),
			})
			return err
		}
		return nil
	}

	// fn for idempotently creating a team
	getOrCreateTeam := func(ct cloud.Team) (*Team, error) {
		team, err := s.GetTeam(ctx, ct.Organization, ct.Name)
		if err == otf.ErrResourceNotFound {
			return s.CreateTeam(ctx, NewTeamOptions{
				Name:         ct.Name,
				Organization: ct.Organization,
			})
		}
		return team, err
	}

	// Create org and team for each cloud team
	var teams []*Team
	for _, ct := range from.Teams {
		if err := getOrCreateOrg(ct.Organization); err != nil {
			return nil, err
		}
		team, err := getOrCreateTeam(ct)
		if err != nil {
			return nil, err
		}
		teams = append(teams, team)
	}

	// Create org matching their username and make them an owner.
	if err := getOrCreateOrg(user.Username); err != nil {
		return nil, err
	}
	team, err := getOrCreateTeam(cloud.Team{Name: "owners", Organization: user.Username})
	if err != nil {
		return nil, err
	}
	teams = append(teams, team)

	// Add/remove team memberships
	if err = s.syncTeams(ctx, user, teams); err != nil {
		return nil, err
	}

	return user, nil
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
