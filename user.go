package otf

import (
	"context"
	"time"
)

const (
	DefaultUserID   = "user-123"
	DefaultUsername = "otf"
)

var (
	SiteAdminID = "user-site-admin"
	SiteAdmin   = User{id: SiteAdminID, username: "site-admin"}
)

// User represents an oTF user account.
type User struct {
	// ID uniquely identifies users
	id        string
	createdAt time.Time
	updatedAt time.Time
	username  string
	// A user has many sessions
	Sessions []*Session
	// A user has many tokens
	Tokens []*Token
	// A user belongs to many organizations
	Organizations []*Organization
	// A user belongs to many teams
	Teams []*Team
}

// AttachNewSession creates and attaches a new session to the user.
func (u *User) AttachNewSession(data *SessionData) (*Session, error) {
	session, err := NewSession(u.ID(), data)
	if err != nil {
		return nil, err
	}
	u.Sessions = append(u.Sessions, session)
	return session, nil
}

func (u *User) ID() string           { return u.id }
func (u *User) Username() string     { return u.username }
func (u *User) CreatedAt() time.Time { return u.createdAt }
func (u *User) UpdatedAt() time.Time { return u.updatedAt }
func (u *User) String() string       { return u.username }

func (u *User) IsSiteAdmin() bool { return u.id == SiteAdminID }

func (u *User) CanAccessSite(action Action) bool {
	// Only site admin can perform actions on the site
	return u.IsSiteAdmin()
}

func (u *User) CanAccessOrganization(action Action, name string) bool {
	// Site admin can perform any action on any organization
	if u.IsSiteAdmin() {
		return true
	}

	for _, team := range u.Teams {
		if team.OrganizationName() == name {
			if team.IsOwners() {
				// owner team members can perform all actions on organization
				return true
			}
			if team.access.ManageWorkspaces {
				// check if workspace manager role allows action
				return workspaceManagerPermissions[action]
			}
			// TODO: as we add more organization-level features, such as a
			// registry, policies, etc, we'll introduce further manager roles
			// and check if roles allow action here.

			switch action {
			case GetOrganizationAction:
				// members can retrieve info about their organization
				return true
			}
		}
	}
	return false
}

func (u *User) CanAccessWorkspace(action Action, policy *WorkspacePolicy) bool {
	// Site admin can access any workspace
	if u.IsSiteAdmin() {
		return true
	}
	// user must be a member of a team with perms
	for _, team := range u.Teams {
		if team.OrganizationName() == policy.OrganizationName {
			if team.IsOwners() {
				// owner team members can perform all actions on all workspaces
				return true
			}
			if team.access.ManageWorkspaces {
				// workspace managers can perform all actions on all workspaces
				return true
			}
			for _, perm := range policy.Permissions {
				if team.id == perm.Team.id {
					return IsAllowed(action, perm.Permission)
				}
			}
		}
	}
	return false
}

func (u *User) ActiveSession() *Session {
	for _, s := range u.Sessions {
		if s.active {
			return s
		}
	}
	return nil
}

// SyncMemberships synchronises the user's organization and team memberships to
// match those given, adding and removing memberships in the persistence store accordingly.
func (u *User) SyncMemberships(ctx context.Context, store UserStore, orgs []*Organization, teams []*Team) error {
	// Iterate thru orgs and check if already member; if not then
	// add membership to store
	for _, org := range orgs {
		if !inOrganizationList(u.Organizations, org.ID()) {
			if err := store.AddOrganizationMembership(ctx, u.ID(), org.ID()); err != nil {
				return err
			}
		}
	}
	// Iterate thru receiver's orgs and check if in the given orgs; if not then
	// remove membership from store
	for _, org := range u.Organizations {
		if !inOrganizationList(orgs, org.ID()) {
			if err := store.RemoveOrganizationMembership(ctx, u.ID(), org.ID()); err != nil {
				return err
			}
		}
	}
	u.Organizations = orgs

	// Iterate thru teams and check if already member; if not then
	// add membership to store
	for _, team := range teams {
		if !inTeamList(u.Teams, team.ID()) {
			if err := store.AddTeamMembership(ctx, u.ID(), team.ID()); err != nil {
				return err
			}
		}
	}
	// Iterate thru receiver's teams and check if in the given teams; if
	// not then remove membership from store
	for _, team := range u.Teams {
		if !inTeamList(teams, team.ID()) {
			if err := store.RemoveTeamMembership(ctx, u.ID(), team.ID()); err != nil {
				return err
			}
		}
	}
	u.Teams = teams

	return nil
}

// CanLock always returns an error because nothing can replace a user lock
func (u *User) CanLock(requestor Identity) error {
	return ErrWorkspaceAlreadyLocked
}

// CanUnlock decides whether to permits requestor to unlock a user lock
func (u *User) CanUnlock(requestor Identity, force bool) error {
	if force {
		// TODO: only grant admin user
		return nil
	}
	if user, ok := requestor.(*User); ok {
		if u.ID() == user.ID() {
			// only same user can unlock
			return nil
		}
		return ErrWorkspaceLockedByDifferentUser
	}
	// any other entity cannot unlock
	return ErrWorkspaceUnlockDenied
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
	SyncUserMemberships(ctx context.Context, user *User, orgs []*Organization, teams []*Team) (*User, error)
	// ListUsers lists users.
	ListUsers(ctx context.Context, opts UserListOptions) ([]*User, error)
	// CreateSession creates a user session.
	CreateSession(ctx context.Context, user *User, data *SessionData) (*Session, error)
	// DeleteSession deletes the session with the given token
	DeleteSession(ctx context.Context, token string) error
	// CreateToken creates a user token.
	CreateToken(ctx context.Context, userID string, opts *TokenCreateOptions) (*Token, error)
	// DeleteToken deletes a user token.
	DeleteToken(ctx context.Context, userID string, tokenID string) error
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
	OrganizationName *string
	TeamName         *string
}

type UserSpec struct {
	UserID                *string
	Username              *string
	SessionToken          *string
	AuthenticationTokenID *string
	AuthenticationToken   *string
}

type TokenCreateOptions struct {
	Description string
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
	if spec.AuthenticationTokenID != nil {
		return []interface{}{"authentication_token_id", *spec.AuthenticationTokenID}
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

func WithOrganizationMemberships(memberships ...*Organization) NewUserOption {
	return func(user *User) {
		user.Organizations = memberships
	}
}

func WithTeamMemberships(memberships ...*Team) NewUserOption {
	return func(user *User) {
		user.Teams = memberships
	}
}

func WithActiveSession(token string) NewUserOption {
	return func(user *User) {
		for _, session := range user.Sessions {
			if session.Token == token {
				session.active = true
			}
		}
	}
}

func inOrganizationList(orgs []*Organization, orgID string) bool {
	for _, org := range orgs {
		if org.ID() == orgID {
			return true
		}
	}
	return false
}

func inTeamList(teams []*Team, teamID string) bool {
	for _, team := range teams {
		if team.ID() == teamID {
			return true
		}
	}
	return false
}
