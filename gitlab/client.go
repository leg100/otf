package gitlab

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/leg100/otf"
	"github.com/leg100/otf/cloud"
	"github.com/xanzy/go-gitlab"
)

type Client struct {
	client *gitlab.Client
}

func NewClient(ctx context.Context, cfg otf.CloudClientOptions) (*Client, error) {
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

	return &Client{client: client}, nil
}

// GetUser retrieves a user from gitlab. The user's organizations map to gitlab
// groups and the user's teams map to their access level on the groups, e.g. a
// user with maintainer access level on group acme maps to a user in the
// maintainer team in the acme organization.
func (g *Client) GetUser(ctx context.Context) (*cloud.User, error) {
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

	user := cloud.User{Name: guser.Username}
	for _, group := range groups {
		// Create org for each top-level group
		user.Organizations = append(user.Organizations, group.Path)

		// Get group membership info
		membership, _, err := g.client.GroupMembers.GetGroupMember(group.ID, guser.ID)
		if err != nil {
			return nil, err
		}
		var team string
		switch membership.AccessLevel {
		case gitlab.OwnerPermissions:
			team = "owners"
		case gitlab.DeveloperPermissions:
			team = "developers"
		case gitlab.MaintainerPermissions:
			team = "maintainers"
		case gitlab.ReporterPermissions:
			team = "reporters"
		case gitlab.GuestPermissions:
			team = "guests"
		default:
			// TODO: skip unknown access levels without error
			return nil, fmt.Errorf("unknown gitlab access level: %d", membership.AccessLevel)
		}
		user.Teams = append(user.Teams, cloud.Team{
			Name:         team,
			Organization: group.Path,
		})
	}
	return &user, nil
}

func (g *Client) GetRepository(ctx context.Context, identifier string) (*otf.Repo, error) {
	proj, _, err := g.client.Projects.GetProject(identifier, nil)
	if err != nil {
		return nil, err
	}

	return &otf.Repo{
		Identifier: proj.PathWithNamespace,
		Branch:     proj.DefaultBranch,
	}, nil
}

func (g *Client) ListRepositories(ctx context.Context, lopts otf.ListOptions) (*otf.RepoList, error) {
	opts := &gitlab.ListProjectsOptions{
		ListOptions: gitlab.ListOptions{
			Page:    lopts.PageNumber,
			PerPage: lopts.PageSize,
		},
		// limit results to those repos the authenticated user is a member of,
		// otherwise we'll get *all* accessible repos, public and private.
		Membership: otf.Bool(true),
	}
	projects, resp, err := g.client.Projects.ListProjects(opts, nil)
	if err != nil {
		return nil, err
	}

	// convert to common repo type before returning
	var items []*otf.Repo
	for _, proj := range projects {
		items = append(items, &otf.Repo{
			Identifier: proj.PathWithNamespace,
			Branch:     proj.DefaultBranch,
		})
	}
	return &otf.RepoList{
		Items:      items,
		Pagination: otf.NewPagination(lopts, resp.TotalItems),
	}, nil
}

func (g *Client) ListTags(ctx context.Context, opts otf.ListTagsOptions) ([]string, error) {
	results, _, err := g.client.Tags.ListTags(opts.Identifier, &gitlab.ListTagsOptions{
		Search: otf.String("^" + opts.Prefix),
	})
	if err != nil {
		return nil, err
	}

	var tags []string
	for _, ref := range results {
		tags = append(tags, fmt.Sprintf("tags/%s", ref.Name))
	}
	return tags, nil
}

func (g *Client) GetRepoTarball(ctx context.Context, opts otf.GetRepoTarballOptions) ([]byte, error) {
	tarball, _, err := g.client.Repositories.Archive(opts.Identifier, &gitlab.ArchiveOptions{
		Format: otf.String("tar.gz"),
		SHA:    otf.String(opts.Ref),
	})
	if err != nil {
		return nil, err
	}

	return tarball, nil
}

func (g *Client) CreateWebhook(ctx context.Context, opts otf.CreateWebhookOptions) (string, error) {
	addOpts := &gitlab.AddProjectHookOptions{
		EnableSSLVerification: otf.Bool(true),
		PushEvents:            otf.Bool(true),
		Token:                 otf.String(opts.Secret),
		URL:                   otf.String(opts.Endpoint),
	}
	for _, event := range opts.Events {
		switch event {
		case otf.VCSPushEventType:
			addOpts.PushEvents = otf.Bool(true)
		case otf.VCSPullEventType:
			addOpts.MergeRequestsEvents = otf.Bool(true)
		}
	}

	hook, _, err := g.client.Projects.AddProjectHook(opts.Identifier, addOpts)
	if err != nil {
		return "", err
	}
	return strconv.Itoa(hook.ID), nil
}

func (g *Client) UpdateWebhook(ctx context.Context, opts otf.UpdateWebhookOptions) error {
	id, err := strconv.Atoi(opts.ID)
	if err != nil {
		return err
	}

	editOpts := &gitlab.EditProjectHookOptions{
		EnableSSLVerification: otf.Bool(true),
		Token:                 otf.String(opts.Secret),
		URL:                   otf.String(opts.Endpoint),
	}
	for _, event := range opts.Events {
		switch event {
		case otf.VCSPushEventType:
			editOpts.PushEvents = otf.Bool(true)
		case otf.VCSPullEventType:
			editOpts.MergeRequestsEvents = otf.Bool(true)
		}
	}

	_, _, err = g.client.Projects.EditProjectHook(opts.Identifier, id, editOpts)
	if err != nil {
		return err
	}
	return nil
}

func (g *Client) GetWebhook(ctx context.Context, opts otf.GetWebhookOptions) (*otf.VCSWebhook, error) {
	id, err := strconv.Atoi(opts.ID)
	if err != nil {
		return nil, err
	}

	hook, resp, err := g.client.Projects.GetProjectHook(opts.Identifier, id)
	if err != nil {
		if resp.StatusCode == http.StatusNotFound {
			return nil, otf.ErrResourceNotFound
		}
		return nil, err
	}

	var events []otf.VCSEventType
	if hook.PushEvents {
		events = append(events, otf.VCSPushEventType)
	}
	if hook.MergeRequestsEvents {
		events = append(events, otf.VCSPullEventType)
	}

	return &otf.VCSWebhook{
		ID:         strconv.Itoa(id),
		Identifier: opts.Identifier,
		Events:     events,
		Endpoint:   hook.URL,
	}, nil
}

func (g *Client) DeleteWebhook(ctx context.Context, opts otf.DeleteWebhookOptions) error {
	id, err := strconv.Atoi(opts.ID)
	if err != nil {
		return err
	}

	_, err = g.client.Projects.DeleteProjectHook(opts.Identifier, id)
	return err
}

func (g *Client) SetStatus(ctx context.Context, opts otf.SetStatusOptions) error {
	return nil
}
