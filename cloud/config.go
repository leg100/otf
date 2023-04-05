package cloud

import (
	"context"
	"crypto/tls"
	"errors"
	"net/http"
	"strings"

	"golang.org/x/oauth2"
)

var ErrInvalidOrganizationPolicy = errors.New("must provide a valid policy consisting of \"organization name,owner role, admin role, write role, plan role, read role\"")

// Config is configuration for a cloud provider
type Config struct {
	Name                string
	Hostname            string
	SkipTLSVerification bool

	Cloud
}

func (cfg Config) String() string {
	return string(cfg.Name)
}

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
	// tokens are mutually-exclusive - at least one must be specified
	OAuthToken    *oauth2.Token
	PersonalToken *string
}

// CloudOAuthConfig is the configuration for a cloud provider and its OAuth
// configuration.
type CloudOAuthConfig struct {
	Config
	OAuthConfig *oauth2.Config
}

// OIDCConfig is the configuration for a generic oidc provider.
type OIDCConfig struct {
	// Name is the user friendly identifier of the oidc endpoint.
	Name string
	// IssuerURL is the issuer url for the oidc provider.
	IssuerURL string
	// RedirectURL is the redirect url for the oidc provider.
	RedirectURL string
	// ClientID is the client id for the oidc provider.
	ClientID string
	// ClientSecret is the client secret for the oidc provider.
	ClientSecret string
	// Scopes is a list of optional scopes to pass to the oidc provider.
	Scopes []string
	// OrganizationPolicies is a comma separated list containing the organization name, then the groups that should map
	// to the following roles: owner, admin, write, plan, read.
	OrganizationPolicies []string
}

// OIDCOrganizationPolicy is a struct containing the policies for an oidc organization.
type OIDCOrganizationPolicy struct {
	// Organization is the name of the organization.
	Organization string
	// OwnerRole is the name of the team that owns the organization.
	OwnerRole string
	// AdminRole is the name of the team that has admin privileges on the organization.
	AdminRole string
	// WriteRole is the name of the team that has write permissions on the organization.
	WriteRole string
	// PlanRole is the name of the team that has plan permissions on the organization.
	PlanRole string
	// ReadRole is the name of the team that has read permissions on the organization.
	ReadRole string
}

// GetOrganizationPolicies parses the comma separated OrganizationPolicies and turns it into OIDCOrganizationPolicy structs.
func (o OIDCConfig) GetOrganizationPolicies() ([]OIDCOrganizationPolicy, error) {
	var orgs []OIDCOrganizationPolicy
	for _, org := range o.OrganizationPolicies {
		tokens := strings.Split(org, ",")

		if len(tokens) != 6 {
			return nil, ErrInvalidOrganizationPolicy
		}

		orgs = append(orgs, OIDCOrganizationPolicy{
			Organization: tokens[0],
			OwnerRole:    tokens[1],
			AdminRole:    tokens[2],
			WriteRole:    tokens[3],
			PlanRole:     tokens[4],
			ReadRole:     tokens[5],
		})
	}

	return orgs, nil
}
