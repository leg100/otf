package gitlab

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path"
	"slices"
	"strconv"
	"strings"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/authenticator"
	otfhttp "github.com/leg100/otf/internal/http"
	"github.com/leg100/otf/internal/user"
	"github.com/leg100/otf/internal/vcs"
	gitlab "gitlab.com/gitlab-org/api/client-go"
	"golang.org/x/oauth2"
)

type (
	Client struct {
		client *gitlab.Client
	}

	ClientOptions struct {
		Hostname            string
		SkipTLSVerification bool

		OAuthToken    *oauth2.Token
		PersonalToken *string
	}
)

func NewClient(cfg ClientOptions) (*Client, error) {
	var (
		client  *gitlab.Client
		err     error
		options = []gitlab.ClientOptionFunc{
			gitlab.WithBaseURL(
				(&url.URL{Scheme: "https", Host: cfg.Hostname}).String(),
			),
		}
	)
	if cfg.SkipTLSVerification {
		options = append(options, gitlab.WithHTTPClient(
			&http.Client{Transport: otfhttp.InsecureTransport},
		))
	}
	if cfg.OAuthToken != nil {
		client, err = gitlab.NewOAuthClient(cfg.OAuthToken.AccessToken, options...)
	} else if cfg.PersonalToken != nil {
		client, err = gitlab.NewClient(*cfg.PersonalToken, options...)
	} else {
		// anonymous client
		client, err = gitlab.NewClient("", options...)
	}
	if err != nil {
		return nil, err
	}
	return &Client{client: client}, nil
}

func NewTokenClient(opts vcs.NewTokenClientOptions) (vcs.Client, error) {
	return NewClient(ClientOptions{
		Hostname:            opts.Hostname,
		PersonalToken:       &opts.Token,
		SkipTLSVerification: opts.SkipTLSVerification,
	})
}

func NewOAuthClient(cfg authenticator.OAuthConfig, token *oauth2.Token) (authenticator.IdentityProviderClient, error) {
	return NewClient(ClientOptions{
		Hostname:            cfg.Hostname,
		OAuthToken:          token,
		SkipTLSVerification: cfg.SkipTLSVerification,
	})
}

func (g *Client) GetCurrentUser(ctx context.Context) (authenticator.UserInfo, error) {
	guser, _, err := g.client.Users.CurrentUser()
	if err != nil {
		return authenticator.UserInfo{}, err
	}
	username, err := user.NewUsername(guser.Username)
	if err != nil {
		return authenticator.UserInfo{}, err
	}
	return authenticator.UserInfo{
		Username:  username,
		AvatarURL: &guser.AvatarURL,
	}, nil
}

func (g *Client) GetDefaultBranch(ctx context.Context, identifier string) (string, error) {
	proj, _, err := g.client.Projects.GetProject(identifier, nil)
	if err != nil {
		return "", err
	}

	return proj.DefaultBranch, nil
}

func (g *Client) ListRepositories(ctx context.Context, lopts vcs.ListRepositoriesOptions) ([]vcs.Repo, error) {
	opts := &gitlab.ListProjectsOptions{
		ListOptions: gitlab.ListOptions{
			PerPage: lopts.PageSize,
		},
		// limit results to those repos the authenticated user is a member of,
		// otherwise we'll get *all* accessible repos, public and private.
		Membership: internal.Ptr(true),
	}
	projects, _, err := g.client.Projects.ListProjects(opts, nil)
	if err != nil {
		return nil, err
	}

	repos := make([]vcs.Repo, len(projects))
	for i, proj := range projects {
		repos[i] = vcs.NewMustRepo(proj.Namespace.Path, proj.Path)
	}
	return repos, nil
}

func (g *Client) ListTags(ctx context.Context, opts vcs.ListTagsOptions) ([]string, error) {
	results, _, err := g.client.Tags.ListTags(opts.Repo, &gitlab.ListTagsOptions{
		Search: internal.Ptr("^" + opts.Prefix),
	})
	if err != nil {
		return nil, err
	}

	tags := make([]string, len(results))
	for i, ref := range results {
		tags[i] = fmt.Sprintf("tags/%s", ref.Name)
	}
	return tags, nil
}

func (g *Client) GetRepoTarball(ctx context.Context, opts vcs.GetRepoTarballOptions) ([]byte, string, error) {
	tarball, _, err := g.client.Repositories.Archive(opts.Repo.String(), &gitlab.ArchiveOptions{
		Format: internal.Ptr("tar.gz"),
		SHA:    opts.Ref,
	})
	if err != nil {
		return nil, "", err
	}

	// Gitlab tarball contents are contained within a top-level directory
	// formatted <ref>-<sha>. We want the tarball without this directory,
	// so we re-tar the contents without the top-level directory.
	untarpath, err := os.MkdirTemp("", fmt.Sprintf("gitlab-%s-%s-*", opts.Repo.Owner(), opts.Repo.Name()))
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
	if len(parts) < 2 {
		return nil, "", fmt.Errorf("malformed directory name found in tarball: %s", dir)
	}
	tarball, err = internal.Pack(path.Join(untarpath, dir))
	if err != nil {
		return nil, "", err
	}
	return tarball, parts[len(parts)-1], nil
}

func (g *Client) CreateWebhook(ctx context.Context, opts vcs.CreateWebhookOptions) (string, error) {
	addOpts := &gitlab.AddProjectHookOptions{
		EnableSSLVerification: internal.Ptr(true),
		PushEvents:            internal.Ptr(true),
		Token:                 internal.Ptr(opts.Secret),
		URL:                   internal.Ptr(opts.Endpoint),
	}
	for _, event := range opts.Events {
		switch event {
		case vcs.EventTypePush:
			addOpts.PushEvents = internal.Ptr(true)
		case vcs.EventTypePull:
			addOpts.MergeRequestsEvents = internal.Ptr(true)
		}
	}

	hook, _, err := g.client.Projects.AddProjectHook(opts.Repo.String(), addOpts)
	if err != nil {
		return "", err
	}
	return strconv.Itoa(hook.ID), nil
}

func (g *Client) UpdateWebhook(ctx context.Context, id string, opts vcs.UpdateWebhookOptions) error {
	intID, err := strconv.Atoi(id)
	if err != nil {
		return err
	}

	editOpts := &gitlab.EditProjectHookOptions{
		EnableSSLVerification: internal.Ptr(true),
		Token:                 internal.Ptr(opts.Secret),
		URL:                   internal.Ptr(opts.Endpoint),
	}
	for _, event := range opts.Events {
		switch event {
		case vcs.EventTypePush:
			editOpts.PushEvents = internal.Ptr(true)
		case vcs.EventTypePull:
			editOpts.MergeRequestsEvents = internal.Ptr(true)
		}
	}

	_, _, err = g.client.Projects.EditProjectHook(opts.Repo.String(), intID, editOpts)
	if err != nil {
		return err
	}
	return nil
}

func (g *Client) GetWebhook(ctx context.Context, opts vcs.GetWebhookOptions) (vcs.Webhook, error) {
	id, err := strconv.Atoi(opts.ID)
	if err != nil {
		return vcs.Webhook{}, err
	}

	hook, resp, err := g.client.Projects.GetProjectHook(opts.Repo.String(), id)
	if err != nil {
		if resp.StatusCode == http.StatusNotFound {
			return vcs.Webhook{}, internal.ErrResourceNotFound
		}
		return vcs.Webhook{}, err
	}

	var events []vcs.EventType
	if hook.PushEvents {
		events = append(events, vcs.EventTypePush)
	}
	if hook.MergeRequestsEvents {
		events = append(events, vcs.EventTypePull)
	}

	return vcs.Webhook{
		ID:       strconv.Itoa(id),
		Repo:     opts.Repo,
		Events:   events,
		Endpoint: hook.URL,
	}, nil
}

func (g *Client) DeleteWebhook(ctx context.Context, opts vcs.DeleteWebhookOptions) error {
	id, err := strconv.Atoi(opts.ID)
	if err != nil {
		return err
	}

	_, err = g.client.Projects.DeleteProjectHook(opts.Repo.String(), id)
	return err
}

func (g *Client) SetStatus(ctx context.Context, opts vcs.SetStatusOptions) error {
	// unsupported
	return nil
}

func (g *Client) ListPullRequestFiles(ctx context.Context, repo vcs.Repo, pull int) ([]string, error) {
	diffs, _, err := g.client.MergeRequests.ListMergeRequestDiffs(repo.String(), pull, &gitlab.ListMergeRequestDiffsOptions{})
	if err != nil {
		return nil, err
	}
	var changed []string
	for _, diff := range diffs {
		changed = append(changed, diff.OldPath)
		changed = append(changed, diff.NewPath)
	}
	// remove duplicates
	slices.Sort(changed)
	return slices.Compact(changed), nil
}

func (g *Client) GetCommit(ctx context.Context, repo vcs.Repo, ref string) (vcs.Commit, error) {
	commit, _, err := g.client.Commits.GetCommit(repo.String(), ref, &gitlab.GetCommitOptions{})
	if err != nil {
		return vcs.Commit{}, err
	}
	return vcs.Commit{
		SHA: commit.ID,
		URL: commit.WebURL,
		Author: vcs.CommitAuthor{
			Username: commit.AuthorName,
		},
	}, nil
}
