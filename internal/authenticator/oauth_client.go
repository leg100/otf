package authenticator

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/a-h/templ"
	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/authz"
	otfhttp "github.com/leg100/otf/internal/http"
	"github.com/leg100/otf/internal/http/decode"
	"github.com/leg100/otf/internal/http/html"
	"github.com/leg100/otf/internal/http/html/paths"
	"github.com/leg100/otf/internal/logr"
	"github.com/leg100/otf/internal/resource"
	userpkg "github.com/leg100/otf/internal/user"
	"golang.org/x/oauth2"
)

const oauthCookieName = "oauth-state"

var ErrOAuthCredentialsIncomplete = errors.New("must specify both client ID and client secret")

type (
	// tokenHandler takes an OAuth access token and returns the username
	// associated with the token.
	tokenHandler interface {
		parseUserInfo(context.Context, *oauth2.Token) (UserInfo, error)
	}

	sessionStarter interface {
		StartSession(w http.ResponseWriter, r *http.Request, userID resource.TfeID) error
	}

	// OAuthClient performs the client role in an oauth handshake, requesting
	// authorization from the user to access their account details on a particular
	// cloud.
	OAuthClient struct {
		logr.Logger
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
		// Name of oauth client. Should be a lowercase string because it is used
		// in URL paths.
		Name                string
		BaseURL             *internal.WebURL
		ClientID            string
		ClientSecret        string
		Endpoint            oauth2.Endpoint
		Scopes              []string
		SkipTLSVerification bool
		Icon                templ.Component
	}
)

func newOAuthClient(
	logger logr.Logger,
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
	if cfg.BaseURL != nil {
		authURL, err := updateHost(cfg.Endpoint.AuthURL, cfg.BaseURL.Host)
		if err != nil {
			return nil, err
		}
		tokenURL, err := updateHost(cfg.Endpoint.TokenURL, cfg.BaseURL.Host)
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
		users:           userService,
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
	// Extract user info from OAuth token
	userInfo, err := a.parseUserInfo(r.Context(), token)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Retrieve user with extracted username, creating the user if necessary.
	//
	// Use privileged context to authorize access to user service endpoints.
	ctx := authz.AddSubjectToContext(r.Context(), &authz.Superuser{Username: "oauth_client"})
	user, err := a.users.GetUser(ctx, userpkg.UserSpec{Username: &userInfo.Username})
	if errors.Is(err, internal.ErrResourceNotFound) {
		user, err = a.users.Create(ctx, userInfo.Username.String())
	}
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// If there is an avatar then refresh user's avatar regardless of whether
	// it's changed from last time.
	if userInfo.AvatarURL != nil {
		if err := a.users.UpdateAvatar(ctx, user.Username, *userInfo.AvatarURL); err != nil {
			// Just log the error because a failure to update a user's avatar
			// should not prevent them from logging in.
			a.Error(err, "updating user avatar url")
			return
		}
	}

	if err := a.sessions.StartSession(w, r, user.ID); err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
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
			Transport: otfhttp.InsecureTransport,
		})
	}
	return ctx
}
