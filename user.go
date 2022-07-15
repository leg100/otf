package otf

import (
	"context"
	"net/http"
	"time"

	jsonapi "github.com/leg100/otf/http/dto"
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

func (u *User) CanAccess(organizationName *string) bool {
	// Site admin can access any organization
	if u.id == SiteAdminID {
		return true
	}
	// Normal users cannot access *any* organization...
	if organizationName == nil {
		return false
	}
	// ...but they can access an organization they are a member of.
	for _, org := range u.Organizations {
		if org.Name() == *organizationName {
			return true
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

// SyncOrganizationMemberships synchronises a user's organization memberships,
// taking an authoritative list of memberships and ensuring its memberships
// match, adding and removing memberships accordingly.
func (u *User) SyncOrganizationMemberships(ctx context.Context, authoritative []*Organization, store UserStore) error {
	// Iterate thru authoritative and if not in user's membership, add to db
	for _, auth := range authoritative {
		if !inOrganizationList(auth.ID(), u.Organizations) {
			if err := store.AddOrganizationMembership(ctx, u.ID(), auth.ID()); err != nil {
				return err
			}
		}
	}
	// Iterate thru existing and if not in authoritative list, remove from db
	for _, existing := range u.Organizations {
		if !inOrganizationList(existing.ID(), authoritative) {
			if err := store.RemoveOrganizationMembership(ctx, u.ID(), existing.ID()); err != nil {
				return err
			}
		}
	}
	// ...and update receiver too.
	u.Organizations = authoritative
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

// ToJSONAPI assembles a JSON-API DTO.
func (u *User) ToJSONAPI(req *http.Request) any {
	return &jsonapi.User{
		ID:       u.id,
		Username: u.username,
	}
}

// UserService provides methods to interact with user accounts and their
// sessions.
type UserService interface {
	// Create creates a user with the given username.
	Create(ctx context.Context, username string) (*User, error)
	// EnsureCreated retrieves a user; if they don't exist they'll be created.
	EnsureCreated(ctx context.Context, username string) (*User, error)
	// Get retrieves a user according to the spec.
	Get(ctx context.Context, spec UserSpec) (*User, error)
	// CreateSession creates a user session.
	CreateSession(ctx context.Context, user *User, data *SessionData) (*Session, error)
	// DeleteSession deletes the session with the given token
	DeleteSession(ctx context.Context, token string) error
	// CreateToken creates a user token.
	CreateToken(ctx context.Context, user *User, opts *TokenCreateOptions) (*Token, error)
	// DeleteToken deletes a user token.
	DeleteToken(ctx context.Context, user *User, tokenID string) error
	// SyncOrganizationMemberships synchronises a user's organization
	// memberships, adding and removing them accordingly.
	SyncOrganizationMemberships(ctx context.Context, user *User, orgs []*Organization) (*User, error)
}

// UserStore is a persistence store for user accounts.
type UserStore interface {
	CreateUser(ctx context.Context, user *User) error
	GetUser(ctx context.Context, spec UserSpec) (*User, error)
	ListUsers(ctx context.Context) ([]*User, error)
	DeleteUser(ctx context.Context, spec UserSpec) error
	// AddOrganizationMembership adds a user as a member of an organization
	AddOrganizationMembership(ctx context.Context, id, orgID string) error
	// RemoveOrganizationMembership removes a user as a member of an
	// organization
	RemoveOrganizationMembership(ctx context.Context, id, orgID string) error
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
		return []interface{}{"authentication_token", *spec.AuthenticationToken}
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

func WithActiveSession(token string) NewUserOption {
	return func(user *User) {
		for _, session := range user.Sessions {
			if session.Token == token {
				session.active = true
			}
		}
	}
}

func inOrganizationList(orgID string, orgs []*Organization) bool {
	for _, org := range orgs {
		if org.ID() == orgID {
			return true
		}
	}
	return false
}
