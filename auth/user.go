package auth

import (
	"context"
	"errors"
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
	username      string     // username is globally unique
	organizations []string   // user belongs to many organizations
	teams         []otf.Team // user belongs to many teams
}

func (u *User) ID() string              { return u.id }
func (u *User) Username() string        { return u.username }
func (u *User) CreatedAt() time.Time    { return u.createdAt }
func (u *User) UpdatedAt() time.Time    { return u.updatedAt }
func (u *User) String() string          { return u.username }
func (u *User) Organizations() []string { return u.organizations }
func (u *User) Teams() []otf.Team       { return u.teams }

// ToJSONAPI assembles a JSON-API DTO.
func (u *User) ToJSONAPI() any {
	return &jsonapiUser{
		ID:       u.ID(),
		Username: u.Username(),
	}
}

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

func (u *User) IsSiteAdmin() bool { return u.id == SiteAdminID }

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

func (u *User) CanAccessWorkspace(action rbac.Action, policy *WorkspacePolicy) bool {
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

// SyncMemberships synchronises the user's organization and team memberships to
// match those given, adding and removing memberships in the persistence store accordingly.
func (u *User) SyncMemberships(ctx context.Context, store UserStore, orgs []string, teams []*Team) error {
	// Iterate through orgs and check if already a member; if not then
	// add membership to store
	for _, org := range orgs {
		if !otf.Contains(u.organizations, org) {
			if err := store.AddOrganizationMembership(ctx, u.ID(), org); err != nil {
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
	// Iterate thru receiver's orgs and check if in the given orgs; if not then
	// remove membership from store
	for _, org := range u.organizations {
		if !otf.Contains(orgs, org) {
			if err := store.RemoveOrganizationMembership(ctx, u.ID(), org); err != nil {
				return err
			}
		}
	}
	u.organizations = orgs

	// Iterate thru teams and check if already member; if not then
	// add membership to store
	for _, team := range teams {
		if !inTeamList(u.teams, team.ID()) {
			if err := store.AddTeamMembership(ctx, u.ID(), team.ID()); err != nil {
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
	// Iterate thru receiver's teams and check if in the given teams; if
	// not then remove membership from store
	for _, team := range u.teams {
		if !inTeamList(teams, team.ID()) {
			if err := store.RemoveTeamMembership(ctx, u.ID(), team.ID()); err != nil {
				return err
			}
		}
	}
	u.teams = teams

	return nil
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

// UserService provides methods to interact with user accounts and their
// sessions.
type UserService interface {
	// CreateUser creates a user with the given username.
	CreateUser(ctx context.Context, username string) (*User, error)
	// EnsureCreatedUser retrieves a user; if they don't exist they'll be
	// created.
	EnsureCreatedUser(ctx context.Context, username string) (*User, error)
	// Get retrieves a user according to the spec.
	GetUser(ctx context.Context, spec UserSpec) (*User, error)
	// SyncUserMemberships makes the user a member of the specified organizations
	// and teams and removes any existing memberships not specified.
	SyncUserMemberships(ctx context.Context, user *User, orgs []string, teams []*Team) (*User, error)
	// ListUsers lists users.
	ListUsers(ctx context.Context, opts UserListOptions) ([]*User, error)
}

// UserStore is a persistence store for user accounts.
type UserStore interface {
	CreateUser(ctx context.Context, user *User) error
	GetUser(ctx context.Context, spec UserSpec) (*User, error)
	// ListUsers lists users.
	ListUsers(ctx context.Context, opts UserListOptions) ([]*User, error)
	DeleteUser(ctx context.Context, spec UserSpec) error
	// AddOrganizationMembership adds a user as a member of an organization
	AddOrganizationMembership(ctx context.Context, id, orgID string) error
	// RemoveOrganizationMembership removes a user as a member of an
	// organization
	RemoveOrganizationMembership(ctx context.Context, id, orgID string) error
	// AddTeamMembership adds a user as a member of a team
	AddTeamMembership(ctx context.Context, id, teamID string) error
	// RemoveTeamMembership removes a user as a member of an
	// team
	RemoveTeamMembership(ctx context.Context, id, teamID string) error
}

// UserListOptions are options for the ListUsers endpoint.
type UserListOptions struct {
	Organization *string
	TeamName     *string
}

type UserSpec struct {
	UserID              *string
	Username            *string
	SessionToken        *string
	AuthenticationToken *string
}

// KeyValue returns the user spec in key-value form. Useful for logging
// purposes.
func (spec *UserSpec) KeyValue() []interface{} {
	if spec.Username != nil {
		return []interface{}{"username", *spec.Username}
	}
	if spec.SessionToken != nil {
		return []interface{}{"token", *spec.SessionToken}
	}
	if spec.AuthenticationToken != nil {
		return []interface{}{"authentication_token", "*****"}
	}

	return []interface{}{"invalid user spec", ""}
}

func NewUser(username string, opts ...NewUserOption) *User {
	user := User{
		id:        NewID("user"),
		username:  username,
		createdAt: CurrentTimestamp(),
		updatedAt: CurrentTimestamp(),
	}
	for _, o := range opts {
		o(&user)
	}
	return &user
}

type NewUserOption func(*User)

func WithOrganizationMemberships(organizations ...string) NewUserOption {
	return func(user *User) {
		user.organizations = organizations
	}
}

func WithTeamMemberships(memberships ...*Team) NewUserOption {
	return func(user *User) {
		user.teams = memberships
	}
}

func inTeamList(teams []*Team, teamID string) bool {
	for _, team := range teams {
		if team.ID() == teamID {
			return true
		}
	}
	return false
}
