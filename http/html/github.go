package html

import (
	"context"
	"fmt"
	"net/http"

	"github.com/dghubble/gologin/v2"
	"golang.org/x/oauth2"
)

// LoginHandler handles OAuth2 login requests by reading the state value from
// the ctx and redirecting requests to the AuthURL with that state value.
func LoginHandler(config *oauth2.Config, failure http.Handler) http.Handler {
	if failure == nil {
		failure = gologin.DefaultFailureHandler
	}
	fn := func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		state, err := StateFromContext(ctx)
		if err != nil {
			ctx = gologin.WithError(ctx, err)
			failure.ServeHTTP(w, req.WithContext(ctx))
			return
		}
		authURL := config.AuthCodeURL(state)
		http.Redirect(w, req, authURL, http.StatusFound)
	}
	return http.HandlerFunc(fn)
}

// StateFromContext returns the state value from the ctx.
func StateFromContext(ctx context.Context) (string, error) {
	state, ok := ctx.Value(stateKey).(string)
	if !ok {
		return "", fmt.Errorf("oauth2: Context missing state value")
	}
	return state, nil
}
