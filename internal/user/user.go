// Package user manages user accounts and their team membership.
package user

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/rbac"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/team"
)

const (
	SiteAdminID       = "user-site-admin"
	SiteAdminUsername = "site-admin"
)

var (
	SiteAdmin                  = User{ID: SiteAdminID, Username: SiteAdminUsername}
	_         internal.Subject = (*User)(nil)
)

type (
	// User represents an OTF user account.
	User struct {
		ID        string    `jsonapi:"primary,users"`
		CreatedAt time.Time `jsonapi:"attribute" json:"created-at"`
		UpdatedAt time.Time `jsonapi:"attribute" json:"updated-at"`
		SiteAdmin bool      `jsonapi:"attribute" json:"site-admin"`

		// username is globally unique
		Username string `jsonapi:"attribute" json:"username"`

		// user belongs to many teams
		Teams []*team.Team
	}

	// UserListOptions are options for the ListUsers endpoint.
	UserListOptions struct {
		Organization *string
		TeamName     *string
	}

	NewUserOption func(*User)

	CreateUserOptions struct {
		Username string `json:"username"`
	}

	UserSpec struct {
		UserID                *string
		Username              *string
		AuthenticationTokenID *string
	}
)

func NewUser(username string, opts ...NewUserOption) *User {
	user := &User{
		ID:        resource.NewID("user"),
		Username:  username,
		CreatedAt: internal.CurrentTimestamp(nil),
		UpdatedAt: internal.CurrentTimestamp(nil),
	}
	for _, fn := range opts {
		fn(user)
	}
	return user
}

func WithTeams(memberships ...*team.Team) NewUserOption {
	return func(user *User) {
		user.Teams = memberships
	}
}

func (u *User) String() string { return u.Username }

// IsTeamMember determines whether user is a member of the given team.
func (u *User) IsTeamMember(teamID string) bool {
	for _, t := range u.Teams {
		if t.ID == teamID {
			return true
		}
	}
	return false
}

// Organizations returns the user's membership of organizations (indirectly via
// their membership of teams).
//
// NOTE: always returns a non-nil slice
func (u *User) Organizations() []string {
	// De-dup organizations using map
	seen := make(map[string]bool)
	for _, t := range u.Teams {
		if _, ok := seen[t.Organization]; ok {
			continue
		}
		seen[t.Organization] = true
	}

	// Turn map into slice
	organizations := make([]string, len(seen))
	var i int
	for org := range seen {
		organizations[i] = org
		i++
	}
	return organizations
}

// IsSiteAdmin determines whether user is a site admin. A user is a site admin
// in either of two cases:
// (1) their account has been promoted to site admin (think sudo)
// (2) the account is *the* site admin (think root)
func (u *User) IsSiteAdmin() bool {
	return u.SiteAdmin || u.ID == SiteAdminID
}

func (u *User) CanAccessSite(action rbac.Action) bool {
	switch action {
	case rbac.GetGithubAppAction:
		return true
	case rbac.CreateUserAction, rbac.ListUsersAction:
		// A user can perform these actions only if they are an owner of at
		// least one organization. This permits an owner to search users or create
		// a user before adding them to a team.
		for _, team := range u.Teams {
			if team.IsOwners() {
				return true
			}
		}
	}
	// Otherwise only the site admin can perform site actions.
	return u.IsSiteAdmin()
}

func (u *User) CanAccessTeam(action rbac.Action, teamID string) bool {
	// coarser-grained site-level perms take precedence
	if u.CanAccessSite(action) {
		return true
	}
	for _, team := range u.Teams {
		if team.ID == teamID {
			return true
		}
	}
	return false
}

func (u *User) CanAccessOrganization(action rbac.Action, org string) bool {
	// coarser-grained site-level perms take precedence
	if u.CanAccessSite(action) {
		return true
	}
	// fallback to finer-grained organization-level perms
	for _, team := range u.Teams {
		if team.CanAccessOrganization(action, org) {
			return true
		}
	}
	return false
}

func (u *User) CanAccessWorkspace(action rbac.Action, policy internal.WorkspacePolicy) bool {
	// coarser-grained organization perms take precedence.
	if u.CanAccessOrganization(action, policy.Organization) {
		return true
	}
	// fallback to checking finer-grained workspace perms
	for _, team := range u.Teams {
		if team.CanAccessWorkspace(action, policy) {
			return true
		}
	}
	return false
}

// IsOwner determines if user is an owner of an organization
func (u *User) IsOwner(organization string) bool {
	for _, team := range u.Teams {
		if team.IsOwner(organization) {
			return true
		}
	}
	return false
}

func (s UserSpec) LogValue() slog.Value {
	if s.Username != nil {
		return slog.String("username", *s.Username).Value
	}
	if s.UserID != nil {
		return slog.String("id", *s.UserID).Value
	}
	if s.AuthenticationTokenID != nil {
		return slog.String("token_id", "*****").Value
	}
	return slog.String("unknown key", "unknown value").Value
}

// UserFromContext retrieves a user from a context
func UserFromContext(ctx context.Context) (*User, error) {
	subj, err := internal.SubjectFromContext(ctx)
	if err != nil {
		return nil, err
	}
	user, ok := subj.(*User)
	if !ok {
		return nil, fmt.Errorf("no user in context")
	}
	return user, nil
}
