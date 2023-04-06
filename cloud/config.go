package cloud

import (
	"bufio"
	"context"
	"crypto/tls"
	"errors"
	"net/http"
	"os"
	"strings"

	"golang.org/x/oauth2"
)

var ErrInvalidOrganizationPolicy = errors.New("must provide a valid policy consisting of \"organization, group, team:team-name\"")

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
	// Scopes is a list of optional scopes to pass to the oidc provider.
	Scopes []string
	// OrganizationPolicies is a comma separated list containing the organization name, then the groups that should map
	// to the following roles: owner, admin, write, plan, read.
	OrganizationPolicies []string
	// OrganizationPoliciesFile
	OrganizationPoliciesFile string
}

type OIDCOrganizationPolicy struct {
	Organization string
	Group        string
	Team         string
}

const (
	oidcPolicyGroup      = "g"
	oidcPolicyTeamPrefix = "team:"
)

// GetOrganizationPolicies parses the comma separated OrganizationPolicies and turns it into OIDCOrganizationPolicy structs.
func (o OIDCConfig) GetOrganizationPolicies() ([]OIDCOrganizationPolicy, error) {
	policies := o.OrganizationPolicies

	if o.OrganizationPoliciesFile != "" {
		file, err := os.Open(o.OrganizationPoliciesFile)
		if err != nil {
			return nil, err
		}

		fileScanner := bufio.NewScanner(file)
		fileScanner.Split(bufio.ScanLines)

		for fileScanner.Scan() {
			line := fileScanner.Text()
			if strings.HasPrefix(line, "#") {
				continue
			}

			policies = append(policies, line)
		}
	}

	var orgs []OIDCOrganizationPolicy
	for _, org := range policies {
		tokens := strings.Split(org, ",")

		if len(tokens) == 0 {
			return nil, ErrInvalidOrganizationPolicy
		}

		policyType := strings.Trim(tokens[0], " ")
		switch policyType {
		case oidcPolicyGroup:
			if len(tokens) != 4 {
				return nil, ErrInvalidOrganizationPolicy
			}

			organization := strings.Trim(tokens[1], " ")
			group := strings.Trim(tokens[2], " ")
			team := strings.Trim(tokens[3], " ")

			if !strings.HasPrefix(team, oidcPolicyTeamPrefix) {
				return nil, ErrInvalidOrganizationPolicy
			}

			teamToken := strings.TrimPrefix(team, oidcPolicyTeamPrefix)

			orgs = append(orgs, OIDCOrganizationPolicy{
				Organization: organization,
				Group:        group,
				Team:         teamToken,
			})
		default:
			return nil, ErrInvalidOrganizationPolicy
		}
	}

	return orgs, nil
}
