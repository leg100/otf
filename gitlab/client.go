package gitlab

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/leg100/otf"
	"github.com/leg100/otf/cloud"
	"github.com/xanzy/go-gitlab"
)

type Client struct {
	client *gitlab.Client
}

func NewClient(ctx context.Context, cfg cloud.ClientOptions) (*Client, error) {
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

func (g *Client) GetRepository(ctx context.Context, identifier string) (string, error) {
	proj, _, err := g.client.Projects.GetProject(identifier, nil)
	if err != nil {
		return "", err
	}

	return proj.PathWithNamespace, nil
}

func (g *Client) ListRepositories(ctx context.Context, lopts cloud.ListRepositoriesOptions) ([]string, error) {
	opts := &gitlab.ListProjectsOptions{
		ListOptions: gitlab.ListOptions{
			PerPage: lopts.PageSize,
		},
		// limit results to those repos the authenticated user is a member of,
		// otherwise we'll get *all* accessible repos, public and private.
		Membership: otf.Bool(true),
	}
	projects, _, err := g.client.Projects.ListProjects(opts, nil)
	if err != nil {
		return nil, err
	}

	var repos []string
	for _, proj := range projects {
		repos = append(repos, proj.PathWithNamespace)
	}
	return repos, nil
}

func (g *Client) ListTags(ctx context.Context, opts cloud.ListTagsOptions) ([]string, error) {
	results, _, err := g.client.Tags.ListTags(opts.Repo, &gitlab.ListTagsOptions{
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

func (g *Client) GetRepoTarball(ctx context.Context, opts cloud.GetRepoTarballOptions) ([]byte, error) {
	owner, name, found := strings.Cut(opts.Repo, "/")
	if !found {
		return nil, fmt.Errorf("malformed identifier: %s", opts.Repo)
	}

	tarball, _, err := g.client.Repositories.Archive(opts.Repo, &gitlab.ArchiveOptions{
		Format: otf.String("tar.gz"),
		SHA:    opts.Ref,
	})
	if err != nil {
		return nil, err
	}

	// Gitlab tarball contents are contained within a top-level directory
	// formatted <repo>-<ref>-<sha>. We want the tarball without this directory,
	// so we re-tar the contents without the top-level directory.
	untarpath, err := os.MkdirTemp("", fmt.Sprintf("gitlab-%s-%s-*", owner, name))
	if err != nil {
		return nil, err
	}
	if err := otf.Unpack(bytes.NewReader(tarball), untarpath); err != nil {
		return nil, err
	}
	contents, err := os.ReadDir(untarpath)
	if err != nil {
		return nil, err
	}
	if len(contents) != 1 {
		return nil, fmt.Errorf("expected only one top-level directory; instead got %s", contents)
	}
	return otf.Pack(path.Join(untarpath, contents[0].Name()))
}

func (g *Client) CreateWebhook(ctx context.Context, opts cloud.CreateWebhookOptions) (string, error) {
	addOpts := &gitlab.AddProjectHookOptions{
		EnableSSLVerification: otf.Bool(true),
		PushEvents:            otf.Bool(true),
		Token:                 otf.String(opts.Secret),
		URL:                   otf.String(opts.Endpoint),
	}
	for _, event := range opts.Events {
		switch event {
		case cloud.VCSPushEventType:
			addOpts.PushEvents = otf.Bool(true)
		case cloud.VCSPullEventType:
			addOpts.MergeRequestsEvents = otf.Bool(true)
		}
	}

	hook, _, err := g.client.Projects.AddProjectHook(opts.Repo, addOpts)
	if err != nil {
		return "", err
	}
	return strconv.Itoa(hook.ID), nil
}

func (g *Client) UpdateWebhook(ctx context.Context, opts cloud.UpdateWebhookOptions) error {
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
		case cloud.VCSPushEventType:
			editOpts.PushEvents = otf.Bool(true)
		case cloud.VCSPullEventType:
			editOpts.MergeRequestsEvents = otf.Bool(true)
		}
	}

	_, _, err = g.client.Projects.EditProjectHook(opts.Repo, id, editOpts)
	if err != nil {
		return err
	}
	return nil
}

func (g *Client) GetWebhook(ctx context.Context, opts cloud.GetWebhookOptions) (cloud.Webhook, error) {
	id, err := strconv.Atoi(opts.ID)
	if err != nil {
		return cloud.Webhook{}, err
	}

	hook, resp, err := g.client.Projects.GetProjectHook(opts.Repo, id)
	if err != nil {
		if resp.StatusCode == http.StatusNotFound {
			return cloud.Webhook{}, otf.ErrResourceNotFound
		}
		return cloud.Webhook{}, err
	}

	var events []cloud.VCSEventType
	if hook.PushEvents {
		events = append(events, cloud.VCSPushEventType)
	}
	if hook.MergeRequestsEvents {
		events = append(events, cloud.VCSPullEventType)
	}

	return cloud.Webhook{
		ID:       strconv.Itoa(id),
		Repo:     opts.Repo,
		Events:   events,
		Endpoint: hook.URL,
	}, nil
}

func (g *Client) DeleteWebhook(ctx context.Context, opts cloud.DeleteWebhookOptions) error {
	id, err := strconv.Atoi(opts.ID)
	if err != nil {
		return err
	}

	_, err = g.client.Projects.DeleteProjectHook(opts.Repo, id)
	return err
}

func (g *Client) SetStatus(ctx context.Context, opts cloud.SetStatusOptions) error {
	return nil
}
