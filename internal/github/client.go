package github

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/google/go-github/v41/github"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/cloud"
	"golang.org/x/oauth2"
)

type Client struct {
	client *github.Client
}

func NewClient(ctx context.Context, cfg cloud.ClientOptions) (*Client, error) {
	var (
		client *github.Client
		err    error
	)

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
		client, err = NewEnterpriseClient(cfg.Hostname, httpClient)
		if err != nil {
			return nil, err
		}
	} else {
		client = github.NewClient(httpClient)
	}
	return &Client{client: client}, nil
}

func NewEnterpriseClient(hostname string, httpClient *http.Client) (*github.Client, error) {
	return github.NewEnterpriseClient(
		"https://"+hostname,
		"https://"+hostname,
		httpClient)
}

// GetUser retrieves a user from github along with their team memberships.
// They are determined to be in the owners team if either:
// (a) they are a member of a github team called 'owners'
// (b) they are an admin of a github organization.
func (g *Client) GetUser(ctx context.Context) (*cloud.User, error) {
	guser, _, err := g.client.Users.Get(ctx, "")
	if err != nil {
		return nil, err
	}

	gteams, _, err := g.client.Teams.ListUserTeams(ctx, nil)
	if err != nil {
		return nil, err
	}

	user := cloud.User{Name: guser.GetLogin()}

	// Convert each github team to a cloud team. Use the github slug rather than
	// the github name because the latter often contains whitespace, and OTF
	// names should not contain whitespace.
	for _, gteam := range gteams {
		user.Teams = append(user.Teams, cloud.Team{
			Name:         gteam.GetSlug(),
			Organization: gteam.GetOrganization().GetLogin(),
		})
	}

	// check if they are an admin of any orgs; if so, make them an owner.
	organizations, _, err := g.client.Organizations.List(ctx, "", nil)
	if err != nil {
		return nil, err
	}
	for _, org := range organizations {
		if user.IsOwner(org.GetLogin()) {
			// user already determined to be an owner, so skip this check.
			continue
		}
		membership, _, err := g.client.Organizations.GetOrgMembership(ctx, "", org.GetLogin())
		if err != nil {
			return nil, err
		}
		if membership.GetRole() == "admin" {
			user.Teams = append(user.Teams, cloud.Team{
				Name:         "owners",
				Organization: org.GetLogin(),
			})
		}
	}

	return &user, nil
}

func (g *Client) GetRepository(ctx context.Context, identifier string) (string, error) {
	owner, name, found := strings.Cut(identifier, "/")
	if !found {
		return "", fmt.Errorf("malformed identifier: %s", identifier)
	}
	repo, _, err := g.client.Repositories.Get(ctx, owner, name)
	if err != nil {
		return "", err
	}

	return repo.GetFullName(), nil
}

func (g *Client) ListRepositories(ctx context.Context, opts cloud.ListRepositoriesOptions) ([]string, error) {
	repos, _, err := g.client.Repositories.List(ctx, "", &github.RepositoryListOptions{
		ListOptions: github.ListOptions{
			PerPage: opts.PageSize,
		},
		// retrieve repositories in order of most recently pushed to
		Sort: "pushed",
	})
	if err != nil {
		return nil, err
	}

	var names []string
	for _, repo := range repos {
		names = append(names, repo.GetFullName())
	}
	return names, nil
}

func (g *Client) ListTags(ctx context.Context, opts cloud.ListTagsOptions) ([]string, error) {
	owner, name, found := strings.Cut(opts.Repo, "/")
	if !found {
		return nil, fmt.Errorf("malformed identifier: %s", opts.Repo)
	}

	results, _, err := g.client.Git.ListMatchingRefs(ctx, owner, name, &github.ReferenceListOptions{
		Ref: "tags/" + opts.Prefix,
	})
	if err != nil {
		return nil, err
	}

	// return tags with the format 'tags/<tag_value>'
	var tags []string
	for _, ref := range results {
		tags = append(tags, strings.TrimPrefix(ref.GetRef(), "refs/"))
	}
	return tags, nil
}

func (g *Client) GetRepoTarball(ctx context.Context, opts cloud.GetRepoTarballOptions) ([]byte, error) {
	owner, name, found := strings.Cut(opts.Repo, "/")
	if !found {
		return nil, fmt.Errorf("malformed identifier: %s", opts.Repo)
	}

	var gopts github.RepositoryContentGetOptions
	if opts.Ref != nil {
		gopts.Ref = *opts.Ref
	}

	link, _, err := g.client.Repositories.GetArchiveLink(ctx, owner, name, github.Tarball, &gopts, true)
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
	//
	// TODO: remove temp dir after finishing
	untarpath, err := os.MkdirTemp("", fmt.Sprintf("github-%s-%s-*", owner, name))
	if err != nil {
		return nil, err
	}
	if err := internal.Unpack(resp.Body, untarpath); err != nil {
		return nil, err
	}
	contents, err := os.ReadDir(untarpath)
	if err != nil {
		return nil, err
	}
	if len(contents) != 1 {
		return nil, fmt.Errorf("expected only one top-level directory; instead got %s", contents)
	}
	parentDir := path.Join(untarpath, contents[0].Name())
	return internal.Pack(parentDir)
}

// CreateWebhook creates a webhook on a github repository.
func (g *Client) CreateWebhook(ctx context.Context, opts cloud.CreateWebhookOptions) (string, error) {
	owner, name, found := strings.Cut(opts.Repo, "/")
	if !found {
		return "", fmt.Errorf("malformed identifier: %s", opts.Repo)
	}

	var events []string
	for _, event := range opts.Events {
		switch event {
		case cloud.VCSPushEventType:
			events = append(events, "push")
		case cloud.VCSPullEventType:
			events = append(events, "pull_request")
		}
	}

	hook, _, err := g.client.Repositories.CreateHook(ctx, owner, name, &github.Hook{
		Events: events,
		Config: map[string]any{
			"url":          opts.Endpoint,
			"secret":       opts.Secret,
			"content_type": "json",
		},
		Active: internal.Bool(true),
	})
	if err != nil {
		return "", err
	}
	return strconv.FormatInt(hook.GetID(), 10), nil
}

func (g *Client) UpdateWebhook(ctx context.Context, id string, opts cloud.UpdateWebhookOptions) error {
	owner, name, found := strings.Cut(opts.Repo, "/")
	if !found {
		return fmt.Errorf("malformed identifier: %s", opts.Repo)
	}

	intID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return err
	}

	var events []string
	for _, event := range opts.Events {
		switch event {
		case cloud.VCSPushEventType:
			events = append(events, "push")
		case cloud.VCSPullEventType:
			events = append(events, "pull_request")
		}
	}

	_, _, err = g.client.Repositories.EditHook(ctx, owner, name, intID, &github.Hook{
		Events: events,
		Config: map[string]any{
			"url":          opts.Endpoint,
			"secret":       opts.Secret,
			"content_type": "json",
		},
		Active: internal.Bool(true),
	})
	if err != nil {
		return err
	}
	return nil
}

func (g *Client) GetWebhook(ctx context.Context, opts cloud.GetWebhookOptions) (cloud.Webhook, error) {
	owner, name, found := strings.Cut(opts.Repo, "/")
	if !found {
		return cloud.Webhook{}, fmt.Errorf("malformed identifier: %s", opts.Repo)
	}

	intID, err := strconv.ParseInt(opts.ID, 10, 64)
	if err != nil {
		return cloud.Webhook{}, err
	}

	hook, resp, err := g.client.Repositories.GetHook(ctx, owner, name, intID)
	if err != nil {
		if resp.StatusCode == http.StatusNotFound {
			return cloud.Webhook{}, internal.ErrResourceNotFound
		}
		return cloud.Webhook{}, err
	}

	var events []cloud.VCSEventType
	for _, event := range hook.Events {
		switch event {
		case "push":
			events = append(events, cloud.VCSPushEventType)
		case "pull_request":
			events = append(events, cloud.VCSPullEventType)
		}
	}

	// extracting OTF endpoint from github's config map is a bit of work...
	rawEndpoint, ok := hook.Config["url"]
	if !ok {
		return cloud.Webhook{}, errors.New("missing url")
	}
	endpoint, ok := rawEndpoint.(string)
	if !ok {
		return cloud.Webhook{}, errors.New("url is not a string")
	}

	return cloud.Webhook{
		ID:       strconv.FormatInt(hook.GetID(), 10),
		Repo:     opts.Repo,
		Events:   events,
		Endpoint: endpoint,
	}, nil
}

func (g *Client) DeleteWebhook(ctx context.Context, opts cloud.DeleteWebhookOptions) error {
	owner, name, found := strings.Cut(opts.Repo, "/")
	if !found {
		return fmt.Errorf("malformed identifier: %s", opts.Repo)
	}

	intID, err := strconv.ParseInt(opts.ID, 10, 64)
	if err != nil {
		return err
	}

	_, err = g.client.Repositories.DeleteHook(ctx, owner, name, intID)
	return err
}

func (g *Client) SetStatus(ctx context.Context, opts cloud.SetStatusOptions) error {
	owner, name, found := strings.Cut(opts.Repo, "/")
	if !found {
		return fmt.Errorf("malformed identifier: %s", opts.Repo)
	}

	var status string
	switch opts.Status {
	case cloud.VCSPendingStatus, cloud.VCSRunningStatus:
		status = "pending"
	case cloud.VCSSuccessStatus:
		status = "success"
	case cloud.VCSErrorStatus:
		status = "error"
	case cloud.VCSFailureStatus:
		status = "failure"
	default:
		return fmt.Errorf("invalid vcs status: %s", opts.Status)
	}

	_, _, err := g.client.Repositories.CreateStatus(ctx, owner, name, opts.Ref, &github.RepoStatus{
		Context:     internal.String(fmt.Sprintf("otf/%s", opts.Workspace)),
		TargetURL:   internal.String(opts.TargetURL),
		Description: internal.String(opts.Description),
		State:       internal.String(status),
	})
	return err
}
