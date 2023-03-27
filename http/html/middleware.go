package html

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

// unexported key type prevents collisions
type ctxKey int

const (
	organizationCtxKey ctxKey = iota
	// organizationCookie stores the current organization for the session
	organizationCookie = "organization"
)

func newOrganizationContext(ctx context.Context, organizationName string) context.Context {
	return context.WithValue(ctx, organizationCtxKey, organizationName)
}

func organizationFromContext(ctx context.Context) (string, error) {
	name, ok := ctx.Value(organizationCtxKey).(string)
	if !ok {
		return "", fmt.Errorf("no organization in context")
	}
	return name, nil
}

// SetOrganization ensures the session's organization reflects the most recently
// visited organization route.
func SetOrganization(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		current, ok := mux.Vars(r)["organization_name"]
		cookie, err := r.Cookie(organizationCookie)
		if ok {
			if err == http.ErrNoCookie || current != cookie.Value {
				// update session organization
				SetCookie(w, organizationCookie, current, nil)
			}
		} else {
			if err == http.ErrNoCookie {
				// not yet visited an organization route
				next.ServeHTTP(w, r)
				return
			}
			// restore session org from cookie
			current = cookie.Value
		}
		// wrap session organization in context
		ctx := newOrganizationContext(r.Context(), current)
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}

// UIRouter wraps the given router with a router suitable for web UI routes.
func UIRouter(r *mux.Router) *mux.Router {
	r = r.PathPrefix("/app").Subrouter()
	r.Use(SetOrganization)
	return r
}
