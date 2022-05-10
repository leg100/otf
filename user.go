package otf

import (
	"context"
	"time"
)

const (
	// Session data keys
	UsernameSessionKey = "username"
	AddressSessionKey  = "ip_address"
	FlashSessionKey    = "flash"

	DefaultSessionExpiry = 24 * time.Hour

	AnonymousUsername string = "anonymous"
)

// User represents an oTF user account.
type User struct {
	// ID uniquely identifies users
	ID string `db:"user_id" jsonapi:"primary,users"`

	// Username is the SSO-provided username
	Username string

	// Timestamps records timestamps of lifecycle transitions
	Timestamps

	// Name of the current Organization the user is using on the web app.
	CurrentOrganization *string

	// A user has many sessions
	Sessions []*Session

	// A user has many tokens
	Tokens []*Token

	// A user belongs to many organizations
	Organizations []*Organization
}

// AttachNewSession creates and attaches a new session to the user.
func (u *User) AttachNewSession(data *SessionData) (*Session, error) {
	session, err := NewSession(u.ID, data)
	if err != nil {
		return nil, err
	}

	u.Sessions = append(u.Sessions, session)

	return session, nil
}

// IsAuthenticated determines if the user is authenticated, i.e. not an
// anonymous user.
func (u *User) IsAuthenticated() bool {
	return u.Username != AnonymousUsername
}

func (u *User) String() string {
	return u.Username
}

// TransferSession transfers a session from the receiver to another user.
func (u *User) TransferSession(session *Session, to *User) {
	// Update session's user reference
	session.UserID = to.ID

	// Remove session from receiver
	for i, s := range u.Sessions {
		if s.Token != session.Token {
			u.Sessions = append(u.Sessions[0:i], u.Sessions[i+1:len(u.Sessions)]...)
			break
		}
	}

	// Add session to destination user
	to.Sessions = append(to.Sessions, session)
}

// UserService provides methods to interact with user accounts and their
// sessions.
type UserService interface {
	// Create creates a user with the given username.
	Create(ctx context.Context, username string) (*User, error)

	// EnsureCreated retrieves a user; if they don't exist they'll be created.
	EnsureCreated(ctx context.Context, username string) (*User, error)

	// Update updates the user. The username identifies the user to update.
	Update(ctx context.Context, username string, updated *User) error

	// Get retrieves a user according to the spec.
	Get(ctx context.Context, spec UserSpec) (*User, error)

	// Get retrieves the anonymous user.
	GetAnonymous(ctx context.Context) (*User, error)

	// CreateSession creates a user session.
	CreateSession(ctx context.Context, user *User, data *SessionData) (*Session, error)

	// UpdateSession persists any updates to the user's session data
	UpdateSession(ctx context.Context, user *User, session *Session) error

	// DeleteSession deletes the session with the given token
	DeleteSession(ctx context.Context, token string) error

	// CreateToken creates a user token.
	CreateToken(ctx context.Context, user *User, opts *TokenCreateOptions) (*Token, error)

	// DeleteToken deletes a user token.
	DeleteToken(ctx context.Context, id string) error

	// SyncOrganizationMemberships synchronises a user's organization
	// memberships, adding and removing them accordingly.
	SyncOrganizationMemberships(ctx context.Context, id string, orgs []*Organization) (*User, error)
}

// OrganizationMembershipUpdater manages changes to a user's organization
// memberships
type OrganizationMembershipUpdater interface {
	Add(ctx context.Context, orgID string) error
	Remove(ctx context.Context, orgID string) error
}

// SyncOrganizationMemberships synchronises a user's organization
// memberships, taking an authoritative list of memberships and ensuring its
// memberships match, adding and removing memberships accordingly.
func (u *User) SyncOrganizationMemberships(ctx context.Context, authoritative []*Organization, updater OrganizationMembershipUpdater) error {
	// Iterate thru authoritative and if not in user's membership, add to db
	for _, auth := range authoritative {
		if !inOrganizationList(auth.ID, u.Organizations) {
			if err := updater.Add(ctx, auth.ID); err != nil {
				return err
			}
		}
	}

	// Iterate thru existing and if not in authoritative list, remove from db
	for _, existing := range u.Organizations {
		if !inOrganizationList(existing.ID, authoritative) {
			if err := updater.Remove(ctx, existing.ID); err != nil {
				return err
			}
		}
	}

	// ...and update receiver too.
	u.Organizations = authoritative

	return nil
}

// UserStore is a persistence store for user accounts and their associated
// sessions.
type UserStore interface {
	Create(ctx context.Context, user *User) error

	// Update updates the user account in the persistence store.
	Update(ctx context.Context, spec UserSpec, updated *User) error

	Get(ctx context.Context, spec UserSpec) (*User, error)

	List(ctx context.Context) ([]*User, error)

	Delete(ctx context.Context, spec UserSpec) error

	// CreateSession persists a new session to the store.
	CreateSession(ctx context.Context, session *Session) error

	// UpdateSession persists any updates to a user's session
	UpdateSession(ctx context.Context, token string, updated *Session) error

	// DeleteSession deletes a session
	DeleteSession(ctx context.Context, token string) error

	// CreateToken creates a user token.
	CreateToken(ctx context.Context, token *Token) error

	// DeleteToken deletes a user token.
	DeleteToken(ctx context.Context, id string) error
}

type UserSpec struct {
	UserID                *string
	Username              *string
	SessionToken          *string
	AuthenticationTokenID *string
	AuthenticationToken   *string
}

type TokenCreateOptions struct {
	UserID      string
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

func NewUser(username string) *User {
	user := User{
		ID:         NewID("user"),
		Timestamps: NewTimestamps(),
		Username:   username,
	}

	return &user
}

func inOrganizationList(orgID string, orgs []*Organization) bool {
	for _, org := range orgs {
		if org.ID == orgID {
			return true
		}
	}
	return false
}
