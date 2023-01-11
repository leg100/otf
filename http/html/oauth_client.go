package html

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"path"

	"github.com/leg100/otf"
	"github.com/leg100/otf/cloud"
	"github.com/leg100/otf/http/decode"
	"golang.org/x/oauth2"
)

const oauthCookieName = "oauth-state"

var ErrOAuthCredentialsIncomplete = errors.New("must specify both client ID and client secret")

// OAuthClient performs the client role in an oauth handshake, requesting
// authorization from the user to access their account details on a particular
// cloud.
type OAuthClient struct {
	cloudConfig cloud.Config
	*oauth2.Config
}

// NewOAuthClientConfig is configuration for constructing an OAuth client
type NewOAuthClientConfig struct {
	*cloud.CloudOAuthConfig
	hostname string // otf system hostname
}

func NewOAuthClient(cfg NewOAuthClientConfig) (*OAuthClient, error) {
	if cfg.OAuthConfig.ClientID == "" && cfg.OAuthConfig.ClientSecret != "" {
		return nil, ErrOAuthCredentialsIncomplete
	}
	if cfg.OAuthConfig.ClientID != "" && cfg.OAuthConfig.ClientSecret == "" {
		return nil, ErrOAuthCredentialsIncomplete
	}

	authURL, err := updateHost(cfg.OAuthConfig.Endpoint.AuthURL, cfg.Hostname)
	if err != nil {
		return nil, err
	}
	tokenURL, err := updateHost(cfg.OAuthConfig.Endpoint.TokenURL, cfg.Hostname)
	if err != nil {
		return nil, err
	}
	cfg.OAuthConfig.Endpoint.AuthURL = authURL
	cfg.OAuthConfig.Endpoint.TokenURL = tokenURL

	client := &OAuthClient{cloudConfig: cfg.Config, Config: cfg.OAuthConfig}
	cfg.OAuthConfig.RedirectURL = (&url.URL{Scheme: "https", Host: cfg.hostname, Path: client.CallbackPath()}).String()

	return client, nil
}

// String provides a human-readable identifier for the oauth client, using the
// name of its underlying cloud provider
func (a *OAuthClient) String() string { return a.cloudConfig.Name }

func (a *OAuthClient) RequestPath() string {
	return path.Join("/oauth", a.cloudConfig.Name, "login")
}

// RequestHandler initiates the oauth flow, redirecting user to the auth server
func (a *OAuthClient) RequestHandler(w http.ResponseWriter, r *http.Request) {
	state, err := otf.GenerateToken()
	if err != nil {
		http.Error(w, "unable to generate state token: "+err.Error(), http.StatusInternalServerError)
		return
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

	http.Redirect(w, r, a.AuthCodeURL(state), http.StatusFound)
}

func (a *OAuthClient) CallbackPath() string {
	return path.Join("/oauth", a.cloudConfig.Name, "callback")
}

func (a *OAuthClient) CallbackHandler(r *http.Request) (*oauth2.Token, error) {
	// Parse query string
	type response struct {
		AuthCode         string `schema:"code"`
		State            string
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

	ctx := context.WithValue(r.Context(), oauth2.HTTPClient, a.cloudConfig.HTTPClient())

	// Exchange code for an access token
	return a.Exchange(ctx, resp.AuthCode)
}

func (a *OAuthClient) NewClient(ctx context.Context, token *oauth2.Token) (cloud.Client, error) {
	return a.cloudConfig.NewClient(ctx, cloud.Credentials{
		OAuthToken: token,
	})
}

// updateHost updates the hostname in a URL
func updateHost(u, host string) (string, error) {
	parsed, err := url.Parse(u)
	if err != nil {
		return "", err
	}

	parsed.Host = host

	return parsed.String(), nil
}
