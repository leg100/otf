package auth

import (
	"net/http"

	"github.com/go-logr/logr"
	"github.com/leg100/otf"
	"github.com/leg100/otf/cloud"
	"github.com/leg100/otf/http/html"
	"github.com/leg100/otf/http/html/paths"
)

// authenticator logs people onto the system using an OAuth handshake with an
// Identity provider before synchronising their user account and various organization
// and team memberships from the provider.
type authenticator struct {
	otf.HostnameService

	AuthService
	oauthClient
}

type authenticatorOptions struct {
	logr.Logger
	otf.HostnameService
	AuthService
	configs []cloud.CloudOAuthConfig
}

func newAuthenticators(opts authenticatorOptions) ([]*authenticator, error) {
	var authenticators []*authenticator

	for _, cfg := range opts.configs {
		if cfg.OAuthConfig.ClientID == "" && cfg.OAuthConfig.ClientSecret == "" {
			// skip creating oauth client when creds are unspecified
			continue
		}
		client, err := NewOAuthClient(OAuthClientConfig{
			CloudOAuthConfig: cfg,
			otfHostname:      opts.HostnameService,
		})
		if err != nil {
			return nil, err
		}
		authenticators = append(authenticators, &authenticator{
			oauthClient: client,
			AuthService: opts.AuthService,
		})
		opts.V(2).Info("activated oauth client", "name", cfg, "hostname", cfg.Hostname)
	}
	return authenticators, nil
}

// exchanging its auth code for a token.
func (a *authenticator) responseHandler(w http.ResponseWriter, r *http.Request) {
	// Handle oauth response; if there is an error, return user to login page
	// along with flash error.
	token, err := a.CallbackHandler(r)
	if err != nil {
		html.FlashError(w, err.Error())
		http.Redirect(w, r, paths.Login(), http.StatusFound)
		return
	}

	client, err := a.NewClient(r.Context(), token)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// give authenticator unlimited access to services
	ctx := otf.AddSubjectToContext(r.Context(), &otf.Superuser{Username: "authenticator"})

	// Get cloud user
	cuser, err := client.GetUser(ctx)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Synchronise cloud user with otf user
	user, err := a.sync(ctx, *cuser)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	session, err := a.createSession(r, user.ID)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	session.setCookie(w)

	returnUserOriginalPage(w, r)
}
