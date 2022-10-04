package gitlab

import (
	"context"
	"net/url"

	"github.com/leg100/otf"
	"github.com/xanzy/go-gitlab"
	"golang.org/x/oauth2"
)

type gitlabCloud struct{}

type gitlabProvider struct {
	client *gitlab.Client
}

func (g *gitlabCloud) Name() string            { return "gitlab" }
func (g *gitlabCloud) DefaultHostname() string { return "gitlab.com" }

func (g *gitlabCloud) Scopes() []string {
	return []string{"read_user", "profile"}
}

func (g *gitlabCloud) Endpoint(hostname string) (oauth2.Endpoint, error) {
	return updateEndpoint(oauth2gitlab.Endpoint, hostname)
}

func (g *gitlabCloud) NewDirectoryClient(ctx context.Context, opts DirectoryClientOptions) (DirectoryClient, error) {
	var err error
	var client *gitlab.Client

	baseURL := (&url.URL{Scheme: "https", Host: opts.Hostname}).String()

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
