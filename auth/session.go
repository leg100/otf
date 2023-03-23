package auth

import (
	"fmt"
	"net/http"
	"time"

	"github.com/leg100/otf"
	otfhttp "github.com/leg100/otf/http"
	"github.com/leg100/otf/http/html"
	"github.com/leg100/otf/http/html/paths"
)

const (
	defaultExpiry          = 24 * time.Hour
	defaultCleanupInterval = 5 * time.Minute
	// path cookie stores the last path the user attempted to access
	pathCookie = "path"
)

type (
	// Session is a user session for the web UI
	Session struct {
		token     string
		expiry    time.Time
		address   string
		createdAt time.Time

		// Session belongs to a user
		userID string
	}

	CreateSessionOptions struct {
		Request *http.Request
		UserID  *string
		Expiry  *time.Time
	}
)

// newSession constructs a new Session
func newSession(opts CreateSessionOptions) (*Session, error) {
	// required options
	if opts.Request == nil {
		return nil, fmt.Errorf("missing HTTP request")
	}
	if opts.UserID == nil {
		return nil, fmt.Errorf("missing user ID")
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
		userID:    *opts.UserID,
	}

	return &session, nil
}

func (s *Session) CreatedAt() time.Time { return s.createdAt }
func (s *Session) ID() string           { return s.token }
func (s *Session) Token() string        { return s.token }
func (s *Session) UserID() string       { return s.userID }
func (s *Session) Address() string      { return s.address }
func (s *Session) Expiry() time.Time    { return s.expiry }

func (s *Session) setCookie(w http.ResponseWriter) {
	html.SetCookie(w, sessionCookie, s.token, otf.Time(s.Expiry()))
}

// sendUserToLoginPage sends user to the login prompt page, saving the original
// page they tried to access so it can return them there after login.
func sendUserToLoginPage(w http.ResponseWriter, r *http.Request) {
	html.SetCookie(w, pathCookie, r.URL.Path, nil)
	http.Redirect(w, r, paths.Login(), http.StatusFound)
}

// returnUserOriginalPage returns a user to the original page they tried to
// access before they were redirected to the login page.
func returnUserOriginalPage(w http.ResponseWriter, r *http.Request) {
	// Return user to the original path they attempted to access
	if cookie, err := r.Cookie(pathCookie); err == nil {
		html.SetCookie(w, pathCookie, "", &time.Time{})
		http.Redirect(w, r, cookie.Value, http.StatusFound)
	} else {
		http.Redirect(w, r, paths.Profile(), http.StatusFound)
	}
}
