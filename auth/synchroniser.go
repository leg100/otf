package auth

import (
	"context"
	"errors"

	"github.com/go-logr/logr"
	"github.com/leg100/otf"
)

// synchroniser updates a user's memberships of organizations and teams in the
// db
type synchroniser struct {
	logr.Logger
	db
}

// syncOrganizations updates a user's organization memberships in the db to
// match those in wanted
func (s *synchroniser) syncOrganizations(ctx context.Context, u *User, wanted []string) error {
	// Add org memberships
	for _, want := range wanted {
		if !otf.Contains(u.organizations, want) {
			if err := s.addOrganizationMembership(ctx, u.ID(), want); err != nil {
				if errors.Is(err, otf.ErrResourceAlreadyExists) {
					// ignore conflicts - sometimes the caller may provide
					// duplicate orgs
					continue
				} else {
					return err
				}
			}
			s.V(0).Info("added organization membership", "user", u, "org", want)
		}
	}

	// Remove org memberships
	for _, got := range u.organizations {
		if !otf.Contains(wanted, got) {
			if err := s.removeOrganizationMembership(ctx, u.ID(), got); err != nil {
				return err
			}
			s.V(0).Info("removed organization membership", "user", u, "org", got)
		}
	}

	return nil
}

// syncTeams updates a user's team memberships in the db to match those in
// wanted.
func (s *synchroniser) syncTeams(ctx context.Context, u *User, wanted []*Team) error {
	// Add team memberships
	for _, want := range wanted {
		if !u.IsTeamMember(want.id) {
			if err := s.addTeamMembership(ctx, u.ID(), want.ID()); err != nil {
				if errors.Is(err, otf.ErrResourceAlreadyExists) {
					// ignore conflicts - sometimes the caller may provide
					// duplicate teams
					continue
				} else {
					return err
				}
			}
			s.V(0).Info("added team membership", "user", u, "org", want.organization, "team", want.name)
		}
	}

	// Remove team memberships
	for _, team := range u.teams {
		if !inTeamList(wanted, team.id) {
			if err := s.removeTeamMembership(ctx, u.ID(), team.ID()); err != nil {
				return err
			}
			s.V(0).Info("removed team membership", "user", u, "org", team.organization, "team", team.name)
		}
	}

	return nil
}

func inTeamList(teams []*Team, name string) bool {
	for _, team := range teams {
		if team.name == name {
			return true
		}
	}
	return false
}
