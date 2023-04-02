package authenticator

import (
	"context"
	"errors"

	"github.com/go-logr/logr"
	"github.com/leg100/otf"
	"github.com/leg100/otf/auth"
	"github.com/leg100/otf/cloud"
	"github.com/leg100/otf/organization"
	"github.com/leg100/otf/orgcreator"
)

// Synchroniser synchronises a user account from the cloud to a user account in
// otf:
// * user account created if it doesn't already exist
// * for each cloud team:
//   - if owners then create corresponding organization if it doesn't already
//     exist as well as owners team and make them an owner.
//   - if a non-owners team then the organization is expected to already exist
//     and if so then team is created if it doesn't already exist, and they are
//     made a member.
//
// * for each otf team the user is a member of but for which there is no
// corresponding cloud team:
//   - if owners then as long as there is at least one other owner then they are
//     removed from the owners team in otf; if they are the last owner then the
//     membership is not removed (every otf organization must have at least one
//     owner).
//   - if non-owners then they are removed from the team.
//
// * a personal organization is created matching username if it doesn't already
// exist, along with an owners team of which they are made a member.
type synchroniser struct {
	logr.Logger

	organization.OrganizationService
	orgcreator.OrganizationCreatorService
	auth.AuthService
}

func (s *synchroniser) Sync(ctx context.Context, user cloud.User) error {
	// To make comparisons easier, convert the user from a cloud.User struct to an
	// auth.User struct
	want := &auth.User{Username: user.Name}
	for _, t := range user.Teams {
		want.Teams = append(want.Teams, &auth.Team{
			Name:         t.Name,
			Organization: t.Organization,
		})
	}
	// The wanted user should also be an owner of an organization matching its
	// username
	want.Teams = append(want.Teams, &auth.Team{
		Name:         "owners",
		Organization: want.Username,
	})

	// Create user account
	got, err := s.GetUser(ctx, auth.UserSpec{Username: otf.String(user.Name)})
	if err == otf.ErrResourceNotFound {
		got, err = s.CreateUser(ctx, user.Name)
		if err != nil {
			return err
		}
	}

	// user context for performing actions as user.
	userCtx := otf.AddSubjectToContext(ctx, got)

	// work out which teams to add and remove
	add, remove := teamdiff(want.Teams, got.Teams)

	// add orgs, teams, memberships
	for _, t := range add {
		_, err := s.GetOrganization(ctx, t.Organization)
		if err == otf.ErrResourceNotFound {
			if t.Name != "owners" {
				// skip: non-owner cannot create an org, and therefore a team
				// cannot be created either
				continue
			}
			_, err = s.CreateOrganization(userCtx, orgcreator.OrganizationCreateOptions{
				Name: otf.String(t.Organization),
			})
			if err != nil {
				return err
			}
			// creating an org automatically creates an owners team and
			// automatically makes the user an owner, therefore team and
			// membership creation is skipped.
			continue
		} else if err != nil {
			return err
		}
		team, err := s.GetTeam(ctx, t.Organization, t.Name)
		if err == otf.ErrResourceNotFound {
			if team.Name == "owners" {
				// this should not happen: an organization always has an owners
				// team
				return errors.New("owners team not found")
			}
			_, err := s.CreateTeam(ctx, auth.CreateTeamOptions{
				Name:         t.Name,
				Organization: t.Organization,
			})
			if err != nil {
				return err
			}
		} else if err != nil {
			return err
		}
		err = s.AddTeamMembership(ctx, auth.TeamMembershipOptions{
			TeamID:   team.ID,
			Username: got.Username,
		})
		if err != nil {
			return err
		}
	}

	// remove team memberships (orgs and teams are not removed during a user
	// sync).
	for _, t := range remove {
		err := s.RemoveTeamMembership(ctx, auth.TeamMembershipOptions{
			TeamID:   t.ID,
			Username: got.Username,
		})
		if err == auth.ErrCannotDeleteOnlyOwner {
			// user is last owner of org, in which case they remain an owner.
			// The alternative design decision would be to delete the
			// organization, but that is pretty drastic. Neither option is good,
			// because the former could cause a security lapse: an admin has
			// removed the user from the upstream owners team yet they remain an
			// owner in otf...
			continue
		} else if err != nil {
			return err
		}
	}

	return nil
}

// teamdiff works out which teams to add and which to remove in order to make
// the current user's teams match the teams of the wanted user.
func teamdiff(want, got []*auth.Team) (add, remove []*auth.Team) {
	// inTeam checks whether teams contains the wanted team.
	inTeam := func(teams []*auth.Team, want *auth.Team) bool {
		for _, got := range teams {
			if got.Organization == want.Organization && got.Name == want.Name {
				return true
			}
		}
		return false
	}

	// fn for determining team memberships to add
	addTeams := func(want, got []*auth.Team) (add []*auth.Team) {
		for _, wt := range want {
			if inTeam(got, wt) {
				continue
			}
			add = append(add, wt)
		}
		return
	}

	// fn for determining team memberships to remove
	removeTeams := func(want, got []*auth.Team) (remove []*auth.Team) {
		for _, gt := range got {
			if inTeam(want, gt) {
				continue
			}
			remove = append(remove, gt)
		}
		return
	}

	return addTeams(want, got), removeTeams(want, got)
}
