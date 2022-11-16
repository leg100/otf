package otf

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/google/go-github/v41/github"
	"golang.org/x/oauth2"
	oauth2github "golang.org/x/oauth2/github"
)

const (
	GithubCloudName       CloudName = "github"
	DefaultGithubHostname string    = "github.com"
)

func GithubDefaultConfig() *CloudConfig {
	return &CloudConfig{
		Name:     GithubCloudName,
		Hostname: "github.com",
		Endpoint: oauth2github.Endpoint,
		Scopes:   []string{"user:email", "read:org"},
		Cloud:    GithubCloud{},
	}
}

type GithubCloud struct{}

func (GithubCloud) NewClient(ctx context.Context, cfg ClientConfig) (CloudClient, error) {
	return NewGithubClient(ctx, cfg)
}

type GithubClient struct {
	client *github.Client
}

func NewGithubClient(ctx context.Context, cfg ClientConfig) (*GithubClient, error) {
	var err error
	var client *github.Client

	// Optionally skip TLS verification of github API
	if cfg.SkipTLSVerification {
		ctx = context.WithValue(ctx, oauth2.HTTPClient, &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		})
	}

	// Github's oauth access token never expires
	var src oauth2.TokenSource
	if cfg.OAuthToken != nil {
		src = oauth2.StaticTokenSource(cfg.OAuthToken)
	} else if cfg.PersonalToken != nil {
		src = oauth2.StaticTokenSource(&oauth2.Token{AccessToken: *cfg.PersonalToken})
	} else {
		return nil, fmt.Errorf("no credentials provided")
	}

	httpClient := oauth2.NewClient(ctx, src)

	if cfg.Hostname != DefaultGithubHostname {
		client, err = NewGithubEnterpriseClient(cfg.Hostname, httpClient)
		if err != nil {
			return nil, err
		}
	} else {
		client = github.NewClient(httpClient)
	}
	return &GithubClient{client: client}, nil
}

func NewGithubEnterpriseClient(hostname string, httpClient *http.Client) (*github.Client, error) {
	return github.NewEnterpriseClient(
		"https://"+hostname,
		"https://"+hostname,
		httpClient)
}

func (g *GithubClient) GetUser(ctx context.Context) (*User, error) {
	guser, _, err := g.client.Users.Get(ctx, "")
	if err != nil {
		return nil, err
	}

	var orgs []*Organization
	var teams []*Team

	gorgs, _, err := g.client.Organizations.List(ctx, "", nil)
	if err != nil {
		return nil, err
	}
	for _, gorg := range gorgs {
		org, err := NewOrganization(OrganizationCreateOptions{
			Name: String(gorg.GetLogin()),
		})
		if err != nil {
			return nil, err
		}
		orgs = append(orgs, org)

		// Determine if they are an admin; if so, add them to the owners team.
		membership, _, err := g.client.Organizations.GetOrgMembership(ctx, "", org.Name())
		if err != nil {
			return nil, err
		}
		if membership.GetRole() == "admin" {
			teams = append(teams, NewTeam("owners", org))
		}
	}

	gteams, _, err := g.client.Teams.ListUserTeams(ctx, nil)
	if err != nil {
		return nil, err
	}
	for _, gteam := range gteams {
		org, err := NewOrganization(OrganizationCreateOptions{
			Name: String(gteam.GetOrganization().GetLogin()),
		})
		if err != nil {
			return nil, err
		}
		teams = append(teams, NewTeam(gteam.GetName(), org))
	}

	user := NewUser(guser.GetLogin(), WithOrganizationMemberships(orgs...), WithTeamMemberships(teams...))
	return user, nil
}

func (g *GithubClient) GetRepository(ctx context.Context, identifier string) (*Repo, error) {
	owner, name, found := strings.Cut(identifier, "/")
	if !found {
		return nil, fmt.Errorf("malformed identifier: %s", identifier)
	}
	repo, _, err := g.client.Repositories.Get(ctx, owner, name)
	if err != nil {
		return nil, err
	}

	return &Repo{
		Identifier: repo.GetFullName(),
		HTTPURL:    repo.GetURL(),
		Branch:     repo.GetDefaultBranch(),
	}, nil
}

func (g *GithubClient) ListRepositories(ctx context.Context, opts ListOptions) (*RepoList, error) {
	repos, resp, err := g.client.Repositories.List(ctx, "", &github.RepositoryListOptions{
		ListOptions: github.ListOptions{
			Page:    opts.SanitizedPageNumber(),
			PerPage: opts.SanitizedPageSize(),
		},
	})
	if err != nil {
		return nil, err
	}

	// convert to common repo type before returning
	var items []*Repo
	for _, repo := range repos {
		items = append(items, &Repo{
			Identifier: repo.GetFullName(),
			HTTPURL:    repo.GetURL(),
			Branch:     repo.GetDefaultBranch(),
		})
	}

	return &RepoList{
		Items:      items,
		Pagination: NewPagination(opts, resp.LastPage*opts.SanitizedPageSize()),
	}, nil
}

func (g *GithubClient) GetRepoTarball(ctx context.Context, repo *VCSRepo) ([]byte, error) {
	opts := github.RepositoryContentGetOptions{
		Ref: repo.Branch,
	}
	link, _, err := g.client.Repositories.GetArchiveLink(ctx, repo.Owner(), repo.Repo(), github.Tarball, &opts, true)
	if err != nil {
		return nil, err
	}

	resp, err := g.client.Client().Get(link.String())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// github tarball contains a parent directory of the format
	// <owner>-<repo>-<commit>. We need a tarball without this parent directory,
	// so we untar it to a temp dir, then tar it up the contents of the parent
	// directory.
	untarpath, err := os.MkdirTemp("", fmt.Sprintf("github-%s-%s-*", repo.Owner(), repo.Repo()))
	if err != nil {
		return nil, err
	}
	if err := Unpack(resp.Body, untarpath); err != nil {
		return nil, err
	}
	contents, err := os.ReadDir(untarpath)
	if err != nil {
		return nil, err
	}
	if len(contents) != 1 {
		return nil, fmt.Errorf("malformed tarball archive")
	}
	parentDir := path.Join(untarpath, contents[0].Name())
	return Pack(parentDir)
}

// CreateWebhook creates a webhook on the github repository, subscribing the
// workspace to certain github events so that runs can be triggered.
func (g *GithubClient) CreateWebhook(ctx context.Context, opts CreateWebhookOptions) error {
	owner, name, found := strings.Cut(opts.Identifier, "/")
	if !found {
		return fmt.Errorf("malformed identifier: %s", opts.Identifier)
	}
	_, _, err := g.client.Repositories.CreateHook(ctx, owner, name, &github.Hook{
		Name: String("web"),
		// default is [push]
		Events: []string{"push"},
		Config: map[string]any{
			"url": opts.URL,
			// For now use global secret as key that github takes, stores, and uses to
			// generate HMAC hex digest signature value. This isn't ideal
			// because 1) github have it and we don't trust 'em 2) if otf admin
			// changes secret then they'll have to re-create all webhooks.
			// TODO: create a secret for each webhook, and persist. On every
			// incoming event, lookup secret corresponding to org/workspace and
			// verify signature. We would progress to using a cache to make this
			// lookup efficient.
			"secret": opts.Secret,
			// default is form
			"content_type": "form",
		},
		Active: Bool(true),
	})
	if err != nil {
		return err
	}
	return nil
}
