package tokens

import (
	"net/http"
	"strings"

	"github.com/leg100/otf/internal/authz"
	otfhttp "github.com/leg100/otf/internal/http"
	"github.com/leg100/otf/internal/logr"
	"github.com/leg100/otf/internal/tfeapi"
	"github.com/leg100/otf/internal/ui/paths"
)

// AuthenticatedPrefixes are those URL path prefixes requiring authentication.
var AuthenticatedPrefixes = []string{
	tfeapi.APIPrefixV2,
	tfeapi.ModuleV1Prefix,
	otfhttp.APIBasePath,
	paths.UIPrefix,
}

type (
	Middleware struct {
		authenticators []authenticator
		logger         logr.Logger
	}

	authenticator interface {
		Authenticate(w http.ResponseWriter, r *http.Request) (authz.Subject, error)
	}
)

// Authenticate is middleware that verifies that all requests
// to protected endpoints possess a valid token.
//
// Where authentication succeeds, the authenticated subject is attached to the request
// context and the upstream handler is called.
func (m *Middleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !isProtectedPath(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}

		// Until request is authenticated, call service endpoints using
		// superuser privileges. Once authenticated, the authenticated user
		// replaces the superuser in the context.
		ctx := authz.AddSubjectToContext(r.Context(), &authz.Superuser{
			Username: "auth",
		})
		r = r.WithContext(ctx)

		for _, auth := range m.authenticators {
			subj, err := auth.Authenticate(w, r)
			if err != nil {
				m.logger.Error(err, "authenticating request")
				http.Error(w, err.Error(), http.StatusUnauthorized)
				return
			}
			if subj != nil {
				// Successfully authenticated
				ctx = authz.AddSubjectToContext(r.Context(), subj)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}
		}
		http.Error(w, "no authentication token found", http.StatusUnauthorized)
	})
}

func isProtectedPath(path string) bool {
	for _, prefix := range AuthenticatedPrefixes {
		if strings.HasPrefix(path, prefix) {
			return true
		}
	}
	return false
}
