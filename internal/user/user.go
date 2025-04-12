// Package user manages user accounts and their team membership.
package user

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/authz"
	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/team"
)

var (
	SiteAdminUsername = Username{name: "site-admin"}
	// SiteAdminID is the hardcoded user id for the site admin user. The ID must
	// be the same as the hardcoded value in the database migrations.
	SiteAdminID               = resource.MustHardcodeTfeID(resource.UserKind, "36atQC2oGQng7pVz")
	SiteAdmin                 = User{ID: SiteAdminID, Username: SiteAdminUsername}
	_           authz.Subject = (*User)(nil)
)

type (
	// User represents an OTF user account.
	User struct {
		ID        resource.TfeID `jsonapi:"primary,users" json:"user_id"`
		CreatedAt time.Time      `jsonapi:"attribute" json:"created-at"`
		UpdatedAt time.Time      `jsonapi:"attribute" json:"updated-at"`
		SiteAdmin bool           `jsonapi:"attribute" json:"site-admin"`

		// username is globally unique
		Username Username `jsonapi:"attribute" json:"username"`

		// user belongs to many teams
		Teams []*team.Team
	}

	// ListOptions are options for the ListUsers endpoint.
	ListOptions struct {
		resource.PageOptions

		Organization organization.Name `schema:"name"`
	}

	NewUserOption func(*User)

	CreateUserOptions struct {
		Username string `json:"username"`
	}

	UserSpec struct {
		UserID                *resource.TfeID
		Username              *Username
		AuthenticationTokenID *resource.TfeID
	}
)

func NewUser(usernameStr string, opts ...NewUserOption) (*User, error) {
	username, err := NewUsername(usernameStr)
	if err != nil {
		return nil, err
	}
	user := &User{
		ID:        resource.NewTfeID(resource.UserKind),
		Username:  username,
		CreatedAt: internal.CurrentTimestamp(nil),
		UpdatedAt: internal.CurrentTimestamp(nil),
	}
	for _, fn := range opts {
		fn(user)
	}
	return user, nil
}

func WithTeams(memberships ...*team.Team) NewUserOption {
	return func(user *User) {
		user.Teams = memberships
	}
}

func (u *User) String() string { return u.Username.String() }

// IsTeamMember determines whether user is a member of the given team.
func (u *User) IsTeamMember(teamID resource.TfeID) bool {
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
func (u *User) Organizations() []organization.Name {
	// De-dup organizations using map
	seen := make(map[organization.Name]bool)
	for _, t := range u.Teams {
		if _, ok := seen[t.Organization]; ok {
			continue
		}
		seen[t.Organization] = true
	}

	// Turn map into slice
	organizations := make([]organization.Name, len(seen))
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

func (u *User) CanAccess(action authz.Action, req authz.Request) bool {
	// Site admin can do whatever it wants
	if u.IsSiteAdmin() {
		return true
	}
	switch action {
	case authz.CreateOrganizationAction, authz.GetGithubAppAction, authz.GetUserAction:
		// These actions are available to any user at any level.
		return true
	case authz.CreateUserAction, authz.ListUsersAction:
		// A user can perform these actions only if they are an owner of at
		// least one organization. This permits an owner to search users or create
		// a user before adding them to a team.
		for _, team := range u.Teams {
			if team.IsOwners() {
				return true
			}
		}
	}
	if req.ID == resource.SiteID {
		// no further allowed actions that are available to user at the
		// site-level.
		return false
	}
	// All other user perms are inherited from team memberships.
	for _, team := range u.Teams {
		if team.CanAccess(action, req) {
			return true
		}
	}
	return false
}

// IsOwner determines if user is an owner of an organization
func (u *User) IsOwner(organization resource.ID) bool {
	for _, team := range u.Teams {
		if team.IsOwner(organization) {
			return true
		}
	}
	return false
}

func (s UserSpec) LogValue() slog.Value {
	if s.Username != nil {
		return slog.Any("username", *s.Username).Value
	}
	if s.UserID != nil {
		return slog.String("id", s.UserID.String()).Value
	}
	if s.AuthenticationTokenID != nil {
		return slog.String("token_id", "*****").Value
	}
	return slog.String("unknown key", "unknown value").Value
}

// UserFromContext retrieves a user from a context
func UserFromContext(ctx context.Context) (*User, error) {
	subj, err := authz.SubjectFromContext(ctx)
	if err != nil {
		return nil, err
	}
	user, ok := subj.(*User)
	if !ok {
		return nil, fmt.Errorf("no user in context")
	}
	return user, nil
}
