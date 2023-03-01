package auth

import (
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/leg100/otf"
	"github.com/leg100/otf/http/html"
	"github.com/leg100/otf/http/html/paths"
)

const (
	defaultExpiry          = 24 * time.Hour
	defaultCleanupInterval = 5 * time.Minute
)

// Session is a user session for the web UI
type Session struct {
	token     string
	expiry    time.Time
	address   string
	createdAt time.Time

	// Session belongs to a user
	userID string
}

// newSession constructs a new Session
func newSession(r *http.Request, userID string) (*Session, error) {
	ip, err := getClientIP(r)
	if err != nil {
		return nil, err
	}

	token, err := otf.GenerateToken()
	if err != nil {
		return nil, fmt.Errorf("generating session token: %w", err)
	}

	session := Session{
		createdAt: otf.CurrentTimestamp(),
		token:     token,
		address:   ip,
		expiry:    otf.CurrentTimestamp().Add(defaultExpiry),
		userID:    userID,
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

// getClientIP gets the client's IP address
func getClientIP(r *http.Request) (string, error) {
	// reverse proxy adds client IP to an HTTP header, and each successive proxy
	// adds a client IP, so we want the leftmost IP.
	if hdr := r.Header.Get("X-Forwarded-For"); hdr != "" {
		first, _, _ := strings.Cut(hdr, ",")
		addr := strings.TrimSpace(first)
		if isIP := net.ParseIP(addr); isIP != nil {
			return addr, nil
		}
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	return host, err
}

// returnUserOriginalPage returns a user to the original page they tried to
// access before they were redirected elsewhere.
func returnUserOriginalPage(w http.ResponseWriter, r *http.Request) {
	// Return user to the original path they attempted to access
	if cookie, err := r.Cookie(pathCookie); err == nil {
		html.SetCookie(w, pathCookie, "", &time.Time{})
		http.Redirect(w, r, cookie.Value, http.StatusFound)
	} else {
		http.Redirect(w, r, paths.Profile(), http.StatusFound)
	}
}
