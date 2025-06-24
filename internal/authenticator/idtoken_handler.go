package authenticator

import (
	"context"
	"errors"
	"fmt"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/leg100/otf/internal/user"
	"golang.org/x/oauth2"
)

// claim is the name of the claim which provides the username in an ID token.
type claim string

const (
	EmailClaim   claim = "email"
	SubClaim           = "sub"
	NameClaim          = "name"
	DefaultClaim       = NameClaim
)

var (
	// "openid" is a required scope for OpenID Connect flows, and profile
	// gives OTF access to the user's username.
	DefaultOIDCScopes       = []string{oidc.ScopeOpenID, "profile"}
	ErrMissingOIDCIssuerURL = errors.New("missing oidc-issuer-url")
)

type (
	// idTokenHandler handles specifically an OIDC ID token, extracting the
	// username from a claim within the token.
	idTokenHandler struct {
		provider      *oidc.Provider
		verifier      *oidc.IDTokenVerifier
		usernameClaim claim
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
		// Skip TLS Verification when communicating with issuer.
		SkipTLSVerification bool
		// Scopes to request from the OIDC provider.
		Scopes []string
		// UsernameClaim is the claim that provides the username.
		UsernameClaim string
	}
)

func newIDTokenHandler(ctx context.Context, opts OIDCConfig) (*idTokenHandler, error) {
	if opts.IssuerURL == "" {
		return nil, ErrMissingOIDCIssuerURL
	}

	switch claim(opts.UsernameClaim) {
	case EmailClaim, SubClaim, NameClaim:
	default:
		return nil, errors.New("unknown username claim: must be one of email, sub, or name")
	}

	// construct oidc provider, using our own http client, which lets us disable
	// tls verification for testing purposes.
	ctx = contextWithClient(ctx, opts.SkipTLSVerification)
	provider, err := oidc.NewProvider(ctx, opts.IssuerURL)
	if err != nil {
		return nil, fmt.Errorf("constructing OIDC provider: %w", err)
	}

	return &idTokenHandler{
		verifier:      provider.Verifier(&oidc.Config{ClientID: opts.ClientID}),
		provider:      provider,
		usernameClaim: claim(opts.UsernameClaim),
	}, nil
}

// parseUserInfo parses the user info from an oauth access token
func (o idTokenHandler) parseUserInfo(ctx context.Context, oauthToken *oauth2.Token) (UserInfo, error) {
	// Extract the ID Token from OAuth2 token.
	rawIDToken, ok := oauthToken.Extra("id_token").(string)
	if !ok {
		return UserInfo{}, errors.New("id_token missing")
	}

	// Parse and verify ID Token payload.
	idToken, err := o.verifier.Verify(ctx, rawIDToken)
	if err != nil {
		return UserInfo{}, err
	}

	// Extract claims from id token
	var claims struct {
		Name      user.Username `json:"name"`
		Sub       user.Username `json:"sub"`
		Email     user.Username `json:"email"`
		AvatarURL string        `json:"picture"`
	}
	if err := idToken.Claims(&claims); err != nil {
		return UserInfo{}, err
	}

	// Extract user info from claims
	var userInfo UserInfo
	switch o.usernameClaim {
	case NameClaim:
		userInfo.Username = claims.Name
	case SubClaim:
		userInfo.Username = claims.Sub
	case EmailClaim:
		userInfo.Username = claims.Email
	}
	if claims.AvatarURL != "" {
		userInfo.AvatarURL = &claims.AvatarURL
	}
	return userInfo, nil
}
