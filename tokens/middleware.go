package tokens

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/leg100/otf"
	"github.com/leg100/otf/auth"
	otfhttp "github.com/leg100/otf/http"
	"github.com/leg100/otf/http/html"
	"github.com/leg100/otf/http/html/paths"
	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"google.golang.org/api/idtoken"
)

const (
	// HTTP header in Google Cloud IAP request containing JWT
	googleIAPHeader string = "x-goog-iap-jwt-assertion"
)

type (
	middlewareOptions struct {
		AgentTokenService
		auth.AuthService

		GoogleIAPConfig
		SiteToken string

		key jwk.Key
	}

	GoogleIAPConfig struct {
		Audience string
	}

	middleware struct {
		middlewareOptions
	}
)

// newMiddleware constructs middleware that verifies that all requests
// to protected endpoints possess a valid token, applying the following logic:
//
// 1. Skip authentication for non-protected paths and allow request.
// 2. If Google IAP header is present then authenticate its token and allow or deny
// accordingly.
// 3. If Bearer token is present then authenticate it and allow or deny accordingly.
// 4. If requested path is for a UI endpoint then check for session cookie. If
// present then authenticate its token. If cookie is missing or authentication fails
// then redirect user to login page.
// 5. Otherwise, return 401
//
// Where authentication succeeds, the authenticated subject is attached to the request
// context and the upstream handler is called.
func newMiddleware(opts middlewareOptions) mux.MiddlewareFunc {
	mw := middleware{middlewareOptions: opts}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var (
				subject otf.Subject
				err     error
			)
			if !isProtectedPath(r.URL.Path) {
				next.ServeHTTP(w, r)
				return
			}
			if token := r.Header.Get(googleIAPHeader); token != "" {
				subject, err = mw.validateIAPToken(r.Context(), token)
				if err != nil {
					http.Error(w, err.Error(), http.StatusUnauthorized)
					return
				}
			} else if bearer := r.Header.Get("Authorization"); bearer != "" {
				subject, err = mw.validateBearer(r.Context(), bearer)
				if err != nil {
					http.Error(w, err.Error(), http.StatusUnauthorized)
					return
				}
			} else if strings.HasPrefix(r.URL.Path, paths.UIPrefix) {
				var ok bool
				subject, ok = mw.validateUIRequest(w, r)
				if !ok {
					html.SendUserToLoginPage(w, r)
					return
				}
			} else {
				http.Error(w, "no authentication token found", http.StatusUnauthorized)
				return
			}
			ctx := otf.AddSubjectToContext(r.Context(), subject)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func (m *middleware) validateIAPToken(ctx context.Context, token string) (otf.Subject, error) {
	payload, err := idtoken.Validate(ctx, token, m.Audience)
	if err != nil {
		return nil, err
	}
	email, ok := payload.Claims["email"]
	if !ok {
		return nil, fmt.Errorf("IAP token is missing email claim")
	}
	return m.GetUser(ctx, auth.UserSpec{Username: otf.String(email.(string))})
}

func (m *middleware) validateBearer(ctx context.Context, bearer string) (otf.Subject, error) {
	splitToken := strings.Split(bearer, "Bearer ")
	if len(splitToken) != 2 {
		return nil, fmt.Errorf("malformed bearer token")
	}
	token := splitToken[1]

	if m.SiteToken != "" && m.SiteToken == token {
		return &auth.SiteAdmin, nil
	}
	//
	// parse jwt from cookie and verify signature
	parsed, err := jwt.Parse([]byte(token), jwt.WithKey(jwa.HS256, m.key))
	if err != nil {
		return nil, err
	}
	kindClaim, ok := parsed.Get("kind")
	if !ok {
		return nil, fmt.Errorf("missing claim: kind")
	}
	switch kind(kindClaim.(string)) {
	case agentTokenKind:
		return m.GetAgentToken(ctx, parsed.Subject())
	case userTokenKind:
		return m.GetUser(ctx, auth.UserSpec{AuthenticationTokenID: otf.String(parsed.Subject())})
	case registrySessionKind:
		return NewRegistrySessionFromJWT(parsed)
	default:
		return nil, fmt.Errorf("unknown authentication kind")
	}
}

func (m *middleware) validateUIRequest(w http.ResponseWriter, r *http.Request) (otf.Subject, bool) {
	cookie, err := r.Cookie(sessionCookie)
	if err == http.ErrNoCookie {
		return nil, false
	}
	// parse jwt from cookie and verify signature
	token, err := jwt.Parse([]byte(cookie.Value), jwt.WithKey(jwa.HS256, m.key))
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired()) {
			html.FlashError(w, "session expired")
		} else {
			html.FlashError(w, "unable to verify session token: "+err.Error())
		}
		return nil, false
	}
	user, err := m.GetUser(r.Context(), auth.UserSpec{
		Username: otf.String(token.Subject()),
	})
	if err != nil {
		html.FlashError(w, "unable to find user: "+err.Error())
		return nil, false
	}
	return user, true
}

func isProtectedPath(path string) bool {
	for _, prefix := range otfhttp.AuthenticatedPrefixes {
		if strings.HasPrefix(path, prefix) {
			return true
		}
	}
	return false
}
