package auth

import (
	"fmt"
	"time"

	"github.com/leg100/otf"
	"github.com/leg100/otf/rbac"
)

var SiteAdmin = User{id: otf.SiteAdminID, username: "site-admin"}

// User represents an otf user account.
type User struct {
	id            string // ID uniquely identifies users
	createdAt     time.Time
	updatedAt     time.Time
	username      string   // username is globally unique
	organizations []string // user belongs to many organizations
	teams         []*Team  // user belongs to many teams
}

func newUser(username string) *User {
	return &User{
		id:        otf.NewID("user"),
		username:  username,
		createdAt: otf.CurrentTimestamp(),
		updatedAt: otf.CurrentTimestamp(),
	}
}

func (u *User) ID() string              { return u.id }
func (u *User) Username() string        { return u.username }
func (u *User) CreatedAt() time.Time    { return u.createdAt }
func (u *User) UpdatedAt() time.Time    { return u.updatedAt }
func (u *User) String() string          { return u.username }
func (u *User) Organizations() []string { return u.organizations }

// TeamsByOrganization return a user's teams filtered by organization name
func (u *User) TeamsByOrganization(organization string) []otf.Team {
	var orgTeams []otf.Team
	for _, t := range u.teams {
		if t.Organization() == organization {
			orgTeams = append(orgTeams, t)
		}
	}
	return orgTeams
}

// Team retrieves the named team in the given organization.
func (u *User) Team(name, organization string) (otf.Team, error) {
	for _, t := range u.teams {
		if t.Name() == name && t.Organization() == organization {
			return t, nil
		}
	}
	return nil, fmt.Errorf("no team found with the name: %s", name)
}

// IsTeamMember determines whether user is a member of the given team.
func (u *User) IsTeamMember(teamID string) bool {
	for _, t := range u.teams {
		if t.ID() == teamID {
			return true
		}
	}
	return false
}

func (u *User) IsUnprivilegedUser(organization string) bool {
	return !u.IsSiteAdmin() && !u.IsOwner(organization)
}

func (u *User) IsSiteAdmin() bool { return u.id == otf.SiteAdminID }

func (u *User) CanAccessSite(action rbac.Action) bool {
	// Only site admin can perform actions on the site
	return u.IsSiteAdmin()
}

func (u *User) CanAccessOrganization(action rbac.Action, name string) bool {
	// Site admin can perform any action on any organization
	if u.IsSiteAdmin() {
		return true
	}

	for _, team := range u.teams {
		if team.Organization() == name {
			if team.IsOwners() {
				// owner team members can perform all actions on organization
				return true
			}
			if rbac.OrganizationGuestRole.IsAllowed(action) {
				return true
			}
			if team.access.ManageWorkspaces {
				if rbac.WorkspaceManagerRole.IsAllowed(action) {
					return true
				}
			}
			if team.access.ManageVCS {
				if rbac.VCSManagerRole.IsAllowed(action) {
					return true
				}
			}
			if team.access.ManageRegistry {
				if rbac.VCSManagerRole.IsAllowed(action) {
					return true
				}
			}
		}
	}
	return false
}

func (u *User) CanAccessWorkspace(action rbac.Action, policy *otf.WorkspacePolicy) bool {
	// Site admin can access any workspace
	if u.IsSiteAdmin() {
		return true
	}
	// user must be a member of a team with perms
	for _, team := range u.teams {
		if team.Organization() == policy.Organization {
			if team.IsOwners() {
				// owner team members can perform all actions on all workspaces
				return true
			}
			if team.access.ManageWorkspaces {
				// workspace managers can perform all actions on all workspaces
				return true
			}
			for _, perm := range policy.Permissions {
				if team.id == perm.TeamID {
					return perm.Role.IsAllowed(action)
				}
			}
		}
	}
	return false
}

// IsOwner determines if user is an owner of an organization
func (u *User) IsOwner(organization string) bool {
	for _, team := range u.teams {
		if team.Organization() == organization {
			if team.IsOwners() {
				return true
			}
		}
	}
	return false
}

// CanLock always returns an error because nothing can replace a user lock
func (u *User) CanLock(requestor otf.Identity) error {
	return otf.ErrWorkspaceAlreadyLocked
}

// CanUnlock decides whether to permits requestor to unlock a user lock
func (u *User) CanUnlock(requestor otf.Identity, force bool) error {
	if force {
		// TODO: only grant admin user
		return nil
	}
	if user, ok := requestor.(*User); ok {
		if u.ID() == user.ID() {
			// only same user can unlock
			return nil
		}
		return otf.ErrWorkspaceLockedByDifferentUser
	}
	// any other entity cannot unlock
	return otf.ErrWorkspaceUnlockDenied
}

// UserListOptions are options for the ListUsers endpoint.
type UserListOptions struct {
	Organization *string
	TeamName     *string
}
