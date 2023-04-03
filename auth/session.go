package auth

import (
	"fmt"
	"net/http"
	"time"

	"github.com/leg100/otf"
	otfhttp "github.com/leg100/otf/http"
	"github.com/leg100/otf/http/html"
)

const (
	defaultExpiry          = 24 * time.Hour
	defaultCleanupInterval = 5 * time.Minute
)

type (
	// Session is a user session for the web UI
	Session struct {
		token     string
		expiry    time.Time
		address   string
		createdAt time.Time

		// Session belongs to a user
		username string
	}

	CreateSessionOptions struct {
		Request  *http.Request
		Username *string
		Expiry   *time.Time
	}
)

// newSession constructs a new Session
func newSession(opts CreateSessionOptions) (*Session, error) {
	// required options
	if opts.Request == nil {
		return nil, fmt.Errorf("missing HTTP request")
	}
	if opts.Username == nil {
		return nil, fmt.Errorf("missing username")
	}

	ip, err := otfhttp.GetClientIP(opts.Request)
	if err != nil {
		return nil, err
	}
	token, err := otf.GenerateToken()
	if err != nil {
		return nil, fmt.Errorf("generating session token: %w", err)
	}
	expiry := otf.CurrentTimestamp().Add(defaultExpiry)
	if opts.Expiry != nil {
		expiry = *opts.Expiry
	}

	session := Session{
		createdAt: otf.CurrentTimestamp(),
		token:     token,
		address:   ip,
		expiry:    expiry,
		username:  *opts.Username,
	}

	return &session, nil
}

func (s *Session) CreatedAt() time.Time { return s.createdAt }
func (s *Session) ID() string           { return s.token }
func (s *Session) Token() string        { return s.token }
func (s *Session) Username() string     { return s.username }
func (s *Session) Address() string      { return s.address }
func (s *Session) Expiry() time.Time    { return s.expiry }

func (s *Session) SetCookie(w http.ResponseWriter) {
	html.SetCookie(w, sessionCookie, s.token, otf.Time(s.Expiry()))
}
