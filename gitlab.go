package otf

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/xanzy/go-gitlab"
	"golang.org/x/oauth2"
	oauth2gitlab "golang.org/x/oauth2/gitlab"
)

func GitlabDefaults() CloudConfig {
	return CloudConfig{
		Name:     "gitlab",
		Hostname: "gitlab.com",
		Cloud:    &GitlabCloud{},
	}
}

func GitlabOAuthDefaults() *oauth2.Config {
	return &oauth2.Config{
		Endpoint: oauth2gitlab.Endpoint,
		Scopes:   []string{"read_user", "read_api"},
	}
}

type GitlabCloud struct{}

func (g *GitlabCloud) NewClient(ctx context.Context, opts CloudClientOptions) (CloudClient, error) {
	return NewGitlabClient(ctx, opts)
}

func (GitlabCloud) HandleEvent(w http.ResponseWriter, r *http.Request, opts HandleEventOptions) *VCSEvent {
	return nil
}

type GitlabClient struct {
	client *gitlab.Client
}

func NewGitlabClient(ctx context.Context, cfg CloudClientOptions) (*GitlabClient, error) {
	var err error
	var client *gitlab.Client

	baseURL := (&url.URL{Scheme: "https", Host: cfg.Hostname}).String()

	// TODO: apply skipTLS option

	if cfg.OAuthToken != nil {
		client, err = gitlab.NewOAuthClient(cfg.OAuthToken.AccessToken, gitlab.WithBaseURL(baseURL))
		if err != nil {
			return nil, err
		}
	} else if cfg.PersonalToken != nil {
		client, err = gitlab.NewClient(*cfg.PersonalToken, gitlab.WithBaseURL(baseURL))
		if err != nil {
			return nil, err
		}
	} else {
		return nil, fmt.Errorf("no credentials provided")
	}

	return &GitlabClient{client: client}, nil
}

// GetUser retrieves a user from gitlab. The user's organizations map to gitlab
// groups and the user's teams map to their access level on the groups, e.g. a
// user with maintainer access level on group acme maps to a user in the
// maintainer team in the acme organization.
func (g *GitlabClient) GetUser(ctx context.Context) (*User, error) {
	guser, _, err := g.client.Users.CurrentUser()
	if err != nil {
		return nil, err
	}

	groups, _, err := g.client.Groups.ListGroups(&gitlab.ListGroupsOptions{
		TopLevelOnly: Bool(true),
	})
	if err != nil {
		return nil, err
	}
	var orgs []*Organization
	var teams []*Team
	for _, group := range groups {
		// Create org for each top-level group
		org, err := NewOrganization(OrganizationCreateOptions{
			Name: String(group.Path),
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
		teams = append(teams, NewTeam(teamName, org))
	}
	user := NewUser(guser.Username, WithOrganizationMemberships(orgs...), WithTeamMemberships(teams...))
	return user, nil
}

func (g *GitlabClient) GetRepository(ctx context.Context, identifier string) (*Repo, error) {
	proj, _, err := g.client.Projects.GetProject(identifier, nil)
	if err != nil {
		return nil, err
	}

	return &Repo{
		Identifier: proj.PathWithNamespace,
		HTTPURL:    proj.WebURL,
		Branch:     proj.DefaultBranch,
	}, nil
}

func (g *GitlabClient) ListRepositories(ctx context.Context, lopts ListOptions) (*RepoList, error) {
	opts := &gitlab.ListProjectsOptions{
		ListOptions: gitlab.ListOptions{
			Page:    lopts.PageNumber,
			PerPage: lopts.PageSize,
		},
	}
	projects, resp, err := g.client.Projects.ListProjects(opts, nil)
	if err != nil {
		return nil, err
	}

	// convert to common repo type before returning
	var items []*Repo
	for _, proj := range projects {
		items = append(items, &Repo{
			Identifier: proj.PathWithNamespace,
			HTTPURL:    proj.WebURL,
			Branch:     proj.DefaultBranch,
		})
	}
	return &RepoList{
		Items:      items,
		Pagination: NewPagination(lopts, resp.TotalItems),
	}, nil
}

func (g *GitlabClient) GetRepoTarball(ctx context.Context, opts GetRepoTarballOptions) ([]byte, error) {
	tarball, _, err := g.client.Repositories.Archive(opts.Identifier, &gitlab.ArchiveOptions{
		Format: String("tar.gz"),
		SHA:    String(opts.Ref),
	})
	if err != nil {
		return nil, err
	}

	return tarball, nil
}

func (g *GitlabClient) CreateWebhook(ctx context.Context, opts CreateWebhookOptions) (string, error) {
	addOpts := &gitlab.AddProjectHookOptions{
		EnableSSLVerification: Bool(true),
		PushEvents:            Bool(true),
		Token:                 String(opts.Secret),
		URL:                   String(opts.Endpoint),
	}
	for _, event := range opts.Events {
		switch event {
		case VCSPushEventType:
			addOpts.PushEvents = Bool(true)
		case VCSPullEventType:
			addOpts.MergeRequestsEvents = Bool(true)
		}
	}

	hook, _, err := g.client.Projects.AddProjectHook(opts.Identifier, addOpts)
	if err != nil {
		return "", err
	}
	return strconv.Itoa(hook.ID), nil
}

func (g *GitlabClient) UpdateWebhook(ctx context.Context, opts UpdateWebhookOptions) error {
	id, err := strconv.Atoi(opts.ID)
	if err != nil {
		return err
	}

	editOpts := &gitlab.EditProjectHookOptions{
		EnableSSLVerification: Bool(true),
		Token:                 String(opts.Secret),
		URL:                   String(opts.Endpoint),
	}
	for _, event := range opts.Events {
		switch event {
		case VCSPushEventType:
			editOpts.PushEvents = Bool(true)
		case VCSPullEventType:
			editOpts.MergeRequestsEvents = Bool(true)
		}
	}

	_, _, err = g.client.Projects.EditProjectHook(opts.Identifier, id, editOpts)
	if err != nil {
		return err
	}
	return nil
}

func (g *GitlabClient) GetWebhook(ctx context.Context, opts GetWebhookOptions) (*VCSWebhook, error) {
	id, err := strconv.Atoi(opts.ID)
	if err != nil {
		return nil, err
	}

	hook, resp, err := g.client.Projects.GetProjectHook(opts.Identifier, id)
	if err != nil {
		if resp.StatusCode == http.StatusNotFound {
			return nil, ErrResourceNotFound
		}
		return nil, err
	}

	var events []VCSEventType
	if hook.PushEvents {
		events = append(events, VCSPushEventType)
	}
	if hook.MergeRequestsEvents {
		events = append(events, VCSPullEventType)
	}

	return &VCSWebhook{
		ID:         strconv.Itoa(id),
		Identifier: opts.Identifier,
		Events:     events,
		Endpoint:   hook.URL,
	}, nil
}

func (g *GitlabClient) DeleteWebhook(ctx context.Context, opts DeleteWebhookOptions) error {
	id, err := strconv.Atoi(opts.ID)
	if err != nil {
		return err
	}

	_, err = g.client.Projects.DeleteProjectHook(opts.Identifier, id)
	return err
}

// TODO: implement
func (g *GitlabClient) SetStatus(ctx context.Context, opts SetStatusOptions) error {
	return nil
}
