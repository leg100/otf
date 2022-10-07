package html

import (
	"context"
	"net/url"

	"github.com/leg100/otf"
	"github.com/spf13/pflag"
	"github.com/xanzy/go-gitlab"
	oauth2gitlab "golang.org/x/oauth2/gitlab"
)

const DefaultGitlabHostname = "gitlab.com"

type gitlabCloud struct {
	*GitlabConfig
}

// TODO: rename to gitlabClient
type gitlabProvider struct {
	client *gitlab.Client
}

type GitlabConfig struct {
	cloudConfig
}

func NewGitlabConfigFromFlags(flags *pflag.FlagSet) *GitlabConfig {
	cfg := &GitlabConfig{
		cloudConfig: cloudConfig{
			OAuthCredentials: &OAuthCredentials{prefix: "gitlab"},
			cloudName:        "gitlab",
			endpoint:         oauth2gitlab.Endpoint,
			scopes:           []string{"read_user", "read_api"},
		},
	}

	flags.StringVar(&cfg.hostname, "gitlab-hostname", DefaultGitlabHostname, "Gitlab hostname")
	cfg.OAuthCredentials.AddFlags(flags)

	return cfg
}

func (cfg *GitlabConfig) NewCloud() (Cloud, error) {
	return &gitlabCloud{GitlabConfig: cfg}, nil
}

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

func (g *gitlabProvider) ListTeams(ctx context.Context) ([]*otf.Team, error) {
	// First get top-level groups
	groups, _, err := g.client.Groups.ListGroups(&gitlab.ListGroupsOptions{
		TopLevelOnly: otf.Bool(true),
	})
	if err != nil {
		return nil, err
	}
	var teams []*otf.Team
	for _, group := range groups {
		org, err := otf.NewOrganization(otf.OrganizationCreateOptions{
			Name: otf.String(group.Path),
		})
		if err != nil {
			return nil, err
		}
		subs, _, err := g.client.Groups.ListSubGroups(group.ID, nil)
		if err != nil {
			return nil, err
		}
		for _, s := range subs {
			teams = append(teams, otf.NewTeam(s.FullPath, org))
		}
	}
	return teams, nil
}

func (g *gitlabProvider) ListOrganizations(ctx context.Context) ([]*otf.Organization, error) {
	groups, _, err := g.client.Groups.ListGroups(&gitlab.ListGroupsOptions{
		TopLevelOnly: otf.Bool(true),
	})
	if err != nil {
		return nil, err
	}
	var orgs []*otf.Organization
	for _, group := range groups {
		org, err := otf.NewOrganization(otf.OrganizationCreateOptions{
			Name: otf.String(group.Path),
		})
		if err != nil {
			return nil, err
		}
		orgs = append(orgs, org)
	}
	return orgs, nil
}
