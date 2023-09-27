package bitbucketserver

import (
	"bytes"
	"context"
	"fmt"
	"math/rand"
	"mime"
	"net/url"
	"strings"

	bitbucketapi "github.com/gfleury/go-bitbucket-v1"
	"github.com/leg100/otf/internal/vcs"
	"github.com/mitchellh/mapstructure"
	"golang.org/x/exp/slog"
)

type (
	Client struct {
		client  *bitbucketapi.APIClient
		baseURL string
	}

	ClientOptions struct {
		PersonalToken       *string
		Hostname            string
		SkipTLSVerification bool
	}
)

func NewClient(cfg ClientOptions) (*Client, error) {
	if cfg.PersonalToken == nil {
		return nil, fmt.Errorf("no credentials provided")
	}

	baseURL := (&url.URL{Scheme: "https", Host: cfg.Hostname}).String()

	ctx := context.Background()
	ctx = context.WithValue(ctx, bitbucketapi.ContextAccessToken, *cfg.PersonalToken)

	client := bitbucketapi.NewAPIClient(
		ctx,
		bitbucketapi.NewConfiguration(
			baseURL+"/rest",
		),
	)

	return &Client{client: client, baseURL: baseURL}, nil
}

func NewTokenClient(opts vcs.NewTokenClientOptions) (vcs.Client, error) {
	return NewClient(ClientOptions{
		PersonalToken:       &opts.Token,
		Hostname:            opts.Hostname,
		SkipTLSVerification: opts.SkipTLSVerification,
	})
}

func (g *Client) GetCurrentUser(ctx context.Context) (string, error) {
	response, err := g.client.DefaultApi.GetProjects(map[string]interface{}{
		"start": 0,
		"limit": 1,
	})
	if err != nil {
		return "", fmt.Errorf("failed to get bitbucket projects: %w", err)
	}

	username := response.Header.Get("X-AUSERNAME")
	if username == "" {
		return "", fmt.Errorf("no username header on api response")
	}

	return username, nil
}

func (g *Client) GetRepository(ctx context.Context, identifier string) (vcs.Repository, error) {
	owner, name, found := strings.Cut(identifier, "/")
	if !found {
		return vcs.Repository{}, fmt.Errorf("malformed identifier: %s", identifier)
	}

	response, err := g.client.DefaultApi.GetRepository(owner, name)
	if err != nil {
		return vcs.Repository{}, fmt.Errorf("failed to retrieve repository: %w", err)
	}

	repo, err := bitbucketapi.GetRepositoryResponse(response)
	if err != nil {
		return vcs.Repository{}, fmt.Errorf("failed to unmarshal response into repository: %w", err)
	}

	defaultBranchResponse, err := g.client.DefaultApi.GetDefaultBranch(owner, name)
	if err != nil {
		return vcs.Repository{}, fmt.Errorf("failed to retrieve default branch: %w", err)
	}

	defaultBranch, err := bitbucketapi.GetBranchResponse(defaultBranchResponse)
	if err != nil {
		return vcs.Repository{}, fmt.Errorf("failed to unmarshal response into branch: %w", err)
	}

	return vcs.Repository{
		Path:          repo.Slug,
		DefaultBranch: defaultBranch.ID,
	}, nil
}

func (g *Client) ListRepositories(ctx context.Context, lopts vcs.ListRepositoriesOptions) ([]string, error) {
	// here we are returning an empty array. Looping through and returning
	// everything is super slow.
	return []string{}, nil

	// response, err := g.client.DefaultApi.GetProjects(map[string]interface{}{
	// 	"start": 0,
	// 	"limit": 1000,
	// })
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to get bitbucket projects: %w", err)
	// }

	// projectsResponse, err := bitbucketapi.GetProjectsResponse(response)
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to unmarshal projects: %w", err)
	// }

	// var repos []string
	// for _, project := range projectsResponse {
	// 	response, err := g.client.DefaultApi.GetRepositories(project.Key)
	// 	if err != nil {
	// 		return nil, fmt.Errorf("failed to get bitbucket repositories: %w", err)
	// 	}

	// 	repositoriesResponse, err := bitbucketapi.GetRepositoriesResponse(response)
	// 	if err != nil {
	// 		return nil, fmt.Errorf("failed to unmarshal repositories: %w", err)
	// 	}

	// 	for _, repository := range repositoriesResponse {
	// 		repos = append(repos, project.Key+"/"+repository.Slug)
	// 	}
	// }

	// return repos, nil
}

func (g *Client) ListTags(ctx context.Context, opts vcs.ListTagsOptions) ([]string, error) {
	owner, name, found := strings.Cut(opts.Repo, "/")
	if !found {
		return nil, fmt.Errorf("malformed identifier: %s", opts.Repo)
	}

	response, err := g.client.DefaultApi.GetTags(
		owner,
		name,
		map[string]interface{}{"filterText": opts.Prefix},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get tags: %w", err)
	}

	tagsResponse, err := bitbucketapi.GetTagsResponse(response)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response into tags: %w", err)
	}

	// return tags with the format 'tags/<tag_value>'
	var tags []string
	for _, ref := range tagsResponse {
		tags = append(tags, strings.TrimPrefix(ref.ID, "refs/"))
	}

	return tags, nil
}

func (g *Client) GetRepoTarball(ctx context.Context, opts vcs.GetRepoTarballOptions) ([]byte, string, error) {
	owner, name, found := strings.Cut(opts.Repo, "/")
	if !found {
		return nil, "", fmt.Errorf("malformed identifier: %s", opts.Repo)
	}

	tarball := bytes.NewBuffer(nil)

	r, err := g.client.DefaultApi.GetArchive(
		owner,
		name,
		map[string]interface{}{
			"format": "tar.gz",
			"at":     *opts.Ref,
		},
		tarball,
	)
	if err != nil {
		return nil, "", fmt.Errorf("failed to get archive: %w", err)
	}

	rawDisposition := r.Header.Get("Content-Disposition")

	_, params, err := mime.ParseMediaType(rawDisposition)
	if err != nil {
		return nil, "", fmt.Errorf("failed to parse content-disposition mimetype")
	}
	filename := params["filename"]

	// some-repo-main@bbd03249b1c.tar.gz
	ref := strings.TrimSuffix(filename, ".tar.gz")
	ref = ref[strings.LastIndex(filename, "@")+1:]

	return tarball.Bytes(), ref, nil
}

func (g *Client) CreateWebhook(ctx context.Context, opts vcs.CreateWebhookOptions) (string, error) {
	owner, name, found := strings.Cut(opts.Repo, "/")
	if !found {
		return "", fmt.Errorf("malformed identifier: %s", opts.Repo)
	}

	hookName := fmt.Sprintf("otf-%d", rand.Int())

	_, err := g.client.DefaultApi.CreateWebhook(owner, name, bitbucketapi.Webhook{
		Name: hookName,
		Events: []string{
			eventPush,
			eventPullRequestOpened,
			eventPullRequestSourceBranchUpdated,
			eventPullRequestModified,
			eventPullRequestMerged,
			eventPullRequestDeclined,
			eventPullRequestDeleted,
		},
		Url:    opts.Endpoint,
		Active: true,
		Configuration: bitbucketapi.WebhookConfiguration{
			Secret: opts.Secret,
		},
	}, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create webhook: %w", err)
	}

	return hookName, nil
}

func (g *Client) UpdateWebhook(ctx context.Context, id string, opts vcs.UpdateWebhookOptions) error {
	owner, name, found := strings.Cut(opts.Repo, "/")
	if !found {
		return fmt.Errorf("malformed identifier: %s", opts.Repo)
	}

	response, err := g.client.DefaultApi.FindWebhooks(owner, name, nil)
	if err != nil {
		return fmt.Errorf("failed to get webhook from bitbucket server api: %w", err)
	}

	webhooks, err := bitbucketapi.GetWebhooksResponse(response)
	if err != nil {
		return fmt.Errorf("failed to unmarshal webhook: %w", err)
	}

	var webhook bitbucketapi.Webhook
	var foundWebhook bool

	for _, w := range webhooks {
		if w.Name == id {
			webhook = w
			foundWebhook = true
			break
		}
	}
	if !foundWebhook {
		return fmt.Errorf("failed to find a webhook with the name of: %s", id)
	}

	g.client.DefaultApi.UpdateWebhook(owner, name, int32(webhook.ID), bitbucketapi.Webhook{
		Name: webhook.Name,
		Events: []string{
			eventPush,
			eventPullRequestOpened,
			eventPullRequestSourceBranchUpdated,
			eventPullRequestModified,
			eventPullRequestMerged,
			eventPullRequestDeclined,
			eventPullRequestDeleted,
		},
		Url:    opts.Endpoint,
		Active: true,
		Configuration: bitbucketapi.WebhookConfiguration{
			Secret: opts.Secret,
		},
	}, nil)
	return nil
}

const (
	eventPush = "repo:refs_changed"

	eventPullRequestOpened              = "pr:opened"
	eventPullRequestSourceBranchUpdated = "pr:from_ref_updated"
	eventPullRequestModified            = "pr:modified"
	eventPullRequestMerged              = "pr:merged"
	eventPullRequestDeclined            = "pr:declined"
	eventPullRequestDeleted             = "pr:deleted"
)

func (g *Client) GetWebhook(ctx context.Context, opts vcs.GetWebhookOptions) (vcs.Webhook, error) {
	owner, name, found := strings.Cut(opts.Repo, "/")
	if !found {
		return vcs.Webhook{}, fmt.Errorf("malformed identifier: %s", opts.Repo)
	}

	response, err := g.client.DefaultApi.FindWebhooks(owner, name, nil)
	if err != nil {
		return vcs.Webhook{}, fmt.Errorf("failed to get webhook from bitbucket server api: %w", err)
	}

	webhooks, err := bitbucketapi.GetWebhooksResponse(response)
	if err != nil {
		return vcs.Webhook{}, fmt.Errorf("failed to unmarshal webhook: %w", err)
	}

	var webhook bitbucketapi.Webhook
	var foundWebhook bool

	for _, w := range webhooks {
		if w.Name == opts.ID {
			webhook = w
			foundWebhook = true
			break
		}
	}
	if !foundWebhook {
		return vcs.Webhook{}, fmt.Errorf("failed to find a webhook with the name of: %s", opts.ID)
	}

	var events []vcs.EventType
	for _, event := range webhook.Events {
		switch event {
		case eventPush:
			events = append(events, vcs.EventTypePush)

		case eventPullRequestOpened:
			fallthrough
		case eventPullRequestSourceBranchUpdated:
			fallthrough
		case eventPullRequestModified:
			fallthrough
		case eventPullRequestMerged:
			fallthrough
		case eventPullRequestDeclined:
			fallthrough
		case eventPullRequestDeleted:
			events = append(events, vcs.EventTypePull)
		}
	}

	return vcs.Webhook{
		ID:       webhook.Name,
		Repo:     opts.Repo,
		Events:   events,
		Endpoint: webhook.Url,
	}, nil
}

func (g *Client) DeleteWebhook(ctx context.Context, opts vcs.DeleteWebhookOptions) error {
	owner, name, found := strings.Cut(opts.Repo, "/")
	if !found {
		return fmt.Errorf("malformed identifier: %s", opts.Repo)
	}

	response, err := g.client.DefaultApi.FindWebhooks(owner, name, nil)
	if err != nil {
		return fmt.Errorf("failed to get webhook from bitbucket server api: %w", err)
	}

	webhooks, err := bitbucketapi.GetWebhooksResponse(response)
	if err != nil {
		return fmt.Errorf("failed to unmarshal webhook: %w", err)
	}

	var webhook bitbucketapi.Webhook
	var foundWebhook bool

	for _, w := range webhooks {
		if w.Name == opts.ID {
			webhook = w
			foundWebhook = true
			break
		}
	}
	if !foundWebhook {
		return fmt.Errorf("failed to find a webhook with the name of: %s", opts.ID)
	}

	_, err = g.client.DefaultApi.DeleteWebhook(owner, name, int32(webhook.ID))
	if err != nil {
		return fmt.Errorf("failed to delete webhook: %w", err)
	}

	return nil
}

func (g *Client) SetStatus(ctx context.Context, opts vcs.SetStatusOptions) error {
	slog.Info("calling SetStatus")

	return nil
}

func (g *Client) ListPullRequestFiles(ctx context.Context, repo string, pull int) ([]string, error) {
	slog.Info("calling ListPullRequestFiles")
	return nil, nil
}

// getCommitsResponse cast Commits into structure
func getCommitResponse(r *bitbucketapi.APIResponse) (bitbucketapi.Commit, error) {
	var m bitbucketapi.Commit
	err := mapstructure.Decode(r.Values["values"], &m)
	return m, err
}

func (g *Client) getAvatarURL(slug string) (string, error) {
	avatarURL, err := url.JoinPath(g.baseURL, "/users/", url.PathEscape(slug), "/avatar.png?s=64")
	if err != nil {
		return "", fmt.Errorf("failed build avatar url (%s): %w", url.PathEscape(slug), err)
	}

	return avatarURL, nil
}

func (g *Client) getProfileURL(slug string) (string, error) {
	avatarURL, err := url.JoinPath(g.baseURL, "/users/", url.PathEscape(slug))
	if err != nil {
		return "", fmt.Errorf("failed build profile url (%s): %w", url.PathEscape(slug), err)
	}

	return avatarURL, nil
}

func (g *Client) GetCommit(ctx context.Context, repo, ref string) (vcs.Commit, error) {
	owner, name, found := strings.Cut(repo, "/")
	if !found {
		return vcs.Commit{}, fmt.Errorf("malformed identifier: %s", repo)
	}

	response, err := g.client.DefaultApi.GetCommit(owner, name, ref, nil)
	if err != nil {
		return vcs.Commit{}, fmt.Errorf("failed to get commit: %w", err)
	}

	commitResponse, err := getCommitResponse(response)
	if err != nil {
		return vcs.Commit{}, fmt.Errorf("failed to unmarshal commit: %w", err)
	}

	avatarURL, err := g.getAvatarURL(commitResponse.Author.Slug)
	if err != nil {
		return vcs.Commit{}, fmt.Errorf("failed to get avatar url: %w", err)
	}

	profileURL, err := g.getProfileURL(commitResponse.Author.Slug)
	if err != nil {
		return vcs.Commit{}, fmt.Errorf("failed to get profile url: %w", err)
	}

	return vcs.Commit{
		SHA: commitResponse.ID,
		Author: vcs.CommitAuthor{
			Username:   commitResponse.Author.Name,
			AvatarURL:  avatarURL,
			ProfileURL: profileURL,
		},
	}, nil
}
