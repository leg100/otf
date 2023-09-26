package github

import (
	"context"

	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	otfhttp "github.com/leg100/otf/internal/http"

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
			Transport: otfhttp.DefaultTransport(true),
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

func (g *Client) GetCurrentUser(ctx context.Context) (cloud.User, error) {
	guser, _, err := g.client.Users.Get(ctx, "")
	if err != nil {
		return cloud.User{}, err
	}
	return cloud.User{Name: guser.GetLogin()}, nil
}

func (g *Client) GetRepository(ctx context.Context, identifier string) (cloud.Repository, error) {
	owner, name, found := strings.Cut(identifier, "/")
	if !found {
		return cloud.Repository{}, fmt.Errorf("malformed identifier: %s", identifier)
	}
	repo, _, err := g.client.Repositories.Get(ctx, owner, name)
	if err != nil {
		return cloud.Repository{}, err
	}

	return cloud.Repository{
		Path:          identifier,
		DefaultBranch: repo.GetDefaultBranch(),
	}, nil
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

func (g *Client) GetRepoTarball(ctx context.Context, opts cloud.GetRepoTarballOptions) ([]byte, string, error) {
	owner, name, found := strings.Cut(opts.Repo, "/")
	if !found {
		return nil, "", fmt.Errorf("malformed identifier: %s", opts.Repo)
	}

	var gopts github.RepositoryContentGetOptions
	if opts.Ref != nil {
		gopts.Ref = *opts.Ref
	}

	link, _, err := g.client.Repositories.GetArchiveLink(ctx, owner, name, github.Tarball, &gopts, true)
	if err != nil {
		return nil, "", err
	}

	resp, err := g.client.Client().Get(link.String())
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		// parse out the query string which contains credential
		u, _ := url.Parse(link.String())
		return nil, "", fmt.Errorf("non-200 status code: %d: %s", resp.StatusCode, u.Path)
	}

	// github tarball contains a parent directory of the format
	// <owner>-<repo>-<commit>. We need a tarball without this parent directory,
	// so we untar it to a temp dir, then tar it up the contents of the parent
	// directory.
	untarpath, err := os.MkdirTemp("", fmt.Sprintf("github-%s-%s-*", owner, name))
	if err != nil {
		return nil, "", err
	}
	defer os.RemoveAll(untarpath)

	if err := internal.Unpack(resp.Body, untarpath); err != nil {
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
	if len(parts) < 3 {
		return nil, "", fmt.Errorf("malformed directory name found in tarball: %s", dir)
	}
	parentDir := path.Join(untarpath, dir)
	tarball, err := internal.Pack(parentDir)
	if err != nil {
		return nil, "", err
	}
	return tarball, parts[len(parts)-1], nil
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
		case cloud.VCSEventTypePush:
			events = append(events, "push")
		case cloud.VCSEventTypePull:
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
		case cloud.VCSEventTypePush:
			events = append(events, "push")
		case cloud.VCSEventTypePull:
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
			events = append(events, cloud.VCSEventTypePush)
		case "pull_request":
			events = append(events, cloud.VCSEventTypePull)
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

func (g *Client) ListPullRequestFiles(ctx context.Context, repo string, pull int) ([]string, error) {
	owner, name, found := strings.Cut(repo, "/")
	if !found {
		return nil, fmt.Errorf("malformed identifier: %s", repo)
	}

	var files []string
	nextPage := 0

listloop:
	for {
		opts := github.ListOptions{
			PerPage: 300,
		}
		if nextPage != 0 {
			opts.Page = nextPage
		}
		// GitHub has started to return 404's sometimes. They've got some
		// eventual consistency issues going on so we're just going to attempt
		// up to 5 times for each page with exponential backoff.
		maxAttempts := 5
		attemptDelay := 0 * time.Second
		for i := 0; i < maxAttempts; i++ {
			// First don't sleep, then sleep 1, 3, 7, etc.
			time.Sleep(attemptDelay)
			attemptDelay = 2*attemptDelay + 1*time.Second

			pageFiles, resp, err := g.client.PullRequests.ListFiles(ctx, owner, name, pull, &opts)
			if err != nil {
				ghErr, ok := err.(*github.ErrorResponse)
				if ok && ghErr.Response.StatusCode == 404 {
					// (hopefully) transient 404, retry after backoff
					continue
				}
				// something else, give up
				return files, err
			}
			for _, f := range pageFiles {
				files = append(files, f.GetFilename())

				// If the file was renamed, we'll want to run plan in the directory
				// it was moved from as well.
				if f.GetStatus() == "renamed" {
					files = append(files, f.GetPreviousFilename())
				}
			}
			if resp.NextPage == 0 {
				break listloop
			}
			nextPage = resp.NextPage
			break
		}
	}
	return files, nil
}

func (g *Client) GetCommit(ctx context.Context, repo, ref string) (cloud.Commit, error) {
	owner, name, found := strings.Cut(repo, "/")
	if !found {
		return cloud.Commit{}, fmt.Errorf("malformed identifier: %s", repo)
	}

	commit, resp, err := g.client.Repositories.GetCommit(ctx, owner, name, ref, nil)
	if err != nil {
		return cloud.Commit{}, err
	}
	defer resp.Body.Close()

	return cloud.Commit{
		SHA: commit.GetSHA(),
		URL: commit.GetHTMLURL(),
		Author: cloud.CommitAuthor{
			Username:   commit.GetAuthor().GetLogin(),
			AvatarURL:  commit.GetAuthor().GetAvatarURL(),
			ProfileURL: commit.GetAuthor().GetHTMLURL(),
		},
	}, nil
}
