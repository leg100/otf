package middleware

import (
	"errors"
	"net/http"
	"strings"

	"github.com/leg100/otf/internal/authz"
	otfhttp "github.com/leg100/otf/internal/http"
	"github.com/leg100/otf/internal/logr"
	uipath "github.com/leg100/otf/internal/path"
	"github.com/leg100/otf/internal/tfeapi"
	"github.com/leg100/otf/internal/ui/helpers"
)

// AuthenticatedPrefixes are those URL path prefixes requiring authentication.
var AuthenticatedPrefixes = []string{
	tfeapi.APIPrefixV2,
	tfeapi.ModuleV1Prefix,
	otfhttp.APIBasePath,
	uipath.Prefix,
}

type (
	Middleware struct {
		// APIAuthenticators are authenticators that authenticate access to APIs
		APIAuthenticators []Authenticator
		// UIAuthenticators are authenticators that authenticate access to the UI
		UIAuthenticators []Authenticator
		Logger           logr.Logger
	}

	Authenticator interface {
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
		// Until request is authenticated, call service endpoints using
		// superuser privileges. Once authenticated, the authenticated user
		// replaces the superuser in the context.
		ctx := authz.AddSubjectToContext(r.Context(), &authz.Superuser{
			Username: "auth",
		})
		r = r.WithContext(ctx)

		var (
			subject authz.Subject
			err     error
		)

		if strings.HasPrefix(r.URL.Path, uipath.Prefix) {
			subject, err = m.authenticate(m.UIAuthenticators, w, r)
			if err != nil {
				helpers.FlashError(w, err.Error())
				helpers.SendUserToLoginPage(w, r)
			}
		} else if isAPIPath(r.URL.Path) {
			subject, err = m.authenticate(m.APIAuthenticators, w, r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusUnauthorized)
			}
		} else {
			// Unprotected path that doesn't require authentication
			next.ServeHTTP(w, r)
			return
		}
		ctx = authz.AddSubjectToContext(r.Context(), subject)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func isAPIPath(path string) bool {
	return strings.HasPrefix(path, tfeapi.APIPrefixV2) ||
		strings.HasPrefix(path, tfeapi.ModuleV1Prefix) ||
		strings.HasPrefix(path, otfhttp.APIBasePath)
}

func (m *Middleware) authenticate(authenticators []Authenticator, w http.ResponseWriter, r *http.Request) (authz.Subject, error) {
	for _, auth := range authenticators {
		subj, err := auth.Authenticate(w, r)
		if err != nil {
			m.Logger.Error(err, "authenticating request")
			return nil, err
		}
		if subj != nil {
			// Successfully authenticated
			return subj, nil
		}
	}
	return nil, errors.New("no authentication token found")
}
