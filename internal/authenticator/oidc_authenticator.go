package authenticator

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/cloud"
	"github.com/leg100/otf/internal/http/html"
	"github.com/leg100/otf/internal/http/html/paths"
	"github.com/leg100/otf/internal/tokens"
	"golang.org/x/oauth2"
)

// List of valid claims that can be used as username
const (
	EmailClaim UsernameClaim = "email"
	SubClaim   UsernameClaim = "sub"
	NameClaim  UsernameClaim = "name"
)

var (
	_ authenticator = &oidcAuthenticator{}

	// "openid" is a required scope for OpenID Connect flows, and profile
	// gives OTF access to the user's username.
	DefaultScopes = []string{oidc.ScopeOpenID, "profile"}

	ErrMissingOIDCIssuerURL = errors.New("missing oidc-issuer-url")
)

type (
	// oidcAuthenticator is an authenticator that uses OIDC.
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

	// OIDC claim that can be used as a username
	UsernameClaim string
)

func newOIDCAuthenticator(ctx context.Context, opts oidcAuthenticatorOptions) (*oidcAuthenticator, error) {
	if opts.IssuerURL == "" {
		return nil, ErrMissingOIDCIssuerURL
	}

	cloudConfig := cloud.Config{
		Name:                opts.Name,
		SkipTLSVerification: opts.SkipTLSVerification,
	}

	// construct oidc provider, using our own http client, which lets us disable
	// tls verification for testing purposes.
	ctx = oidc.ClientContext(ctx, cloudConfig.HTTPClient())
	provider, err := oidc.NewProvider(ctx, opts.IssuerURL)
	if err != nil {
		return nil, fmt.Errorf("constructing OIDC provider: %w", err)
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
				Scopes:       opts.Scopes,
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

	// oidcClaims depicts the claims returned from the OIDC id-token.
	type oidcClaims struct {
		Name  string `json:"name"`
		Sub   string `json:"sub"`
		Email string `json:"email"`
	}

	// Extract username from claim
	var (
		claims   oidcClaims
		username string
	)
	if err := idt.Claims(&claims); err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError, false)
		return
	}

	switch o.oidcConfig.UsernameClaim {
	case string(EmailClaim):
		username = claims.Email
	case string(SubClaim):
		username = claims.Sub
	case string(NameClaim):
		username = claims.Name
	}

	err = o.StartSession(w, r, tokens.StartSessionOptions{
		Username: &username,
	})
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError, false)
		return
	}
}
