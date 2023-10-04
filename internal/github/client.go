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

	"github.com/bradleyfalzon/ghinstallation/v2"
	"github.com/google/go-github/v55/github"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/authenticator"
	otfhttp "github.com/leg100/otf/internal/http"
	"github.com/leg100/otf/internal/vcs"
	"golang.org/x/oauth2"
)

type (
	// Client is a wrapper around the upstream go-github client
	Client struct {
		client *github.Client
	}

	ClientOptions struct {
		Hostname            string
		SkipTLSVerification bool

		// Only specify one of the following
		OAuthToken    *oauth2.Token
		PersonalToken *string
		*AppCredentials
		*InstallCredentials
	}

	// Credentials for authenticating as an app:
	// https://docs.github.com/en/apps/creating-github-apps/authenticating-with-a-github-app/about-authentication-with-a-github-app#authentication-as-a-github-app
	AppCredentials struct {
		ID         int64
		PrivateKey string
	}

	// Credentials for authenticating as an app installation:
	// https://docs.github.com/en/apps/creating-github-apps/authenticating-with-a-github-app/about-authentication-with-a-github-app#authentication-as-an-app-installation
	InstallCredentials struct {
		ID int64
		AppCredentials
	}
)

func NewClient(cfg ClientOptions) (*Client, error) {
	if cfg.Hostname == "" {
		cfg.Hostname = DefaultHostname
	}
	// build http roundtripper using provided credentials
	var (
		tripper = otfhttp.DefaultTransport(cfg.SkipTLSVerification)
		err     error
	)
	switch {
	case cfg.AppCredentials != nil:
		tripper, err = ghinstallation.NewAppsTransport(tripper, cfg.AppCredentials.ID, []byte(cfg.AppCredentials.PrivateKey))
		if err != nil {
			return nil, err
		}
	case cfg.InstallCredentials != nil:
		creds := cfg.InstallCredentials
		tripper, err = ghinstallation.New(tripper, creds.AppCredentials.ID, creds.ID, []byte(creds.AppCredentials.PrivateKey))
		if err != nil {
			return nil, err
		}
	case cfg.PersonalToken != nil:
		// personal token is actually an OAuth2 *access token, so wrap
		// inside an OAuth2 token and handle it the same as an OAuth2 token
		cfg.OAuthToken = &oauth2.Token{AccessToken: *cfg.PersonalToken}
		fallthrough
	case cfg.OAuthToken != nil:
		tripper = &oauth2.Transport{
			Base: tripper,
			// Github's oauth access token never expires
			Source: oauth2.ReuseTokenSource(nil, oauth2.StaticTokenSource(cfg.OAuthToken)),
		}
	}
	// create upstream client with roundtripper
	client := github.NewClient(&http.Client{Transport: tripper})
	// Assume github enterprise if using non-default hostname
	if cfg.Hostname != DefaultHostname {
		client, err = client.WithEnterpriseURLs(
			"https://"+cfg.Hostname,
			"https://"+cfg.Hostname,
		)
		if err != nil {
			return nil, err
		}
	}
	return &Client{client: client}, nil
}

func NewPersonalTokenClient(hostname, token string) (vcs.Client, error) {
	return NewClient(ClientOptions{
		Hostname:      hostname,
		PersonalToken: &token,
	})
}

func NewOAuthClient(cfg authenticator.OAuthConfig, token *oauth2.Token) (authenticator.IdentityProviderClient, error) {
	return NewClient(ClientOptions{
		Hostname:   cfg.Hostname,
		OAuthToken: token,
	})
}

func (g *Client) GetCurrentUser(ctx context.Context) (string, error) {
	guser, _, err := g.client.Users.Get(ctx, "")
	if err != nil {
		return "", err
	}
	return guser.GetLogin(), nil
}

func (g *Client) GetRepository(ctx context.Context, identifier string) (vcs.Repository, error) {
	owner, name, found := strings.Cut(identifier, "/")
	if !found {
		return vcs.Repository{}, fmt.Errorf("malformed identifier: %s", identifier)
	}
	repo, _, err := g.client.Repositories.Get(ctx, owner, name)
	if err != nil {
		return vcs.Repository{}, err
	}

	return vcs.Repository{
		Path:          identifier,
		DefaultBranch: repo.GetDefaultBranch(),
	}, nil
}

func (g *Client) ListRepositories(ctx context.Context, opts vcs.ListRepositoriesOptions) ([]string, error) {
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

func (g *Client) ListTags(ctx context.Context, opts vcs.ListTagsOptions) ([]string, error) {
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

func (g *Client) ExchangeCode(ctx context.Context, code string) (*github.AppConfig, error) {
	cfg, _, err := g.client.Apps.CompleteAppManifest(ctx, code)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

func (g *Client) GetRepoTarball(ctx context.Context, opts vcs.GetRepoTarballOptions) ([]byte, string, error) {
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
func (g *Client) CreateWebhook(ctx context.Context, opts vcs.CreateWebhookOptions) (string, error) {
	owner, name, found := strings.Cut(opts.Repo, "/")
	if !found {
		return "", fmt.Errorf("malformed identifier: %s", opts.Repo)
	}

	var events []string
	for _, event := range opts.Events {
		switch event {
		case vcs.EventTypePush:
			events = append(events, "push")
		case vcs.EventTypePull:
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

func (g *Client) UpdateWebhook(ctx context.Context, id string, opts vcs.UpdateWebhookOptions) error {
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
		case vcs.EventTypePush:
			events = append(events, "push")
		case vcs.EventTypePull:
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

func (g *Client) GetWebhook(ctx context.Context, opts vcs.GetWebhookOptions) (vcs.Webhook, error) {
	owner, name, found := strings.Cut(opts.Repo, "/")
	if !found {
		return vcs.Webhook{}, fmt.Errorf("malformed identifier: %s", opts.Repo)
	}

	intID, err := strconv.ParseInt(opts.ID, 10, 64)
	if err != nil {
		return vcs.Webhook{}, err
	}

	hook, resp, err := g.client.Repositories.GetHook(ctx, owner, name, intID)
	if err != nil {
		if resp.StatusCode == http.StatusNotFound {
			return vcs.Webhook{}, internal.ErrResourceNotFound
		}
		return vcs.Webhook{}, err
	}

	var events []vcs.EventType
	for _, event := range hook.Events {
		switch event {
		case "push":
			events = append(events, vcs.EventTypePush)
		case "pull_request":
			events = append(events, vcs.EventTypePull)
		}
	}

	// extracting OTF endpoint from github's config map is a bit of work...
	rawEndpoint, ok := hook.Config["url"]
	if !ok {
		return vcs.Webhook{}, errors.New("missing url")
	}
	endpoint, ok := rawEndpoint.(string)
	if !ok {
		return vcs.Webhook{}, errors.New("url is not a string")
	}

	return vcs.Webhook{
		ID:       strconv.FormatInt(hook.GetID(), 10),
		Repo:     opts.Repo,
		Events:   events,
		Endpoint: endpoint,
	}, nil
}

func (g *Client) DeleteWebhook(ctx context.Context, opts vcs.DeleteWebhookOptions) error {
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

func (g *Client) SetStatus(ctx context.Context, opts vcs.SetStatusOptions) error {
	owner, name, found := strings.Cut(opts.Repo, "/")
	if !found {
		return fmt.Errorf("malformed identifier: %s", opts.Repo)
	}

	var status string
	switch opts.Status {
	case vcs.PendingStatus, vcs.RunningStatus:
		status = "pending"
	case vcs.SuccessStatus:
		status = "success"
	case vcs.ErrorStatus:
		status = "error"
	case vcs.FailureStatus:
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

func (g *Client) GetCommit(ctx context.Context, repo, ref string) (vcs.Commit, error) {
	owner, name, found := strings.Cut(repo, "/")
	if !found {
		return vcs.Commit{}, fmt.Errorf("malformed identifier: %s", repo)
	}

	commit, resp, err := g.client.Repositories.GetCommit(ctx, owner, name, ref, nil)
	if err != nil {
		return vcs.Commit{}, err
	}
	defer resp.Body.Close()

	return vcs.Commit{
		SHA: commit.GetSHA(),
		URL: commit.GetHTMLURL(),
		Author: vcs.CommitAuthor{
			Username:   commit.GetAuthor().GetLogin(),
			AvatarURL:  commit.GetAuthor().GetAvatarURL(),
			ProfileURL: commit.GetAuthor().GetHTMLURL(),
		},
	}, nil
}

// ListInstallations lists installations of the currently authenticated app.
func (g *Client) ListInstallations(ctx context.Context) ([]*github.Installation, error) {
	installs, resp, err := g.client.Apps.ListInstallations(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return installs, err
}

// DeleteInstallation deletes an installation of a github app with the given
// installation ID.
func (g *Client) DeleteInstallation(ctx context.Context, installID int64) error {
	resp, err := g.client.Apps.DeleteInstallation(ctx, installID)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return err
}
