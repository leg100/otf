package html

import (
	"context"
	"crypto/tls"
	"net/http"

	"github.com/google/go-github/v41/github"
	"github.com/leg100/otf"
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
		client, err = NewGithubEnterpriseClient(g.hostname, httpClient)
		if err != nil {
			return nil, err
		}
	} else {
		client = github.NewClient(httpClient)
	}
	return &githubProvider{client: client}, nil
}

func (g *githubProvider) GetUser(ctx context.Context) (*otf.User, error) {
	guser, _, err := g.client.Users.Get(ctx, "")
	if err != nil {
		return nil, err
	}

	var orgs []*otf.Organization
	var teams []*otf.Team

	gorgs, _, err := g.client.Organizations.List(ctx, "", nil)
	if err != nil {
		return nil, err
	}
	for _, gorg := range gorgs {
		org, err := otf.NewOrganization(otf.OrganizationCreateOptions{
			Name: otf.String(gorg.GetLogin()),
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
			teams = append(teams, otf.NewTeam("owners", org))
		}
	}

	gteams, _, err := g.client.Teams.ListUserTeams(ctx, nil)
	if err != nil {
		return nil, err
	}
	for _, gteam := range gteams {
		org, err := otf.NewOrganization(otf.OrganizationCreateOptions{
			Name: otf.String(gteam.GetOrganization().GetLogin()),
		})
		if err != nil {
			return nil, err
		}
		teams = append(teams, otf.NewTeam(gteam.GetName(), org))
	}

	user := otf.NewUser(guser.GetLogin(), otf.WithOrganizationMemberships(orgs...), otf.WithTeamMemberships(teams...))
	return user, nil
}

func NewGithubEnterpriseClient(hostname string, httpClient *http.Client) (*github.Client, error) {
	return github.NewEnterpriseClient(
		"https://"+hostname,
		"https://"+hostname,
		httpClient)
}
