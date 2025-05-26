package authenticator

import (
	"context"
	"errors"
	"fmt"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/leg100/otf/internal/user"
	"golang.org/x/oauth2"
)

var (
	// "openid" is a required scope for OpenID Connect flows, and profile
	// gives OTF access to the user's username.
	DefaultOIDCScopes       = []string{oidc.ScopeOpenID, "profile"}
	ErrMissingOIDCIssuerURL = errors.New("missing oidc-issuer-url")
)

type (
	// idtokenHandler handles specifically an OIDC ID token, extracting the
	// username from a claim within the token.
	idtokenHandler struct {
		provider *oidc.Provider
		verifier *oidc.IDTokenVerifier
		username *usernameClaim
	}

	// OIDCConfig is the configuration for a generic OIDC provider.
	OIDCConfig struct {
		// Name is the user-friendly identifier of the OIDC endpoint.
		Name string
		// IssuerURL is the issuer url for the OIDC provider.
		IssuerURL string
		// ClientID is the client id for the OIDC provider.
		ClientID string
		// ClientSecret is the client secret for the OIDC provider.
		ClientSecret string
		// Scopes to request from the OIDC provider.
		Scopes []string
		// UsernameClaim is the claim that provides the username.
		UsernameClaim string
	}
)

func newIDTokenHandler(ctx context.Context, opts OIDCConfig) (*idtokenHandler, error) {
	if opts.IssuerURL == "" {
		return nil, ErrMissingOIDCIssuerURL
	}
	// construct oidc provider, using our own http client, which lets us disable
	// tls verification for testing purposes.
	ctx = contextWithClient(ctx)
	provider, err := oidc.NewProvider(ctx, opts.IssuerURL)
	if err != nil {
		return nil, fmt.Errorf("constructing OIDC provider: %w", err)
	}

	// parse claim to be used for username
	username, err := newUsernameClaim(opts.UsernameClaim)
	if err != nil {
		return nil, err
	}

	return &idtokenHandler{
		verifier: provider.Verifier(&oidc.Config{ClientID: opts.ClientID}),
		username: username,
		provider: provider,
	}, nil
}

func (o idtokenHandler) getUsername(ctx context.Context, token *oauth2.Token) (user.Username, error) {
	// Extract the ID Token from OAuth2 token.
	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		return user.Username{}, errors.New("id_token missing")
	}

	// Parse and verify ID Token payload.
	idt, err := o.verifier.Verify(ctx, rawIDToken)
	if err != nil {
		return user.Username{}, err
	}

	// Extract username from claim
	if err := idt.Claims(&o.username); err != nil {
		return user.Username{}, err
	}

	return o.username.value, nil
}
