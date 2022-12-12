package github

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/google/go-github/v41/github"
	"github.com/leg100/otf"
	"golang.org/x/oauth2"
)

type Client struct {
	client *github.Client
}

func NewClient(ctx context.Context, cfg otf.CloudClientOptions) (*Client, error) {
	var err error
	var client *github.Client

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

func (g *Client) GetUser(ctx context.Context) (*otf.User, error) {
	guser, _, err := g.client.Users.Get(ctx, "")
	if err != nil {
		return nil, err
	}

	var orgs []*otf.Organization
	var teams []*otf.Team

	gorgs, _, err := g.client.Organizations.List(ctx, "", nil)
	if err != nil {
		return nil, err
	}
	for _, gorg := range gorgs {
		org, err := otf.NewOrganization(otf.OrganizationCreateOptions{
			Name: otf.String(gorg.GetLogin()),
		})
		if err != nil {
			return nil, err
		}
		orgs = append(orgs, org)

		// Determine if they are an admin; if so, add them to the owners team.
		membership, _, err := g.client.Organizations.GetOrgMembership(ctx, "", org.Name())
		if err != nil {
			return nil, err
		}
		if membership.GetRole() == "admin" {
			teams = append(teams, otf.NewTeam("owners", org))
		}
	}

	gteams, _, err := g.client.Teams.ListUserTeams(ctx, nil)
	if err != nil {
		return nil, err
	}
	for _, gteam := range gteams {
		org, err := otf.NewOrganization(otf.OrganizationCreateOptions{
			Name: otf.String(gteam.GetOrganization().GetLogin()),
		})
		if err != nil {
			return nil, err
		}
		teams = append(teams, otf.NewTeam(gteam.GetName(), org))
	}

	user := otf.NewUser(guser.GetLogin(), otf.WithOrganizationMemberships(orgs...), otf.WithTeamMemberships(teams...))
	return user, nil
}

func (g *Client) GetRepository(ctx context.Context, identifier string) (*otf.Repo, error) {
	owner, name, found := strings.Cut(identifier, "/")
	if !found {
		return nil, fmt.Errorf("malformed identifier: %s", identifier)
	}
	repo, _, err := g.client.Repositories.Get(ctx, owner, name)
	if err != nil {
		return nil, err
	}

	return &otf.Repo{
		Identifier: repo.GetFullName(),
		HTTPURL:    repo.GetURL(),
		Branch:     repo.GetDefaultBranch(),
	}, nil
}

func (g *Client) ListRepositories(ctx context.Context, opts otf.ListOptions) (*otf.RepoList, error) {
	repos, resp, err := g.client.Repositories.List(ctx, "", &github.RepositoryListOptions{
		ListOptions: github.ListOptions{
			Page:    opts.SanitizedPageNumber(),
			PerPage: opts.SanitizedPageSize(),
		},
	})
	if err != nil {
		return nil, err
	}

	// convert to common repo type before returning
	var items []*otf.Repo
	for _, repo := range repos {
		items = append(items, &otf.Repo{
			Identifier: repo.GetFullName(),
			HTTPURL:    repo.GetURL(),
			Branch:     repo.GetDefaultBranch(),
		})
	}

	return &otf.RepoList{
		Items:      items,
		Pagination: otf.NewPagination(opts, resp.LastPage*opts.SanitizedPageSize()),
	}, nil
}

func (g *Client) ListTags(ctx context.Context, opts otf.ListTagsOptions) ([]otf.TagRef, error) {
	owner, name, found := strings.Cut(opts.Identifier, "/")
	if !found {
		return nil, fmt.Errorf("malformed identifier: %s", opts.Identifier)
	}

	results, _, err := g.client.Git.ListMatchingRefs(ctx, owner, name, &github.ReferenceListOptions{
		Ref: "tags/" + opts.Prefix,
	})
	if err != nil {
		return nil, err
	}

	var tags []otf.TagRef
	for _, ref := range results {
		tags = append(tags, otf.TagRef(ref.GetRef()))
	}
	return tags, nil
}

func (g *Client) GetRepoTarball(ctx context.Context, topts otf.GetRepoTarballOptions) ([]byte, error) {
	owner, name, found := strings.Cut(topts.Identifier, "/")
	if !found {
		return nil, fmt.Errorf("malformed identifier: %s", topts.Identifier)
	}

	opts := github.RepositoryContentGetOptions{
		Ref: topts.Ref,
	}
	link, _, err := g.client.Repositories.GetArchiveLink(ctx, owner, name, github.Tarball, &opts, true)
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
	if err := otf.Unpack(resp.Body, untarpath); err != nil {
		return nil, err
	}
	contents, err := os.ReadDir(untarpath)
	if err != nil {
		return nil, err
	}
	if len(contents) != 1 {
		return nil, fmt.Errorf("malformed tarball archive")
	}
	parentDir := path.Join(untarpath, contents[0].Name())
	return otf.Pack(parentDir)
}

// CreateWebhook creates a webhook on a github repository.
func (g *Client) CreateWebhook(ctx context.Context, opts otf.CreateWebhookOptions) (string, error) {
	owner, name, found := strings.Cut(opts.Identifier, "/")
	if !found {
		return "", fmt.Errorf("malformed identifier: %s", opts.Identifier)
	}

	var events []string
	for _, event := range opts.Events {
		switch event {
		case otf.VCSPushEventType:
			events = append(events, "push")
		case otf.VCSPullEventType:
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
		Active: otf.Bool(true),
	})
	if err != nil {
		return "", err
	}
	return strconv.FormatInt(hook.GetID(), 10), nil
}

func (g *Client) UpdateWebhook(ctx context.Context, opts otf.UpdateWebhookOptions) error {
	owner, name, found := strings.Cut(opts.Identifier, "/")
	if !found {
		return fmt.Errorf("malformed identifier: %s", opts.Identifier)
	}

	intID, err := strconv.ParseInt(opts.ID, 10, 64)
	if err != nil {
		return err
	}

	var events []string
	for _, event := range opts.Events {
		switch event {
		case otf.VCSPushEventType:
			events = append(events, "push")
		case otf.VCSPullEventType:
			events = append(events, "pull_request")
		}
	}

	_, _, err = g.client.Repositories.EditHook(ctx, owner, name, intID, &github.Hook{
		Events: events,
		Config: map[string]any{
			"url":    opts.Endpoint,
			"secret": opts.Secret,
		},
		Active: otf.Bool(true),
	})
	if err != nil {
		return err
	}
	return nil
}

func (g *Client) GetWebhook(ctx context.Context, opts otf.GetWebhookOptions) (*otf.VCSWebhook, error) {
	owner, name, found := strings.Cut(opts.Identifier, "/")
	if !found {
		return nil, fmt.Errorf("malformed identifier: %s", opts.Identifier)
	}

	intID, err := strconv.ParseInt(opts.ID, 10, 64)
	if err != nil {
		return nil, err
	}

	hook, resp, err := g.client.Repositories.GetHook(ctx, owner, name, intID)
	if err != nil {
		if resp.StatusCode == http.StatusNotFound {
			return nil, otf.ErrResourceNotFound
		}
		return nil, err
	}

	var events []otf.VCSEventType
	for _, event := range hook.Events {
		switch event {
		case "push":
			events = append(events, otf.VCSPushEventType)
		case "pull_request":
			events = append(events, otf.VCSPullEventType)
		}
	}

	return &otf.VCSWebhook{
		ID:         strconv.FormatInt(hook.GetID(), 10),
		Identifier: opts.Identifier,
		Events:     events,
		Endpoint:   hook.GetURL(),
	}, nil
}

func (g *Client) DeleteWebhook(ctx context.Context, opts otf.DeleteWebhookOptions) error {
	owner, name, found := strings.Cut(opts.Identifier, "/")
	if !found {
		return fmt.Errorf("malformed identifier: %s", opts.Identifier)
	}

	intID, err := strconv.ParseInt(opts.ID, 10, 64)
	if err != nil {
		return err
	}

	_, err = g.client.Repositories.DeleteHook(ctx, owner, name, intID)
	return err
}

func (g *Client) SetStatus(ctx context.Context, opts otf.SetStatusOptions) error {
	owner, name, found := strings.Cut(opts.Identifier, "/")
	if !found {
		return fmt.Errorf("malformed identifier: %s", opts.Identifier)
	}

	var status string
	switch opts.Status {
	case otf.VCSPendingStatus, otf.VCSRunningStatus:
		status = "pending"
	case otf.VCSSuccessStatus:
		status = "success"
	case otf.VCSErrorStatus:
		status = "error"
	case otf.VCSFailureStatus:
		status = "failure"
	default:
		return fmt.Errorf("invalid vcs status: %s", opts.Status)
	}

	_, _, err := g.client.Repositories.CreateStatus(ctx, owner, name, opts.Ref, &github.RepoStatus{
		Context:     otf.String(fmt.Sprintf("otf/%s", opts.Workspace)),
		TargetURL:   otf.String(opts.TargetURL),
		Description: otf.String(opts.Description),
		State:       otf.String(status),
	})
	return err
}
