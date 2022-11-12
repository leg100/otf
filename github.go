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
	"github.com/spf13/pflag"
	"golang.org/x/oauth2"
	oauth2github "golang.org/x/oauth2/github"
)

const DefaultGithubHostname = "github.com"

func defaultGithubConfig() *GithubConfig {
	return &GithubConfig{
		cloudConfig: cloudConfig{
			OAuthCredentials: &OAuthCredentials{prefix: "github"},
			cloudName:        "github",
			endpoint:         oauth2github.Endpoint,
			scopes:           []string{"user:email", "read:org"},
			hostname:         DefaultGithubHostname,
		},
	}
}

// TODO: rename to githubClient
type githubProvider struct {
	client *github.Client
}

type GithubConfig struct {
	cloudConfig
}

func NewGithubConfigFromFlags(flags *pflag.FlagSet) *GithubConfig {
	cfg := defaultGithubConfig()

	flags.StringVar(&cfg.hostname, "github-hostname", cfg.hostname, "Github hostname")
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

func NewGithubCloud(opts *cloudConfigOptions) *GithubCloud {
	cloud := &GithubCloud{defaultGithubConfig()}
	cloud.override(opts)
	return cloud
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

	// Github's oauth access token never expires
	var src oauth2.TokenSource
	if opts.OAuthToken != nil {
		src = oauth2.StaticTokenSource(opts.OAuthToken)
	} else if opts.PersonalToken != nil {
		src = oauth2.StaticTokenSource(&oauth2.Token{AccessToken: *opts.PersonalToken})
	} else {
		return nil, fmt.Errorf("no credentials provided")
	}

	httpClient := oauth2.NewClient(ctx, src)

	if g.hostname != DefaultGithubHostname {
		client, err = NewGithubEnterpriseClient(g.hostname, httpClient)
		if err != nil {
			return nil, err
		}
	} else {
		client = github.NewClient(httpClient)
	}
	return &githubProvider{client: client}, nil
}

func (g *githubProvider) GetUser(ctx context.Context) (*User, error) {
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

func (g *githubProvider) GetRepository(ctx context.Context, identifier string) (*Repo, error) {
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

func (g *githubProvider) ListRepositories(ctx context.Context, opts ListOptions) (*RepoList, error) {
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

func (g *githubProvider) GetRepoTarball(ctx context.Context, repo *VCSRepo) ([]byte, error) {
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

func NewGithubEnterpriseClient(hostname string, httpClient *http.Client) (*github.Client, error) {
	return github.NewEnterpriseClient(
		"https://"+hostname,
		"https://"+hostname,
		httpClient)
}
