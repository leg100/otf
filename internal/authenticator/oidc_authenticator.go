package authenticator

import (
	"context"
	"net/http"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/cloud"
	"github.com/leg100/otf/internal/http/html"
	"github.com/leg100/otf/internal/http/html/paths"
	"github.com/leg100/otf/internal/tokens"
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
		tokens.TokensService     // for creating session
		internal.HostnameService // for constructing redirect URL
		cloud.OIDCConfig
	}

	// oidcClaims depicts the claims returned from the oidc id-token.
	oidcClaims struct {
		Name   string   `json:"name"`
		Groups []string `json:"groups"`
	}
)

func newOIDCAuthenticator(ctx context.Context, opts oidcAuthenticatorOptions) (*oidcAuthenticator, error) {
	cloudConfig := cloud.Config{
		Name:                opts.Name,
		SkipTLSVerification: opts.SkipTLSVerification,
	}

	// construct oidc provider, using our own http client, which lets us disable
	// tls verification for testing purposes.
	ctx = oidc.ClientContext(ctx, cloudConfig.HTTPClient())
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
			cloudConfig: cloudConfig,
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
		html.Error(w, "id_token missing", http.StatusInternalServerError, false)
		return
	}

	// Parse and verify ID Token payload.
	idt, err := o.verifier.Verify(r.Context(), rawIDToken)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError, false)
		return
	}

	// Extract custom claims
	var claims oidcClaims
	if err := idt.Claims(&claims); err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError, false)
		return
	}

	// Get claims user
	user, err := o.getUserFromClaims(claims)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError, false)
		return
	}

	err = o.StartSession(w, r, tokens.StartSessionOptions{
		Username: &user.Name,
	})
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError, false)
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
