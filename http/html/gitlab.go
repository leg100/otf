package html

import (
	"context"
	"fmt"
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

// GetUser retrieves a user from gitlab. The user's organizations map to gitlab
// groups and the user's teams map to their access level on the groups, e.g. a
// user with maintainer access level on group acme maps to a user in the
// maintainer team in the acme organization.
func (g *gitlabProvider) GetUser(ctx context.Context) (*otf.User, error) {
	guser, _, err := g.client.Users.CurrentUser()
	if err != nil {
		return nil, err
	}

	groups, _, err := g.client.Groups.ListGroups(&gitlab.ListGroupsOptions{
		TopLevelOnly: otf.Bool(true),
	})
	if err != nil {
		return nil, err
	}
	var orgs []*otf.Organization
	var teams []*otf.Team
	for _, group := range groups {
		// Create org for each top-level group
		org, err := otf.NewOrganization(otf.OrganizationCreateOptions{
			Name: otf.String(group.Path),
		})
		if err != nil {
			return nil, err
		}
		orgs = append(orgs, org)

		// Get group membership info
		membership, _, err := g.client.GroupMembers.GetGroupMember(group.ID, guser.ID)
		if err != nil {
			return nil, err
		}
		var teamName string
		switch membership.AccessLevel {
		case gitlab.OwnerPermissions:
			teamName = "owners"
		case gitlab.DeveloperPermissions:
			teamName = "developers"
		case gitlab.MaintainerPermissions:
			teamName = "maintainers"
		case gitlab.ReporterPermissions:
			teamName = "reporters"
		case gitlab.GuestPermissions:
			teamName = "guests"
		default:
			// TODO: skip unknown access levels without error
			return nil, fmt.Errorf("unknown gitlab access level: %d", membership.AccessLevel)
		}
		teams = append(teams, otf.NewTeam(teamName, org))
	}
	user := otf.NewUser(guser.Username, otf.WithOrganizationMemberships(orgs...), otf.WithTeamMemberships(teams...))
	return user, nil
}
