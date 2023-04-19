package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/leg100/otf"
	"github.com/leg100/otf/rbac"
	"golang.org/x/exp/slog"
)

const (
	SiteAdminID       = "user-site-admin"
	SiteAdminUsername = "site-admin"
)

var (
	SiteAdmin             = User{ID: SiteAdminID, Username: SiteAdminUsername}
	_         otf.Subject = (*User)(nil)
)

type (
	// User represents an otf user account.
	User struct {
		ID        string // ID uniquely identifies users
		CreatedAt time.Time
		UpdatedAt time.Time
		Username  string  // username is globally unique
		SiteAdmin bool    // Indicates whether user is a site admin
		Teams     []*Team // user belongs to many teams
	}

	// UserListOptions are options for the ListUsers endpoint.
	UserListOptions struct {
		Organization *string
		TeamName     *string
	}

	NewUserOption func(*User)

	UserSpec struct {
		UserID                *string
		Username              *string
		AuthenticationTokenID *string
	}
)

func NewUser(username string, opts ...NewUserOption) *User {
	user := &User{
		ID:        otf.NewID("user"),
		Username:  username,
		CreatedAt: otf.CurrentTimestamp(),
		UpdatedAt: otf.CurrentTimestamp(),
	}
	for _, fn := range opts {
		fn(user)
	}
	return user
}

func WithTeams(memberships ...*Team) NewUserOption {
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
func (u *User) Organizations() (organizations []string) {
	// De-dup organizations
	seen := make(map[string]bool)
	for _, t := range u.Teams {
		if _, ok := seen[t.Organization]; ok {
			continue
		}
		organizations = append(organizations, t.Organization)
		seen[t.Organization] = true
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
	// Only site admin can perform actions on the site
	return u.IsSiteAdmin()
}

func (u *User) CanAccessOrganization(action rbac.Action, org string) bool {
	// coarser-grained site-level perms take precedence
	if u.CanAccessSite(action) {
		return true
	}
	// fallback to finer-grained organization-level perms
	for _, team := range u.Teams {
		if team.Organization == org {
			if team.IsOwners() {
				// owner team members can perform all actions on organization
				return true
			}
			if rbac.OrganizationMinPermissions.IsAllowed(action) {
				return true
			}
			if team.Access.ManageWorkspaces {
				if rbac.WorkspaceManagerRole.IsAllowed(action) {
					return true
				}
			}
			if team.Access.ManageVCS {
				if rbac.VCSManagerRole.IsAllowed(action) {
					return true
				}
			}
			if team.Access.ManageRegistry {
				if rbac.VCSManagerRole.IsAllowed(action) {
					return true
				}
			}
		}
	}
	return false
}

func (u *User) CanAccessWorkspace(action rbac.Action, policy otf.WorkspacePolicy) bool {
	// coarser-grained organization perms take precedence.
	if u.CanAccessOrganization(action, policy.Organization) {
		return true
	}
	// fallback to checking finer-grained workspace perms
	for _, team := range u.Teams {
		if team.Organization != policy.Organization {
			continue
		}
		for _, perm := range policy.Permissions {
			if team.Name == perm.Team {
				return perm.Role.IsAllowed(action)
			}
		}
	}
	return false
}

// IsOwner determines if user is an owner of an organization
func (u *User) IsOwner(organization string) bool {
	for _, team := range u.Teams {
		if team.Organization == organization {
			if team.IsOwners() {
				return true
			}
		}
	}
	return false
}

func (s *UserSpec) LogValue() slog.Value {
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
	subj, err := otf.SubjectFromContext(ctx)
	if err != nil {
		return nil, err
	}
	user, ok := subj.(*User)
	if !ok {
		return nil, fmt.Errorf("no user in context")
	}
	return user, nil
}
