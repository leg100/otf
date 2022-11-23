package html

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"path"
	"time"

	"github.com/leg100/otf"
	otfhttp "github.com/leg100/otf/http"
	"github.com/leg100/otf/http/decode"
	"golang.org/x/oauth2"
)

// authPrefix is the prefix for all authentication routes
var authPrefix = "/oauth"

const oauthCookieName = "oauth-state"

// Authenticator logs people onto the system using an OAuth handshake with an
// Identity provider before synchronising their user account and various organization
// and team memberships from the provider.
type Authenticator struct {
	// OAuth identity provider config
	*otf.CloudConfig
	otf.Application
}

func newAuthenticators(app otf.Application, configs cloudDB) ([]*Authenticator, error) {
	var authenticators []*Authenticator

	for _, cfg := range configs {
		if cfg.ClientID == "" && cfg.ClientSecret == "" {
			// skip cloud providers for which no oauth credentials have been
			// configured
			continue
		}
		authenticators = append(authenticators, &Authenticator{
			CloudConfig: cfg,
			Application: app,
		})
	}

	return authenticators, nil
}

// String provides a human readable name for the authenticator, reflecting the
// name of the cloud providing authentication.
func (a *Authenticator) String() string {
	return string(a.Name)
}

func (a *Authenticator) RequestPath() string {
	return path.Join(authPrefix, string(a.Name), "login")
}

func (a *Authenticator) callbackPath() string {
	return path.Join(authPrefix, string(a.Name), "callback")
}

// requestHandler initiates the oauth flow, redirecting user to the IdP auth
// endpoint.
func (a *Authenticator) requestHandler(w http.ResponseWriter, r *http.Request) {
	state, err := otf.GenerateToken()
	if err != nil {
		// TODO: explicitly return 500
		panic("unable to generate state token: " + err.Error())
	}

	// TODO: replace with setCookie helper
	http.SetCookie(w, &http.Cookie{
		Name:     oauthCookieName,
		Value:    state,
		Path:     "/",
		MaxAge:   60, // 60 seconds
		HttpOnly: true,
		Secure:   true, // HTTPS only
	})

	http.Redirect(w, r, a.oauthCfg(r).AuthCodeURL(state), http.StatusFound)
}

// oauthConfig generates an OAuth configuration - the current request is
// necessary for obtaining the hostname and scheme for the redirect URL
func (a *Authenticator) oauthCfg(r *http.Request) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     a.ClientID,
		ClientSecret: a.ClientSecret,
		Endpoint:     a.Endpoint,
		Scopes:       a.Scopes,
		RedirectURL:  otfhttp.Absolute(r, a.callbackPath()),
	}
}

// responseHandler completes the oauth flow, handling the callback response and
// exchanging its auth code for a token.
func (a *Authenticator) responseHandler(w http.ResponseWriter, r *http.Request) {
	// Handle oauth response; if there is an error, return user to login page
	// along with flash error.
	token, err := a.handleResponse(r)
	if err != nil {
		flashError(w, err.Error())
		http.Redirect(w, r, loginPath(), http.StatusFound)
		return
	}

	client, err := a.Cloud.NewClient(r.Context(), otf.ClientConfig{
		Hostname:            a.Hostname,
		SkipTLSVerification: a.SkipTLSVerification,
		OAuthToken:          token,
	})
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	user, err := a.synchronise(r.Context(), client)
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := createSession(a.Application, w, r, user.ID()); err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return user to the original path they attempted to access
	if cookie, err := r.Cookie(pathCookie); err == nil {
		setCookie(w, pathCookie, "", &time.Time{})
		http.Redirect(w, r, cookie.Value, http.StatusFound)
	} else {
		http.Redirect(w, r, getProfilePath(), http.StatusFound)
	}
}

func (a *Authenticator) handleResponse(r *http.Request) (*oauth2.Token, error) {
	// Parse query string
	type response struct {
		AuthCode string `schema:"code"`
		State    string

		Error            string
		ErrorDescription string `schema:"error_description"`
		ErrorURI         string `schema:"error_uri"`
	}
	var resp response
	if err := decode.Query(&resp, r.URL.Query()); err != nil {
		return nil, err
	}
	if resp.Error != "" {
		return nil, fmt.Errorf("%s: %s\n\nSee %s", resp.Error, resp.ErrorDescription, resp.ErrorURI)
	}

	// Validate state
	cookie, err := r.Cookie(oauthCookieName)
	if err != nil {
		return nil, fmt.Errorf("missing state cookie (the cookie expires after 60 seconds)")
	}
	if resp.State != cookie.Value || resp.State == "" {
		return nil, fmt.Errorf("state mismatch between cookie and callback response")
	}

	// Optionally skip TLS verification of auth code endpoint
	ctx := r.Context()
	if a.SkipTLSVerification {
		ctx = context.WithValue(ctx, oauth2.HTTPClient, &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		})
	}

	// Exchange code for an access token
	return a.oauthCfg(r).Exchange(ctx, resp.AuthCode)
}

func (a *Authenticator) synchronise(ctx context.Context, client otf.CloudClient) (*otf.User, error) {
	// give authenticator unlimited access to services
	ctx = otf.AddSubjectToContext(ctx, &otf.Superuser{Username: "authenticator"})

	// Get cloud user
	cuser, err := client.GetUser(ctx)
	if err != nil {
		return nil, err
	}

	// Get otf user; if not exist, create user
	user, err := a.EnsureCreatedUser(ctx, cuser.Username())
	if err != nil {
		return nil, err
	}

	// organizations to be synchronised
	var organizations []*otf.Organization
	// teams to be synchronised
	var teams []*otf.Team

	// Create user's organizations as necessary
	for _, org := range cuser.Organizations() {
		org, err = a.EnsureCreatedOrganization(ctx, otf.OrganizationCreateOptions{
			Name: otf.String(org.Name()),
		})
		if err != nil {
			return nil, err
		}
		organizations = append(organizations, org)
	}

	// A user also gets their own personal organization matching their username
	personal, err := a.EnsureCreatedOrganization(ctx, otf.OrganizationCreateOptions{
		Name: otf.String(user.Username()),
	})
	if err != nil {
		return nil, err
	}
	organizations = append(organizations, personal)

	// Create user's teams as necessary
	for _, team := range cuser.Teams() {
		team, err = a.EnsureCreatedTeam(ctx, team.Name(), team.Organization().Name())
		if err != nil {
			return nil, err
		}
		teams = append(teams, team)
	}
	// And make them an owner of their personal org
	team, err := a.EnsureCreatedTeam(ctx, "owners", personal.Name())
	if err != nil {
		return nil, err
	}
	teams = append(teams, team)

	// Synchronise user's memberships so that they match those of the cloud
	// user.
	if _, err = a.SyncUserMemberships(ctx, user, organizations, teams); err != nil {
		return nil, err
	}

	return user, nil
}
