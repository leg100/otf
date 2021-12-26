package html

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"net/http"
	"net/url"

	"golang.org/x/oauth2"
)

const (
	oauthCookieName = "oauth-state"
)

type oauth struct {
	// OAuth2 configuration for authorization
	*oauth2.Config
}

// requestHandler initiates the oauth flow, redirecting to the IdP auth url.
func (o *oauth) requestHandler(w http.ResponseWriter, r *http.Request) {
	state := randomState()

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
	cookie, err := r.Cookie(oauthCookieName)
	if err != nil {
		return nil, err
	}
	cookieState := cookie.Value

	authCode, callbackState, err := parseCallback(r)
	if err != nil {
		return nil, err
	}

	// CSRF protection - verify state in the cookie matches the state in the
	// callback URL query
	if callbackState != cookieState || callbackState == "" {
		return nil, err
	}

	// Use the authorization code to get a Token
	token, err := o.config(r).Exchange(r.Context(), authCode)
	if err != nil {
		return nil, err
	}

	return token, nil
}

func (o *oauth) config(r *http.Request) *oauth2.Config {
	redirectURL := url.URL{
		Scheme: "https",
		Host:   r.Host,
		Path:   githubCallbackPath,
	}

	cfg := o.Config
	cfg.RedirectURL = redirectURL.String()
	return cfg
}

// parseCallback parses the "code" and "state" parameters from the http.Request
// and returns them.
func parseCallback(req *http.Request) (authCode, state string, err error) {
	err = req.ParseForm()
	if err != nil {
		return "", "", err
	}
	authCode = req.Form.Get("code")
	state = req.Form.Get("state")
	if authCode == "" || state == "" {
		return "", "", errors.New("oauth2: Request missing code or state")
	}
	return authCode, state, nil
}

// Returns a base64 encoded random 32 byte string.
func randomState() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}
