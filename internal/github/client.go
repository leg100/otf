package github

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/bradleyfalzon/ghinstallation/v2"
	"github.com/google/go-github/v65/github"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/authenticator"
	otfhttp "github.com/leg100/otf/internal/http"
	"github.com/leg100/otf/internal/user"
	"github.com/leg100/otf/internal/vcs"
	"golang.org/x/oauth2"
)

// maxRedirects is the number of HTTP 301 redirects an API request is permitted
// to follow.
const maxRedirects = 10

type (
	// Client is a wrapper around the upstream go-github client
	Client struct {
		client *github.Client

		// whether authenticated using an installation access token
		iat bool
	}

	ClientOptions struct {
		// BaseURL is the base URL for github.
		BaseURL             *internal.WebURL
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
		// Github app ID
		ID AppID
		// Private key in PEM format
		PrivateKey string
	}

	// Credentials for authenticating as an app installation:
	// https://docs.github.com/en/apps/creating-github-apps/authenticating-with-a-github-app/about-authentication-with-a-github-app#authentication-as-an-app-installation
	InstallCredentials struct {
		// Github installation ID
		ID int64
		// Github username if installed in a user account; mutually exclusive
		// with Organization
		User *string
		// Github organization if installed in an organization; mutually
		// exclusive with User
		Organization *string

		AppCredentials
	}
)

func NewClient(cfg ClientOptions) (*Client, error) {
	baseURL, uploadURL := setClientURLs(cfg.BaseURL)

	// build http roundtripper using provided credentials
	var (
		tripper = http.DefaultTransport
		err     error
		iat     bool
	)
	if cfg.SkipTLSVerification {
		tripper = otfhttp.InsecureTransport
	}
	switch {
	case cfg.AppCredentials != nil:
		tripper, err = ghinstallation.NewAppsTransport(tripper, int64(cfg.AppCredentials.ID), []byte(cfg.AppCredentials.PrivateKey))
		if err != nil {
			return nil, err
		}
	case cfg.InstallCredentials != nil:
		iat = true
		creds := cfg.InstallCredentials
		installTransport, err := ghinstallation.New(tripper, int64(creds.AppCredentials.ID), creds.ID, []byte(creds.AppCredentials.PrivateKey))
		if err != nil {
			return nil, err
		}
		installTransport.BaseURL = baseURL.String()
		tripper = installTransport
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

	client.BaseURL = &baseURL.URL
	client.UploadURL = &uploadURL.URL

	return &Client{client: client, iat: iat}, nil
}

func setClientURLs(githubURL *internal.WebURL) (*internal.WebURL, *internal.WebURL) {
	cloned := *githubURL
	// If using public github (github.com) then use api.github.com
	if cloned.Host == DefaultBaseURL.Host {
		cloned.Host = "api." + cloned.Host
	}

	baseURL := cloned
	if !strings.HasSuffix(baseURL.Path, "/") {
		baseURL.Path += "/"
	}
	if !strings.HasSuffix(baseURL.Path, "/api/v3/") &&
		!strings.HasPrefix(baseURL.Host, "api.") &&
		!strings.Contains(baseURL.Host, ".api.") {
		baseURL.Path += "api/v3/"
	}

	uploadURL := cloned
	if !strings.HasSuffix(uploadURL.Path, "/") {
		uploadURL.Path += "/"
	}
	if !strings.HasSuffix(uploadURL.Path, "/api/uploads/") &&
		!strings.HasPrefix(uploadURL.Host, "api.") &&
		!strings.Contains(uploadURL.Host, ".api.") {
		uploadURL.Path += "api/uploads/"
	}
	return &baseURL, &uploadURL
}

func NewTokenClient(opts vcs.NewTokenClientOptions) (vcs.Client, error) {
	return NewClient(ClientOptions{
		BaseURL:             opts.BaseURL,
		PersonalToken:       &opts.Token,
		SkipTLSVerification: opts.SkipTLSVerification,
	})
}

func NewOAuthClient(cfg authenticator.OAuthConfig, token *oauth2.Token) (authenticator.IdentityProviderClient, error) {
	return NewClient(ClientOptions{
		BaseURL:             cfg.BaseURL,
		OAuthToken:          token,
		SkipTLSVerification: cfg.SkipTLSVerification,
	})
}

func (g *Client) GetCurrentUser(ctx context.Context) (authenticator.UserInfo, error) {
	guser, _, err := g.client.Users.Get(ctx, "")
	if err != nil {
		return authenticator.UserInfo{}, err
	}
	username, err := user.NewUsername(guser.GetLogin())
	if err != nil {
		return authenticator.UserInfo{}, err
	}
	return authenticator.UserInfo{
		Username:  username,
		AvatarURL: guser.AvatarURL,
	}, nil
}

func (g *Client) GetDefaultBranch(ctx context.Context, identifier string) (string, error) {
	owner, name, found := strings.Cut(identifier, "/")
	if !found {
		return "", fmt.Errorf("malformed identifier: %s", identifier)
	}
	repo, _, err := g.client.Repositories.Get(ctx, owner, name)
	if err != nil {
		return "", err
	}

	return repo.GetDefaultBranch(), nil
}

// ListRepositories lists repositories belonging to the authenticated entity: if
// authenticated using a user's oauth token or PAT then their repos are listed;
// if authenticated using a github installation then repos that the installation
// has access to are listed.
//

// ListRepositories has different behaviour depending on the authentication:
// (a) if authenticated as an app installation then repositories accessible to
// the installation are listed; *all* repos are listed, in order of last pushed
// to.
// (b) if authenticated using a personal access token then repositories
// belonging to the user are listed; only the first page of repos is listed,
// those that have most recently been pushed to.
func (g *Client) ListRepositories(ctx context.Context, opts vcs.ListRepositoriesOptions) ([]vcs.Repo, error) {
	var repos []*github.Repository
	if g.iat {
		// Apps.ListRepos endpoint does not support ordering on the server-side,
		// so instead we request *all* repos, page-by-page, and then sort
		// client-side.
		page := 1
		for {
			result, resp, err := g.client.Apps.ListRepos(ctx, &github.ListOptions{
				PerPage: opts.PageSize,
				Page:    page,
			})
			if err != nil {
				return nil, err
			}
			repos = append(repos, result.Repositories...)
			if resp.NextPage != 0 {
				page = resp.NextPage
			} else {
				break
			}
		}
		// sort repositories in order of most recently pushed to
		sort.Slice(repos, func(i, j int) bool { return repos[i].GetPushedAt().After(repos[j].GetPushedAt().Time) })
	} else {
		var err error
		repos, _, err = g.client.Repositories.ListByAuthenticatedUser(ctx, &github.RepositoryListByAuthenticatedUserOptions{
			ListOptions: github.ListOptions{PerPage: opts.PageSize},
			// retrieve repositories in order of most recently pushed to
			Sort: "pushed",
		})
		if err != nil {
			return nil, err
		}
	}
	names := make([]vcs.Repo, len(repos))
	for i, repo := range repos {
		names[i] = vcs.NewMustRepo(repo.Owner.GetLogin(), repo.GetName())
	}
	return names, nil
}

func (g *Client) ListTags(ctx context.Context, opts vcs.ListTagsOptions) ([]string, error) {
	results, _, err := g.client.Git.ListMatchingRefs(ctx, opts.Repo.Owner(), opts.Repo.Name(), &github.ReferenceListOptions{
		Ref: "tags/" + opts.Prefix,
	})
	if err != nil {
		return nil, err
	}

	// return tags with the format 'tags/<tag_value>'
	tags := make([]string, len(results))
	for i, ref := range results {
		tags[i] = strings.TrimPrefix(ref.GetRef(), "refs/")
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
	var gopts github.RepositoryContentGetOptions
	if opts.Ref != nil {
		gopts.Ref = *opts.Ref
	}

	link, _, err := g.client.Repositories.GetArchiveLink(ctx, opts.Repo.Owner(), opts.Repo.Name(), github.Tarball, &gopts, maxRedirects)
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
	untarpath, err := os.MkdirTemp("", fmt.Sprintf("github-%s-%s-*", opts.Repo.Owner(), opts.Repo.Name()))
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
	var events []string
	for _, event := range opts.Events {
		switch event {
		case vcs.EventTypePush:
			events = append(events, "push")
		case vcs.EventTypePull:
			events = append(events, "pull_request")
		}
	}

	hook, _, err := g.client.Repositories.CreateHook(ctx, opts.Repo.Owner(), opts.Repo.Name(), &github.Hook{
		Events: events,
		Config: &github.HookConfig{
			URL:         &opts.Endpoint,
			Secret:      &opts.Secret,
			ContentType: internal.Ptr("json"),
		},
		Active: internal.Ptr(true),
	})
	if err != nil {
		return "", err
	}
	return strconv.FormatInt(hook.GetID(), 10), nil
}

func (g *Client) UpdateWebhook(ctx context.Context, id string, opts vcs.UpdateWebhookOptions) error {
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

	_, _, err = g.client.Repositories.EditHook(ctx, opts.Repo.Owner(), opts.Repo.Name(), intID, &github.Hook{
		Events: events,
		Config: &github.HookConfig{
			URL:         &opts.Endpoint,
			Secret:      &opts.Secret,
			ContentType: internal.Ptr("json"),
		},
		Active: internal.Ptr(true),
	})
	if err != nil {
		return err
	}
	return nil
}

func (g *Client) GetWebhook(ctx context.Context, opts vcs.GetWebhookOptions) (vcs.Webhook, error) {
	intID, err := strconv.ParseInt(opts.ID, 10, 64)
	if err != nil {
		return vcs.Webhook{}, err
	}

	hook, resp, err := g.client.Repositories.GetHook(ctx, opts.Repo.Owner(), opts.Repo.Name(), intID)
	if err != nil {
		if resp != nil && resp.StatusCode == http.StatusNotFound {
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

	url := hook.Config.URL
	if url == nil {
		return vcs.Webhook{}, errors.New("missing url")
	}

	return vcs.Webhook{
		ID:       strconv.FormatInt(hook.GetID(), 10),
		Repo:     opts.Repo,
		Events:   events,
		Endpoint: *url,
	}, nil
}

func (g *Client) DeleteWebhook(ctx context.Context, opts vcs.DeleteWebhookOptions) error {
	intID, err := strconv.ParseInt(opts.ID, 10, 64)
	if err != nil {
		return err
	}

	_, err = g.client.Repositories.DeleteHook(ctx, opts.Repo.Owner(), opts.Repo.Name(), intID)
	return err
}

func (g *Client) SetStatus(ctx context.Context, opts vcs.SetStatusOptions) error {
	var status string
	switch opts.Status {
	case vcs.PendingStatus:
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

	_, _, err := g.client.Repositories.CreateStatus(ctx, opts.Repo.Owner(), opts.Repo.Name(), opts.Ref, &github.RepoStatus{
		Context:     internal.Ptr(fmt.Sprintf("otf/%s", opts.Workspace)),
		TargetURL:   internal.Ptr(opts.TargetURL),
		Description: internal.Ptr(opts.Description),
		State:       internal.Ptr(status),
	})
	return err
}

func (g *Client) ListPullRequestFiles(ctx context.Context, repo vcs.Repo, pull int) ([]string, error) {
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

			pageFiles, resp, err := g.client.PullRequests.ListFiles(ctx, repo.Owner(), repo.Name(), pull, &opts)
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

func (g *Client) GetCommit(ctx context.Context, repo vcs.Repo, ref string) (vcs.Commit, error) {
	commit, resp, err := g.client.Repositories.GetCommit(ctx, repo.Owner(), repo.Name(), ref, nil)
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
