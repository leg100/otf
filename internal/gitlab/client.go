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

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/cloud"
	"github.com/xanzy/go-gitlab"
)

type Client struct {
	client *gitlab.Client
}

func NewClient(ctx context.Context, cfg cloud.ClientOptions) (*Client, error) {
	var (
		client *gitlab.Client
		err    error
	)

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

func (g *Client) GetCurrentUser(ctx context.Context) (cloud.User, error) {
	guser, _, err := g.client.Users.CurrentUser()
	if err != nil {
		return cloud.User{}, err
	}
	return cloud.User{Name: guser.Username}, nil
}

func (g *Client) GetRepository(ctx context.Context, identifier string) (cloud.Repository, error) {
	proj, _, err := g.client.Projects.GetProject(identifier, nil)
	if err != nil {
		return cloud.Repository{}, err
	}

	return cloud.Repository{
		Path:          proj.PathWithNamespace,
		DefaultBranch: proj.DefaultBranch,
	}, nil
}

func (g *Client) ListRepositories(ctx context.Context, lopts cloud.ListRepositoriesOptions) ([]string, error) {
	opts := &gitlab.ListProjectsOptions{
		ListOptions: gitlab.ListOptions{
			PerPage: lopts.PageSize,
		},
		// limit results to those repos the authenticated user is a member of,
		// otherwise we'll get *all* accessible repos, public and private.
		Membership: internal.Bool(true),
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
		Search: internal.String("^" + opts.Prefix),
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

func (g *Client) GetRepoTarball(ctx context.Context, opts cloud.GetRepoTarballOptions) ([]byte, string, error) {
	owner, name, found := strings.Cut(opts.Repo, "/")
	if !found {
		return nil, "", fmt.Errorf("malformed identifier: %s", opts.Repo)
	}

	tarball, _, err := g.client.Repositories.Archive(opts.Repo, &gitlab.ArchiveOptions{
		Format: internal.String("tar.gz"),
		SHA:    opts.Ref,
	})
	if err != nil {
		return nil, "", err
	}

	// Gitlab tarball contents are contained within a top-level directory
	// formatted <ref>-<sha>. We want the tarball without this directory,
	// so we re-tar the contents without the top-level directory.
	untarpath, err := os.MkdirTemp("", fmt.Sprintf("gitlab-%s-%s-*", owner, name))
	if err != nil {
		return nil, "", err
	}
	defer os.RemoveAll(untarpath)

	if err := internal.Unpack(bytes.NewReader(tarball), untarpath); err != nil {
		return nil, "", err
	}
	contents, err := os.ReadDir(untarpath)
	if err != nil {
		return nil, "", err
	}
	if len(contents) != 1 {
		return nil, "", fmt.Errorf("expected only one top-level directory; instead got %s", contents)
	}
	dir := contents[0].Name()
	parts := strings.Split(dir, "-")
	if len(parts) != 2 {
		return nil, "", fmt.Errorf("malformed directory name found in tarball: %s", dir)
	}
	tarball, err = internal.Pack(path.Join(untarpath, dir))
	if err != nil {
		return nil, "", err
	}
	return tarball, parts[1], nil
}

func (g *Client) CreateWebhook(ctx context.Context, opts cloud.CreateWebhookOptions) (string, error) {
	addOpts := &gitlab.AddProjectHookOptions{
		EnableSSLVerification: internal.Bool(true),
		PushEvents:            internal.Bool(true),
		Token:                 internal.String(opts.Secret),
		URL:                   internal.String(opts.Endpoint),
	}
	for _, event := range opts.Events {
		switch event {
		case cloud.VCSEventTypePush:
			addOpts.PushEvents = internal.Bool(true)
		case cloud.VCSEventTypePull:
			addOpts.MergeRequestsEvents = internal.Bool(true)
		}
	}

	hook, _, err := g.client.Projects.AddProjectHook(opts.Repo, addOpts)
	if err != nil {
		return "", err
	}
	return strconv.Itoa(hook.ID), nil
}

func (g *Client) UpdateWebhook(ctx context.Context, id string, opts cloud.UpdateWebhookOptions) error {
	intID, err := strconv.Atoi(id)
	if err != nil {
		return err
	}

	editOpts := &gitlab.EditProjectHookOptions{
		EnableSSLVerification: internal.Bool(true),
		Token:                 internal.String(opts.Secret),
		URL:                   internal.String(opts.Endpoint),
	}
	for _, event := range opts.Events {
		switch event {
		case cloud.VCSEventTypePush:
			editOpts.PushEvents = internal.Bool(true)
		case cloud.VCSEventTypePull:
			editOpts.MergeRequestsEvents = internal.Bool(true)
		}
	}

	_, _, err = g.client.Projects.EditProjectHook(opts.Repo, intID, editOpts)
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
			return cloud.Webhook{}, internal.ErrResourceNotFound
		}
		return cloud.Webhook{}, err
	}

	var events []cloud.VCSEventType
	if hook.PushEvents {
		events = append(events, cloud.VCSEventTypePush)
	}
	if hook.MergeRequestsEvents {
		events = append(events, cloud.VCSEventTypePull)
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

func (g *Client) ListPullRequestFiles(ctx context.Context, repo string, pull int) ([]string, error) {
	return nil, nil
}

func (g *Client) GetCommit(ctx context.Context, repo, ref string) (cloud.Commit, error) {
	return cloud.Commit{}, nil
}
