package html

import (
	"context"
	"net/url"

	"github.com/leg100/otf"
	"github.com/spf13/pflag"
	"github.com/xanzy/go-gitlab"
	"golang.org/x/oauth2"
	oauth2gitlab "golang.org/x/oauth2/gitlab"
)

const DefaultGitlabHostname = "gitlab.com"

type gitlabCloud struct {
	*GitlabConfig
	endpoint oauth2.Endpoint
}

type gitlabProvider struct {
	client *gitlab.Client
}

type GitlabConfig struct {
	*OAuthCredentials
	hostname string
}

func NewGitlabConfigFromFlags(flags *pflag.FlagSet) *GitlabConfig {
	cfg := &GitlabConfig{
		OAuthCredentials: &OAuthCredentials{prefix: "gitlab"},
	}

	flags.StringVar(&cfg.hostname, "gitlab-hostname", DefaultGitlabHostname, "Gitlab hostname")
	cfg.OAuthCredentials.AddFlags(flags)

	return cfg
}

func (cfg *GitlabConfig) NewCloud() (Cloud, error) {
	endpoint, err := updateEndpoint(oauth2gitlab.Endpoint, cfg.hostname)
	if err != nil {
		return nil, err
	}
	return &gitlabCloud{
		endpoint:     endpoint,
		GitlabConfig: cfg,
	}, nil
}

func (g *gitlabCloud) CloudName() string { return "gitlab" }

func (g *gitlabCloud) Scopes() []string {
	return []string{"read_user", "profile"}
}

func (g *gitlabCloud) Endpoint() oauth2.Endpoint { return g.endpoint }

func (g *gitlabCloud) NewDirectoryClient(ctx context.Context, opts DirectoryClientOptions) (DirectoryClient, error) {
	var err error
	var client *gitlab.Client

	baseURL := (&url.URL{Scheme: "https", Host: g.hostname}).String()

	client, err = gitlab.NewOAuthClient(opts.Token.AccessToken, gitlab.WithBaseURL(baseURL))
	if err != nil {
		return nil, err
	}

	return &gitlabProvider{client: client}, nil
}

func (g *gitlabProvider) GetUser(ctx context.Context) (string, error) {
	user, _, err := g.client.Users.CurrentUser()
	if err != nil {
		return "", err
	}
	return user.Username, nil
}

func (g *gitlabProvider) ListOrganizations(ctx context.Context) ([]string, error) {
	groups, _, err := g.client.Groups.ListGroups(&gitlab.ListGroupsOptions{
		TopLevelOnly: otf.Bool(true),
	})
	if err != nil {
		return nil, err
	}
	names := []string{}
	for _, o := range groups {
		names = append(names, o.Name)
	}
	return names, nil
}
