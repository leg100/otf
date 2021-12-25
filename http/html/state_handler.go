package html

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"net/http"

	"github.com/dghubble/gologin/v2"
	"github.com/gorilla/mux"
)

// unexported key type prevents collisions
type key int

const (
	tokenKey key = iota
	stateKey
)

// StateHandler checks for a state cookie. If found, the state value is read and
// added to the ctx. Otherwise, a non-guessable value is added to the ctx and to
// a (short-lived) state cookie issued to the requester.
//
// Implements OAuth 2 RFC 6749 10.12 CSRF Protection. If you wish to issue
// state params differently, write a http.Handler which sets the ctx state,
// using oauth2 WithState(ctx, state) since it is required by LoginHandler
// and CallbackHandler.
func newStateHandler(config gologin.CookieConfig) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			cookie, err := r.Cookie(config.Name)
			if err == nil {
				// add the cookie state to the ctx
				ctx = context.WithValue(ctx, stateKey, cookie.Value)
			} else {
				// add Cookie with a random state
				val := randomState()
				http.SetCookie(w, NewCookie(config, val))
				ctx = context.WithValue(ctx, stateKey, val)
			}
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// Returns a base64 encoded random 32 byte string.
func randomState() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}
