package html

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"path"
	"time"

	"github.com/leg100/otf"
	"github.com/leg100/otf/http/decode"
	"golang.org/x/oauth2"
)

// authPrefix is the prefix for all authentication routes
var authPrefix = "/oauth"

const oauthCookieName = "oauth-state"

// Authenticator logs people onto the system, synchronising their
// user account and various organization and team memberships from an external
// directory.
type Authenticator struct {
	Cloud
	otf.Application
}

// NewAuthenticatorsFromConfig constructs authenticators from the given cloud
// configurations. If config is unspecified then its corresponding cloud
// authenticator is skipped.
func NewAuthenticatorsFromConfig(app otf.Application, configs ...CloudConfig) ([]*Authenticator, error) {
	var authenticators []*Authenticator
	for _, cfg := range configs {
		err := cfg.Valid()
		if errors.Is(err, ErrOAuthCredentialsUnspecified) {
			continue
		} else if err != nil {
			return nil, fmt.Errorf("invalid cloud config: %w", err)
		}
		cloud, err := cfg.NewCloud()
		if err != nil {
			return nil, err
		}
		authenticators = append(authenticators, &Authenticator{cloud, app})
	}
	return authenticators, nil
}

func (a *Authenticator) RequestPath() string {
	return path.Join(authPrefix, a.CloudName(), "login")
}

func (a *Authenticator) callbackPath() string {
	return path.Join(authPrefix, a.CloudName(), "callback")
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

	cfg, err := a.oauthConfig(r)
	if err != nil {
		panic("unable to generate oauth config: " + err.Error())
	}
	http.Redirect(w, r, cfg.AuthCodeURL(state), http.StatusFound)
}

// oauthConfig generates an OAuth configuration - the current request is
// necessary for obtaining the hostname and scheme for the redirect URL
func (a *Authenticator) oauthConfig(r *http.Request) (*oauth2.Config, error) {
	endpoint, err := a.Endpoint()
	if err != nil {
		return nil, err
	}
	return &oauth2.Config{
		ClientID:     a.ClientID(),
		ClientSecret: a.ClientSecret(),
		Endpoint:     endpoint,
		Scopes:       a.Scopes(),
		RedirectURL:  otf.Absolute(r, a.callbackPath()),
	}, nil
}

// responseHandler completes the oauth flow, handling the callback response and
// exchanging its auth code for a token.
func (a *Authenticator) responseHandler(w http.ResponseWriter, r *http.Request) {
	// Generate oauth config
	cfg, err := a.oauthConfig(r)
	if err != nil {
		flashError(w, err.Error())
		http.Redirect(w, r, loginPath(), http.StatusFound)
		return
	}

	// Handle oauth response; if there is an error, return user to login page
	// along with flash error.
	token, err := a.handleResponse(r, cfg)
	if err != nil {
		flashError(w, err.Error())
		http.Redirect(w, r, loginPath(), http.StatusFound)
		return
	}

	client, err := a.NewDirectoryClient(r.Context(), DirectoryClientOptions{
		Token:  token,
		Config: cfg,
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

	// create session data
	data, err := otf.NewSessionData(r)
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	session, err := a.CreateSession(r.Context(), user, data)
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	setCookie(w, sessionCookie, session.Token, &session.Expiry)

	// Return user to the original path they attempted to access
	if cookie, err := r.Cookie(pathCookie); err == nil {
		setCookie(w, pathCookie, "", &time.Time{})
		http.Redirect(w, r, cookie.Value, http.StatusFound)
	} else {
		http.Redirect(w, r, getProfilePath(), http.StatusFound)
	}
}

func (a *Authenticator) handleResponse(r *http.Request, cfg *oauth2.Config) (*oauth2.Token, error) {
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
	if a.SkipTLSVerification() {
		ctx = context.WithValue(ctx, oauth2.HTTPClient, &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		})
	}

	// Exchange code for an access token
	return cfg.Exchange(ctx, resp.AuthCode)
}

func (a *Authenticator) synchronise(ctx context.Context, client DirectoryClient) (*otf.User, error) {
	// service calls are made using the privileged app user
	ctx = otf.AddSubjectToContext(ctx, &otf.AppUser{})

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

	// Create user's organizations as necessary
	for i, org := range cuser.Organizations {
		org, err = a.EnsureCreatedOrganization(ctx, otf.OrganizationCreateOptions{
			Name: otf.String(org.Name()),
		})
		if err != nil {
			return nil, err
		}
		cuser.Organizations[i] = org
	}

	// A user also gets their own personal organization that matches their
	// username
	org, err := a.EnsureCreatedOrganization(ctx, otf.OrganizationCreateOptions{
		Name: otf.String(user.Username()),
	})
	if err != nil {
		return nil, err
	}
	cuser.Organizations = append(cuser.Organizations, org)

	// Create user's teams as necessary
	for i, team := range cuser.Teams {
		team, err = a.EnsureCreatedTeam(ctx, team.Name(), team.OrganizationName())
		if err != nil {
			return nil, err
		}
		cuser.Teams[i] = team
	}

	// Synchronise user's memberships so that they match those of the cloud
	// user.
	if _, err = a.SyncUserMemberships(ctx, user, cuser.Organizations, cuser.Teams); err != nil {
		return nil, err
	}

	return user, nil
}
