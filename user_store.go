package otf

import "context"

// UserStore is a persistence store for user accounts.
type UserStore interface {
	Create(ctx context.Context, user *User) error

	// SetCurrentOrganization sets the user's currently active organization
	SetCurrentOrganization(ctx context.Context, userID, orgName string) error

	Get(ctx context.Context, spec UserSpec) (*User, error)

	List(ctx context.Context) ([]*User, error)

	Delete(ctx context.Context, spec UserSpec) error

	// AddOrganizationMembership adds a user as a member of an organization
	AddOrganizationMembership(ctx context.Context, id, orgID string) error
	// RemoveOrganizationMembership removes a user as a member of an
	// organization
	RemoveOrganizationMembership(ctx context.Context, id, orgID string) error
}
