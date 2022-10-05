package html

import (
	"context"
	"net/url"

	"github.com/google/go-github/v41/github"
	"github.com/spf13/pflag"
	"golang.org/x/oauth2"
	oauth2github "golang.org/x/oauth2/github"
)

const DefaultGithubHostname = "github.com"

type githubProvider struct {
	client *github.Client
}

type GithubConfig struct {
	*OAuthCredentials
	hostname string
}

func NewGithubConfigFromFlags(flags *pflag.FlagSet) *GithubConfig {
	cfg := &GithubConfig{
		OAuthCredentials: &OAuthCredentials{prefix: "github"},
	}

	flags.StringVar(&cfg.hostname, "github-hostname", DefaultGithubHostname, "Github hostname")
	cfg.OAuthCredentials.AddFlags(flags)

	return cfg
}

func (cfg *GithubConfig) NewCloud() (Cloud, error) {
	endpoint, err := updateEndpoint(oauth2github.Endpoint, cfg.hostname)
	if err != nil {
		return nil, err
	}
	return &GithubCloud{
		endpoint:     endpoint,
		GithubConfig: cfg,
	}, nil
}

type GithubCloud struct {
	*GithubConfig
	endpoint oauth2.Endpoint
}

func (g *GithubCloud) CloudName() string { return "github" }

func (g *GithubCloud) Scopes() []string {
	return []string{"user:email", "read:org"}
}

func (g *GithubCloud) Endpoint() oauth2.Endpoint { return g.endpoint }

func (g *GithubCloud) NewDirectoryClient(ctx context.Context, opts DirectoryClientOptions) (DirectoryClient, error) {
	var err error
	var client *github.Client

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
