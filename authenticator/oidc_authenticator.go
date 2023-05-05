package authenticator

import (
	"context"
	"net/http"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/leg100/otf"
	"github.com/leg100/otf/cloud"
	"github.com/leg100/otf/http/html"
	"github.com/leg100/otf/http/html/paths"
	"github.com/leg100/otf/tokens"
	"golang.org/x/oauth2"
)

var _ authenticator = &oidcAuthenticator{}

type (
	// oidcAuthenticator is an authenticator that uses oidc.
	oidcAuthenticator struct {
		tokens.TokensService // for creating session

		oidcConfig cloud.OIDCConfig
		provider   *oidc.Provider
		verifier   *oidc.IDTokenVerifier

		oauthClient
	}

	oidcAuthenticatorOptions struct {
		tokens.TokensService // for creating session
		otf.HostnameService  // for constructing redirect URL
		cloud.OIDCConfig
	}

	// oidcClaims depicts the claims returned from the oidc id-token.
	oidcClaims struct {
		Name   string   `json:"name"`
		Groups []string `json:"groups"`
	}
)

func newOIDCAuthenticator(ctx context.Context, opts oidcAuthenticatorOptions) (*oidcAuthenticator, error) {
	provider, err := oidc.NewProvider(ctx, opts.IssuerURL)
	if err != nil {
		return nil, err
	}

	return &oidcAuthenticator{
		TokensService: opts.TokensService,
		oidcConfig:    opts.OIDCConfig,
		provider:      provider,
		verifier:      provider.Verifier(&oidc.Config{ClientID: opts.ClientID}),
		oauthClient: &OAuthClient{
			HostnameService: opts.HostnameService,
			Config: &oauth2.Config{
				ClientID:     opts.ClientID,
				ClientSecret: opts.ClientSecret,
				Endpoint:     provider.Endpoint(),

				// "openid" is a required scope for OpenID Connect flows.
				// groups is used for managing permissions.
				Scopes: append(opts.Scopes, oidc.ScopeOpenID, "groups", "profile"),
			},
			cloudConfig: cloud.Config{
				Name: opts.Name,
			},
		},
	}, nil
}

func (o oidcAuthenticator) ResponseHandler(w http.ResponseWriter, r *http.Request) {
	// Handle oauth response; if there is an error, return user to login page
	// along with flash error.
	token, err := o.CallbackHandler(r)
	if err != nil {
		html.FlashError(w, err.Error())
		http.Redirect(w, r, paths.Login(), http.StatusFound)
		return
	}

	// Extract the ID Token from OAuth2 token.
	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		html.Error(w, "id_token missing", http.StatusInternalServerError)
		return
	}

	// Parse and verify ID Token payload.
	idt, err := o.verifier.Verify(r.Context(), rawIDToken)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Extract custom claims
	var claims oidcClaims
	if err := idt.Claims(&claims); err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Get claims user
	user, err := o.getUserFromClaims(claims)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = o.StartSession(w, r, tokens.StartSessionOptions{
		Username: &user.Name,
	})
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// getUserFromClaims returns a cloud.User given a user's claims.
func (o oidcAuthenticator) getUserFromClaims(claims oidcClaims) (*cloud.User, error) {
	var teams []cloud.Team

	return &cloud.User{
		Name:  claims.Name,
		Teams: teams,
	}, nil
}
