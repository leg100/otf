package auth

import (
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

func NewUser(username string, opts ...NewUserOption) *User {
	user := &User{
		id:        otf.NewID("user"),
		username:  username,
		createdAt: otf.CurrentTimestamp(),
		updatedAt: otf.CurrentTimestamp(),
	}
	for _, fn := range opts {
		fn(user)
	}
	return user
}

type NewUserOption func(*User)

func WithOrganizations(organizations ...string) NewUserOption {
	return func(user *User) {
		user.organizations = organizations
	}
}

func WithTeams(memberships ...*Team) NewUserOption {
	return func(user *User) {
		user.teams = memberships
	}
}

func (u *User) ID() string              { return u.id }
func (u *User) Username() string        { return u.username }
func (u *User) CreatedAt() time.Time    { return u.createdAt }
func (u *User) UpdatedAt() time.Time    { return u.updatedAt }
func (u *User) String() string          { return u.username }
func (u *User) Organizations() []string { return u.organizations }

// IsTeamMember determines whether user is a member of the given team.
func (u *User) IsTeamMember(teamID string) bool {
	for _, t := range u.teams {
		if t.ID() == teamID {
			return true
		}
	}
	return false
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
			if team.isOwners() {
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
			if team.isOwners() {
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
			if team.isOwners() {
				return true
			}
		}
	}
	return false
}

// UserListOptions are options for the ListUsers endpoint.
type UserListOptions struct {
	Organization *string
	TeamName     *string
}
