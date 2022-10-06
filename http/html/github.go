package html

import (
	"context"
	"crypto/tls"
	"net/http"
	"net/url"

	"github.com/google/go-github/v41/github"
	"github.com/spf13/pflag"
	"golang.org/x/oauth2"
	oauth2github "golang.org/x/oauth2/github"
)

const DefaultGithubHostname = "github.com"

// TODO: rename to githubClient
type githubProvider struct {
	client *github.Client
}

type GithubConfig struct {
	cloudConfig
}

func NewGithubConfigFromFlags(flags *pflag.FlagSet) *GithubConfig {
	cfg := &GithubConfig{
		cloudConfig: cloudConfig{
			OAuthCredentials: &OAuthCredentials{prefix: "github"},
			cloudName:        "github",
			endpoint:         oauth2github.Endpoint,
			scopes:           []string{"user:email", "read:org"},
		},
	}

	flags.StringVar(&cfg.hostname, "github-hostname", DefaultGithubHostname, "Github hostname")
	flags.BoolVar(&cfg.skipTLSVerification, "github-skip-tls-verification", false, "Skip github TLS verification")
	cfg.OAuthCredentials.AddFlags(flags)

	return cfg
}

func (cfg *GithubConfig) NewCloud() (Cloud, error) {
	return &GithubCloud{GithubConfig: cfg}, nil
}

type GithubCloud struct {
	*GithubConfig
}

func (g *GithubCloud) NewDirectoryClient(ctx context.Context, opts DirectoryClientOptions) (DirectoryClient, error) {
	var err error
	var client *github.Client

	// Optionally skip TLS verification of github API
	if g.skipTLSVerification {
		ctx = context.WithValue(ctx, oauth2.HTTPClient, &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		})
	}

	httpClient := opts.Config.Client(ctx, opts.Token)

	if g.hostname != DefaultGithubHostname {
		client, err = github.NewEnterpriseClient(enterpriseBaseURL(g.hostname), enterpriseUploadURL(g.hostname), httpClient)
		if err != nil {
			return nil, err
		}
	} else {
		client = github.NewClient(httpClient)
	}
	return &githubProvider{client: client}, nil
}

func (g *githubProvider) GetUser(ctx context.Context) (string, error) {
	user, _, err := g.client.Users.Get(ctx, "")
	if err != nil {
		return "", err
	}
	return user.GetLogin(), nil
}

func (g *githubProvider) ListOrganizations(ctx context.Context) ([]string, error) {
	orgs, _, err := g.client.Organizations.List(ctx, "", nil)
	if err != nil {
		return nil, err
	}
	names := []string{}
	for _, o := range orgs {
		names = append(names, o.GetLogin())
	}
	return names, nil
}

// Return a github enterprise URL from a hostname
func enterpriseBaseURL(hostname string) string   { return enterpriseURL(hostname, "/api/v3") }
func enterpriseUploadURL(hostname string) string { return enterpriseURL(hostname, "/api/uploads") }
func enterpriseURL(hostname, path string) string {
	return (&url.URL{Scheme: "https", Host: hostname, Path: path}).String()
}
