package session

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/otf"
	"github.com/leg100/otf/http/html"
	"github.com/leg100/otf/http/html/paths"
)

const (
	// session cookie stores the session token
	sessionCookie = "session"
)

// Authenticate middleware checks incoming request possesses a valid session cookie,
// attaching its user and the session to the context.
func Authenticate(svc otf.UserService) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie(sessionCookie)
			if err == http.ErrNoCookie {
				sendUserToLoginPage(w, r)
				return
			}
			user, err := svc.GetUser(r.Context(), otf.UserSpec{
				SessionToken: &cookie.Value,
			})
			if err != nil {
				html.FlashError(w, "unable to find user: "+err.Error())
				sendUserToLoginPage(w, r)
				return
			}

			// add user and session token to context for use by upstream handlers
			ctx := otf.AddSubjectToContext(r.Context(), user)
			ctx = addToContext(ctx, cookie.Value)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func sendUserToLoginPage(w http.ResponseWriter, r *http.Request) {
	html.SetCookie(w, otf.PathCookie, r.URL.Path, nil)
	http.Redirect(w, r, paths.Login(), http.StatusFound)
}
