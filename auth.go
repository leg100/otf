package otf

import (
	"context"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

const (
	SiteAdminID     = "user-site-admin"
	DefaultUserID   = "user-123"
	DefaultUsername = "otf"
)

type RegistrySession interface {
	Token() string
	Organization() string
	Expiry() time.Time

	Subject
}

type RegistrySessionService interface {
	// AddHandlers adds handlers for the http api.
	AddHandlers(*mux.Router)

	RegistrySessionApp
}

type RegistrySessionApp interface {
	CreateRegistrySession(ctx context.Context, organization string) (RegistrySession, error)
	GetRegistrySession(ctx context.Context, token string) (RegistrySession, error)
}

type User interface {
	Username() string
	IsSiteAdmin() bool
	Organizations() []string

	Subject
}

// UserService provides methods to interact with user accounts and their
// sessions.
type UserService interface {
	// Get retrieves a user according to the spec.
	GetUser(ctx context.Context, spec UserSpec) (User, error)
}

type UserSpec struct {
	UserID              *string
	Username            *string
	SessionToken        *string
	AuthenticationToken *string
}

func (s UserSpec) MarshalLog() any {
	if s.AuthenticationToken != nil {
		s.AuthenticationToken = String("*****")
	}
	return s
}

type Team interface {
	ID() string
	Name() string
	Organization() string
	IsOwners() bool
}

type TeamService interface {
	GetTeam(ctx context.Context, teamID string) (Team, error)
}

type CreateTeamOptions struct {
	Name         string `schema:"team_name,required"`
	Organization string `schema:"organization_name,required"`
}

type AgentToken interface {
	Token() string
	Organization() string

	Subject
}

type CreateAgentTokenOptions struct {
	Organization string `schema:"organization_name,required"`
	Description  string `schema:"description,required"`
}

// AgentTokenService provides access to agent tokens
type AgentTokenService interface {
	GetAgentToken(ctx context.Context, token string) (AgentToken, error)
}

type Session interface {
	Expiry() time.Time
	SetCookie(w http.ResponseWriter)
}

type SessionService interface {
	// CreateSession creates a user session.
	CreateSession(r *http.Request, userID string) (Session, error)
	// ListSessions lists current sessions for a user
	ListSessions(ctx context.Context, userID string) ([]Session, error)
	// DeleteSession deletes the session with the given token
	DeleteSession(ctx context.Context, token string) error
}
