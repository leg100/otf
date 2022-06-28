package html

import (
	"fmt"
	"net/http"

	"github.com/leg100/otf"
	"github.com/leg100/otf/http/decode"
	httputil "github.com/leg100/otf/http/util"
	"golang.org/x/oauth2"
)

const (
	oauthCookieName = "oauth-state"
)

type oauth struct {
	// OAuth2 configuration for authorization
	*oauth2.Config
}

type oauthResponse struct {
	AuthCode string `schema:"code"`
	State    string

	Error            string
	ErrorDescription string `schema:"error_description"`
	ErrorURI         string `schema:"error_uri"`
}

// requestHandler initiates the oauth flow, redirecting to the IdP auth url.
func (o *oauth) requestHandler(w http.ResponseWriter, r *http.Request) {
	state, err := otf.GenerateToken()
	if err != nil {
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

	authURL := o.config(r).AuthCodeURL(state)
	http.Redirect(w, r, authURL, http.StatusFound)
}

// responseHandler completes the oauth flow, handling the callback response and
// exchanging its auth code for a token.
func (o *oauth) responseHandler(r *http.Request) (*oauth2.Token, error) {
	var resp oauthResponse
	if err := decode.Query(&resp, r.URL.Query()); err != nil {
		return nil, err
	}
	if resp.Error != "" {
		return nil, fmt.Errorf("%s: %s\n\nSee %s", resp.Error, resp.ErrorDescription, resp.ErrorURI)
	}

	cookie, err := r.Cookie(oauthCookieName)
	if err != nil {
		return nil, err
	}
	cookieState := cookie.Value

	// CSRF protection - verify state in the cookie matches the state in the
	// callback response
	if resp.State != cookieState || resp.State == "" {
		return nil, fmt.Errorf("state mismatch between cookie and callback response")
	}

	// Use the authorization code to get a Token
	token, err := o.config(r).Exchange(r.Context(), resp.AuthCode)
	if err != nil {
		return nil, err
	}

	return token, nil
}

func (o *oauth) config(r *http.Request) *oauth2.Config {
	cfg := o.Config
	cfg.RedirectURL = httputil.Absolute(r, githubCallbackPath)
	return cfg
}
