package html

import (
	"net/http"

	"github.com/gorilla/mux"
)

// setOrganization ensures the session's organization reflects the most recently
// visited organization route.
func setOrganization(next http.Handler) http.Handler {
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
