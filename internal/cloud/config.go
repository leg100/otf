package cloud

import (
	"context"
	"crypto/tls"
	"net/http"

	"golang.org/x/oauth2"
)

// Config is configuration for a cloud provider
type Config struct {
	Name                string
	Hostname            string
	SkipTLSVerification bool

	Cloud
}

func (cfg Config) String() string { return cfg.Name }

func (cfg *Config) NewClient(ctx context.Context, creds Credentials) (Client, error) {
	return cfg.Cloud.NewClient(ctx, ClientOptions{
		Hostname:            cfg.Hostname,
		SkipTLSVerification: cfg.SkipTLSVerification,
		Credentials:         creds,
	})
}

func (cfg *Config) HTTPClient() *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: cfg.SkipTLSVerification,
			},
		},
	}
}

// Credentials are credentials for a cloud client
type Credentials struct {
	// tokens are mutually-exclusive - only one can be specified
	OAuthToken    *oauth2.Token
	PersonalToken *string
	GithubApp     *GithubApp
}

// CloudOAuthConfig is the configuration for a cloud provider and its OAuth
// configuration.
type CloudOAuthConfig struct {
	Config
	OAuthConfig *oauth2.Config
}

// OIDCConfig is the configuration for a generic oidc provider.
type OIDCConfig struct {
	// Name is the user-friendly identifier of the oidc endpoint.
	Name string
	// IssuerURL is the issuer url for the oidc provider.
	IssuerURL string
	// RedirectURL is the redirect url for the oidc provider.
	RedirectURL string
	// ClientID is the client id for the oidc provider.
	ClientID string
	// ClientSecret is the client secret for the oidc provider.
	ClientSecret string
	// Skip TLS Verification when communicating with issuer
	SkipTLSVerification bool
}
