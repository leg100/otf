package authenticator

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"path"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/cloud"
	"github.com/leg100/otf/internal/http/decode"
	"golang.org/x/oauth2"
)

const oauthCookieName = "oauth-state"

var ErrOAuthCredentialsIncomplete = errors.New("must specify both client ID and client secret")

type (
	oauthClient interface {
		RequestHandler(w http.ResponseWriter, r *http.Request)
		CallbackHandler(*http.Request) (*oauth2.Token, error)
		NewClient(ctx context.Context, token *oauth2.Token) (cloud.Client, error)
		RequestPath() string
		CallbackPath() string
		String() string
	}

	// OAuthClient performs the client role in an oauth handshake, requesting
	// authorization from the user to access their account details on a particular
	// cloud.
	OAuthClient struct {
		internal.HostnameService // for retrieving otf system hostname for use in redirects back to otf
		cloudConfig              cloud.Config
		*oauth2.Config
	}

	// OAuthClientConfig is configuration for constructing an OAuth client
	OAuthClientConfig struct {
		cloud.CloudOAuthConfig
		otfHostname internal.HostnameService
	}
)

func NewOAuthClient(cfg OAuthClientConfig) (*OAuthClient, error) {
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

	return &OAuthClient{
		HostnameService: cfg.otfHostname,
		cloudConfig:     cfg.Config,
		Config:          cfg.OAuthConfig,
	}, nil
}

// String provides a human-readable identifier for the oauth client, using the
// name of its underlying cloud provider
func (a *OAuthClient) String() string { return a.cloudConfig.Name }

func (a *OAuthClient) RequestPath() string {
	return path.Join("/oauth", a.cloudConfig.Name, "login")
}

// RequestHandler initiates the oauth flow, redirecting user to the auth server
func (a *OAuthClient) RequestHandler(w http.ResponseWriter, r *http.Request) {
	state, err := internal.GenerateToken()
	if err != nil {
		http.Error(w, "unable to generate state token: "+err.Error(), http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     oauthCookieName,
		Value:    state,
		Path:     "/",
		MaxAge:   60, // 60 seconds
		HttpOnly: true,
		Secure:   true, // HTTPS only
	})

	cfg, err := a.config()
	if err != nil {
		http.Error(w, "unable to get redirect url: "+err.Error(), http.StatusInternalServerError)
		return
	}

	redirectURL := cfg.AuthCodeURL(state)
	http.Redirect(w, r, redirectURL, http.StatusFound)
}

func (a *OAuthClient) CallbackPath() string {
	return path.Join("/oauth", a.cloudConfig.Name, "callback")
}

func (a *OAuthClient) CallbackHandler(r *http.Request) (*oauth2.Token, error) {
	// Parse query string
	var resp struct {
		AuthCode         string `schema:"code"`
		State            string
		Error            string
		ErrorDescription string `schema:"error_description"`
		ErrorURI         string `schema:"error_uri"`
	}
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
	cfg, err := a.config()
	if err != nil {
		return nil, err
	}

	return cfg.Exchange(ctx, resp.AuthCode)
}

// NewClient constructs a cloud client configured with the given oauth token for authentication.
func (a *OAuthClient) NewClient(ctx context.Context, token *oauth2.Token) (cloud.Client, error) {
	return a.cloudConfig.NewClient(ctx, cloud.Credentials{
		OAuthToken: token,
	})
}

func (a *OAuthClient) getRedirectURL() (string, error) {
	return (&url.URL{Scheme: "https", Host: a.Hostname(), Path: a.CallbackPath()}).String(), nil
}

// config generates an oauth2 config for the client - note this is done at
// run-time because the otf hostname may only be determined at run-time.
func (a *OAuthClient) config() (*oauth2.Config, error) {
	redirectURL, err := a.getRedirectURL()
	if err != nil {
		return nil, err
	}

	return &oauth2.Config{
		Endpoint:     a.Config.Endpoint,
		ClientID:     a.Config.ClientID,
		ClientSecret: a.Config.ClientSecret,
		RedirectURL:  redirectURL,
		Scopes:       a.Config.Scopes,
	}, nil
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
