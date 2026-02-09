package tokens

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal/authz"
	otfhttp "github.com/leg100/otf/internal/http"
	"github.com/leg100/otf/internal/http/html"
	"github.com/leg100/otf/internal/logr"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/tfeapi"
	"github.com/leg100/otf/internal/ui/paths"
	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"google.golang.org/api/idtoken"
)

const (
	// HTTP header in Google Cloud IAP request containing JWT
	googleIAPHeader string = "x-goog-iap-jwt-assertion"
)

// AuthenticatedPrefixes are those URL path prefixes requiring authentication.
var AuthenticatedPrefixes = []string{
	tfeapi.APIPrefixV2,
	tfeapi.ModuleV1Prefix,
	otfhttp.APIBasePath,
	paths.UIPrefix,
}

type (
	middlewareOptions struct {
		GoogleIAPConfig
		logr.Logger

		key jwk.Key

		*registry
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
// context and the upstream handler is called. If the authenticated subject is a
// user and the user does not exist the user is first created.
func newMiddleware(opts middlewareOptions) mux.MiddlewareFunc {
	mw := middleware{middlewareOptions: opts}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var (
				subject authz.Subject
				err     error
			)
			// Until request is authenticated, call service endpoints using
			// superuser privileges. Once authenticated, the authenticated user
			// replaces the superuser in the context.
			ctx := authz.AddSubjectToContext(r.Context(), &authz.Superuser{
				Username: "auth",
			})

			if !isProtectedPath(r.URL.Path) {
				next.ServeHTTP(w, r)
				return
			}
			if token := r.Header.Get(googleIAPHeader); token != "" {
				subject, err = mw.validateIAPToken(ctx, token)
				if err != nil {
					http.Error(w, err.Error(), http.StatusUnauthorized)
					return
				}
			} else if bearer := r.Header.Get("Authorization"); bearer != "" {
				subject, err = mw.validateBearer(ctx, bearer)
				if err != nil {
					mw.Error(err, "validating bearer token")
					http.Error(w, err.Error(), http.StatusUnauthorized)
					return
				}
			} else if strings.HasPrefix(r.URL.Path, paths.UIPrefix) {
				var ok bool
				subject, ok = mw.validateUIRequest(ctx, w, r)
				if !ok {
					html.SendUserToLoginPage(w, r)
					return
				}
			} else {
				http.Error(w, "no authentication token found", http.StatusUnauthorized)
				return
			}
			ctx = authz.AddSubjectToContext(r.Context(), subject)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func (m *middleware) validateIAPToken(ctx context.Context, token string) (authz.Subject, error) {
	payload, err := idtoken.Validate(ctx, token, m.Audience)
	if err != nil {
		return nil, err
	}
	email, ok := payload.Claims["email"]
	if !ok {
		return nil, errors.New("IAP token is missing email claim")
	}
	emailString, ok := email.(string)
	if !ok {
		return nil, fmt.Errorf("expected IAP token email to be a string: %#v", email)
	}
	return m.GetOrCreateUser(ctx, emailString)
}

func (m *middleware) validateBearer(ctx context.Context, bearer string) (authz.Subject, error) {
	splitToken := strings.Split(bearer, "Bearer ")
	if len(splitToken) != 2 {
		return nil, fmt.Errorf("malformed bearer token")
	}
	token := splitToken[1]
	if m.SiteToken != "" && m.SiteToken == token {
		return m.SiteAdmin, nil
	}
	id, err := m.parseIDFromJWT([]byte(token))
	if err != nil {
		return nil, err
	}
	return m.GetSubject(ctx, id)
}

func (m *middleware) validateUIRequest(ctx context.Context, w http.ResponseWriter, r *http.Request) (authz.Subject, bool) {
	cookie, err := r.Cookie(SessionCookie)
	if err == http.ErrNoCookie {
		html.FlashSuccess(w, "you need to login to access the requested page")
		return nil, false
	}
	// parse jwt from cookie and verify signature
	id, err := m.parseIDFromJWT([]byte(cookie.Value))
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired()) {
			html.FlashError(w, "session expired")
		} else {
			html.FlashError(w, "unable to verify session token: "+err.Error())
		}
		return nil, false
	}
	user, err := m.GetSubject(ctx, id)
	if err != nil {
		html.FlashError(w, "unable to find user: "+err.Error())
		return nil, false
	}
	return user, true
}

func (m *middleware) parseIDFromJWT(token []byte) (resource.TfeID, error) {
	parsed, err := jwt.Parse(token, jwt.WithKey(jwa.HS256, m.key))
	if err != nil {
		return resource.TfeID{}, err
	}
	return resource.ParseTfeID(parsed.Subject())
}

func isProtectedPath(path string) bool {
	for _, prefix := range AuthenticatedPrefixes {
		if strings.HasPrefix(path, prefix) {
			return true
		}
	}
	return false
}
