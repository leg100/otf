package authn

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/leg100/otf/internal/authz"
	"github.com/leg100/otf/internal/logr"
)

var errLoginNeeded = errors.New("you need to login to access the requested page")

type (
	// Middleware is HTTP middleware that authenticates requests.
	Middleware struct {
		Authenticators []Authenticator
		Route          AuthenticationRoute
		Logger         logr.Logger
	}

	// Authenticator authenticates a request; if successful it returns the
	// authenticated subject; if invalid an error is returned; if there is no
	// appropriate means of authentication found in the request then a nil
	// subject is returned.
	Authenticator interface {
		Authenticate(w http.ResponseWriter, r *http.Request) (authz.Subject, error)
	}

	// AuthenticationRoute is particular route that is authenticated, e.g. api, ui,
	// etc.
	AuthenticationRoute interface {
		// IsPath returns true if the path is within the route that is
		// authenticated.
		IsPath(path string) bool
		// HandleError handles the error to be returned to the client in the
		// event of authentication failure.
		HandleError(err error, w http.ResponseWriter, r *http.Request)
	}
)

// Authenticate authenticates inbound http requests. Where authentication
// succeeds, the authenticated subject is attached to the request context and
// the upstream handler is called.
func (a *Middleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !a.Route.IsPath(r.URL.Path) {
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

		subject, err := a.authenticate(w, r)
		if err != nil {
			a.Route.HandleError(err, w, r)
			return
		}

		ctx = authz.AddSubjectToContext(r.Context(), subject)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (a *Middleware) authenticate(w http.ResponseWriter, r *http.Request) (authz.Subject, error) {
	for _, auth := range a.Authenticators {
		subj, err := auth.Authenticate(w, r)
		if err != nil {
			a.Logger.Error(err, "authenticating request")
			return nil, err
		}
		if subj != nil {
			// Successfully authenticated
			return subj, nil
		}
	}
	return nil, fmt.Errorf("you need to login to access the requested page")
}
