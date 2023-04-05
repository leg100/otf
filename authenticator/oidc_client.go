package authenticator

import (
	"errors"
	"github.com/coreos/go-oidc"
	"github.com/leg100/otf"
	"github.com/leg100/otf/cloud"
	"github.com/leg100/otf/http/html"
	"github.com/leg100/otf/http/html/paths"
	"github.com/leg100/otf/tokens"
	"golang.org/x/oauth2"
	"net/http"
	"path"
)

type (
	// oidcAuthenticator is an authenticator that uses oidc.
	oidcAuthenticator struct {
		tokens.TokensService // for creating session

		oidcConfig   cloud.OIDCConfig
		provider     *oidc.Provider
		verifier     *oidc.IDTokenVerifier
		oauth2Config oauth2.Config
		policies     []cloud.OIDCOrganizationPolicy
	}

	// oidcClaims depicts the claims returned from the oidc id-token.
	oidcClaims struct {
		Name   string   `json:"name"`
		Groups []string `json:"groups"`
	}
)

var (
	_ authenticator = &oidcAuthenticator{}
)

func (o oidcAuthenticator) RequestPath() string {
	return path.Join("/oauth", o.oidcConfig.Name, "login")
}

func (o oidcAuthenticator) CallbackPath() string {
	return path.Join("/oauth", o.oidcConfig.Name, "callback")
}

func (o oidcAuthenticator) RequestHandler(w http.ResponseWriter, r *http.Request) {
	state, err := otf.GenerateToken()
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

	http.Redirect(w, r, o.oauth2Config.AuthCodeURL(state), http.StatusFound)
}

func (o oidcAuthenticator) CallbackHandler(r *http.Request) (*oidcClaims, error) {
	// Verify state and errors.
	ctx := r.Context()

	oauth2Token, err := o.oauth2Config.Exchange(ctx, r.URL.Query().Get("code"))
	if err != nil {
		return nil, err
	}

	// Extract the ID Token from OAuth2 token.
	rawIDToken, ok := oauth2Token.Extra("id_token").(string)
	if !ok {
		return nil, errors.New("id_token missing")
	}

	// Parse and verify ID Token payload.
	idToken, err := o.verifier.Verify(ctx, rawIDToken)
	if err != nil {
		return nil, err
	}

	// Extract custom claims
	var claims oidcClaims
	if err := idToken.Claims(&claims); err != nil {
		return nil, err
	}

	return &claims, nil
}

func (o oidcAuthenticator) buildUserTeamsFromPolicies(
	teams []cloud.Team,
	groupsMap map[string]bool,
	policies []cloud.OIDCOrganizationPolicy,
) []cloud.Team {
	for _, policy := range policies {
		if groupsMap[policy.Group] {
			teams = append(teams, cloud.Team{
				Name:         policy.Team,
				Organization: policy.Organization,
			})
		}
	}

	return teams
}

// getUserFromClaims returns a cloud.User given a user's claims.
func (o oidcAuthenticator) getUserFromClaims(claims *oidcClaims) (*cloud.User, error) {
	var teams []cloud.Team

	groupsMap := map[string]bool{}
	for _, group := range claims.Groups {
		groupsMap[group] = true
	}

	teams = o.buildUserTeamsFromPolicies(teams, groupsMap, o.policies)
	
	return &cloud.User{
		Name:  claims.Name,
		Teams: teams,
	}, nil
}

func (o oidcAuthenticator) ResponseHandler(w http.ResponseWriter, r *http.Request) {
	// Handle oauth response; if there is an error, return user to login page
	// along with flash error.
	claims, err := o.CallbackHandler(r)
	if err != nil {
		html.FlashError(w, err.Error())
		http.Redirect(w, r, paths.Login(), http.StatusFound)
		return
	}

	// Get claims user
	cuser, err := o.getUserFromClaims(claims)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = o.StartSession(w, r, tokens.StartSessionOptions{
		Username: &cuser.Name,
	})
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (o oidcAuthenticator) String() string {
	return o.oidcConfig.Name
}
