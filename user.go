package otf

import "context"

const SiteAdminID = "user-site-admin"

type User interface {
	Username() string

	Subject
}

// UserService provides methods to interact with user accounts and their
// sessions.
type UserService interface {
	// EnsureCreatedUser retrieves a user; if they don't exist they'll be
	// created.
	EnsureCreatedUser(ctx context.Context, username string) (User, error)
	// SyncUserMemberships makes the user a member of the specified organizations
	// and teams and removes any existing memberships not specified.
	SyncUserMemberships(ctx context.Context, user User, orgs []string, teams []Team) (User, error)
	// Get retrieves a user according to the spec.
	GetUser(ctx context.Context, spec UserSpec) (User, error)
}

type UserSpec struct {
	UserID              *string
	Username            *string
	SessionToken        *string
	AuthenticationToken *string
}
