package authenticator

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/http/decode"
	"github.com/leg100/otf/internal/http/html"
	"github.com/leg100/otf/internal/http/html/paths"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/user"
	"golang.org/x/oauth2"
)

const oauthCookieName = "oauth-state"

var ErrOAuthCredentialsIncomplete = errors.New("must specify both client ID and client secret")

type (
	// tokenHandler takes an OAuth access token and returns the username
	// associated with the token.
	tokenHandler interface {
		getUsername(context.Context, *oauth2.Token) (string, error)
	}

	sessionStarter interface {
		StartSession(w http.ResponseWriter, r *http.Request, userID resource.ID) error
	}

	// OAuthClient performs the client role in an oauth handshake, requesting
	// authorization from the user to access their account details on a particular
	// cloud.
	OAuthClient struct {
		OAuthConfig
		// extract username from token
		tokenHandler
		// for retrieving OTF system hostname to construct redirect URLs
		*internal.HostnameService

		sessions sessionStarter
		users    userService
	}

	// OAuthConfig is configuration for constructing an OAuth client
	OAuthConfig struct {
		Hostname            string
		ClientID            resource.ID
		ClientSecret        string
		Endpoint            oauth2.Endpoint
		Scopes              []string
		Name                string
		SkipTLSVerification bool
	}
)

func newOAuthClient(
	handler tokenHandler,
	hostnameService *internal.HostnameService,
	tokensService sessionStarter,
	userService userService,
	cfg OAuthConfig,
) (*OAuthClient, error) {
	if cfg.ClientID == "" && cfg.ClientSecret != "" {
		return nil, ErrOAuthCredentialsIncomplete
	}
	if cfg.ClientID != "" && cfg.ClientSecret == "" {
		return nil, ErrOAuthCredentialsIncomplete
	}
	// if OAuth provider hostname specified then update its OAuth endpoint
	// accordingly.
	if cfg.Hostname != "" {
		authURL, err := updateHost(cfg.Endpoint.AuthURL, cfg.Hostname)
		if err != nil {
			return nil, err
		}
		tokenURL, err := updateHost(cfg.Endpoint.TokenURL, cfg.Hostname)
		if err != nil {
			return nil, err
		}
		cfg.Endpoint.AuthURL = authURL
		cfg.Endpoint.TokenURL = tokenURL
	}
	return &OAuthClient{
		tokenHandler:    handler,
		HostnameService: hostnameService,
		sessions:        tokensService,
		OAuthConfig:     cfg,
	}, nil
}

// String provides a human-readable identifier for the oauth client, using the
// name of its underlying cloud provider
func (a *OAuthClient) String() string { return a.Name }

// requestHandler initiates the oauth flow, redirecting user to the auth server
func (a *OAuthClient) requestHandler(w http.ResponseWriter, r *http.Request) {
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
	redirectURL := a.config().AuthCodeURL(state)
	http.Redirect(w, r, redirectURL, http.StatusFound)
}

// callbackHandler handles the response from the identity provider, exchanging
// the code it receives for an access token, which it then uses to retrieve its
// corresponding username and start a new OTF user session.
func (a *OAuthClient) callbackHandler(w http.ResponseWriter, r *http.Request) {
	// Get token; if there is an error, return user to login page along with
	// flash error.
	token, err := func() (*oauth2.Token, error) {
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
		// Exchange code for an access token (optionally skipping TLS verification
		// for testing purposes).
		ctx := contextWithClient(r.Context(), a.SkipTLSVerification)
		return a.config().Exchange(ctx, resp.AuthCode)
	}()
	if err != nil {
		html.FlashError(w, err.Error())
		http.Redirect(w, r, paths.Login(), http.StatusFound)
		return
	}
	// Extract username from OAuth token
	username, err := a.getUsername(r.Context(), token)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError, false)
		return
	}
	user, err := a.users.GetUser(r.Context(), user.UserSpec{Username: &username})
	if err == internal.ErrResourceNotFound {
		user, err = a.users.Create(r.Context(), username)
	}
	// Lookup user in db, to retrieve user id to embed in session token
	// Get or create user.
	err = a.sessions.StartSession(w, r, user.ID)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError, false)
		return
	}
}

// config generates an oauth2 config for the client - note this is done at
// run-time because the redirect URL uses an OTF hostname that may only be
// determined at run-time.
func (a *OAuthClient) config() *oauth2.Config {
	return &oauth2.Config{
		Endpoint:     a.Endpoint,
		ClientID:     a.ClientID,
		ClientSecret: a.ClientSecret,
		RedirectURL:  a.URL(a.callbackPath()),
		Scopes:       a.Scopes,
	}
}

func (a *OAuthClient) RequestPath() string {
	return "/oauth/" + a.String() + "/login"
}

func (a *OAuthClient) callbackPath() string {
	return "/oauth/" + a.String() + "/callback"
}

func (a *OAuthClient) addHandlers(r *mux.Router) {
	r.HandleFunc(a.RequestPath(), a.requestHandler)
	r.HandleFunc(a.callbackPath(), a.callbackHandler)
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

// contextWithClient returns a context that embeds an OAuth2 http client.
func contextWithClient(ctx context.Context, skipTLSVerification bool) context.Context {
	if skipTLSVerification {
		return context.WithValue(ctx, oauth2.HTTPClient, &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: skipTLSVerification,
				},
			},
		})
	}
	return ctx
}
