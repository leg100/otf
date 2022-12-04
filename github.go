package otf

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
	"golang.org/x/oauth2"
	oauth2github "golang.org/x/oauth2/github"
)

const (
	DefaultGithubHostname string = "github.com"
)

func GithubDefaults() CloudConfig {
	return CloudConfig{
		Name:     "github",
		Hostname: DefaultGithubHostname,
		Cloud:    &GithubCloud{},
	}
}

func GithubOAuthDefaults() *oauth2.Config {
	return &oauth2.Config{
		Endpoint: oauth2github.Endpoint,
		Scopes:   []string{"user:email", "read:org"},
	}
}

type GithubCloud struct{}

func (g *GithubCloud) NewClient(ctx context.Context, opts CloudClientOptions) (CloudClient, error) {
	return NewGithubClient(ctx, opts)
}

func (GithubCloud) HandleEvent(w http.ResponseWriter, r *http.Request, opts HandleEventOptions) *VCSEvent {
	return nil
}

type GithubClient struct {
	client *github.Client
}

func NewGithubClient(ctx context.Context, cfg CloudClientOptions) (*GithubClient, error) {
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
		client, err = NewGithubEnterpriseClient(cfg.Hostname, httpClient)
		if err != nil {
			return nil, err
		}
	} else {
		client = github.NewClient(httpClient)
	}
	return &GithubClient{client: client}, nil
}

func NewGithubEnterpriseClient(hostname string, httpClient *http.Client) (*github.Client, error) {
	return github.NewEnterpriseClient(
		"https://"+hostname,
		"https://"+hostname,
		httpClient)
}

func (g *GithubClient) GetUser(ctx context.Context) (*User, error) {
	guser, _, err := g.client.Users.Get(ctx, "")
	if err != nil {
		return nil, err
	}

	var orgs []*Organization
	var teams []*Team

	gorgs, _, err := g.client.Organizations.List(ctx, "", nil)
	if err != nil {
		return nil, err
	}
	for _, gorg := range gorgs {
		org, err := NewOrganization(OrganizationCreateOptions{
			Name: String(gorg.GetLogin()),
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
			teams = append(teams, NewTeam("owners", org))
		}
	}

	gteams, _, err := g.client.Teams.ListUserTeams(ctx, nil)
	if err != nil {
		return nil, err
	}
	for _, gteam := range gteams {
		org, err := NewOrganization(OrganizationCreateOptions{
			Name: String(gteam.GetOrganization().GetLogin()),
		})
		if err != nil {
			return nil, err
		}
		teams = append(teams, NewTeam(gteam.GetName(), org))
	}

	user := NewUser(guser.GetLogin(), WithOrganizationMemberships(orgs...), WithTeamMemberships(teams...))
	return user, nil
}

func (g *GithubClient) GetRepository(ctx context.Context, identifier string) (*Repo, error) {
	owner, name, found := strings.Cut(identifier, "/")
	if !found {
		return nil, fmt.Errorf("malformed identifier: %s", identifier)
	}
	repo, _, err := g.client.Repositories.Get(ctx, owner, name)
	if err != nil {
		return nil, err
	}

	return &Repo{
		Identifier: repo.GetFullName(),
		HTTPURL:    repo.GetURL(),
		Branch:     repo.GetDefaultBranch(),
	}, nil
}

func (g *GithubClient) ListRepositories(ctx context.Context, opts ListOptions) (*RepoList, error) {
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
	var items []*Repo
	for _, repo := range repos {
		items = append(items, &Repo{
			Identifier: repo.GetFullName(),
			HTTPURL:    repo.GetURL(),
			Branch:     repo.GetDefaultBranch(),
		})
	}

	return &RepoList{
		Items:      items,
		Pagination: NewPagination(opts, resp.LastPage*opts.SanitizedPageSize()),
	}, nil
}

func (g *GithubClient) GetRepoTarball(ctx context.Context, topts GetRepoTarballOptions) ([]byte, error) {
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
	if err := Unpack(resp.Body, untarpath); err != nil {
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
	return Pack(parentDir)
}

// CreateWebhook creates a webhook on a github repository.
func (g *GithubClient) CreateWebhook(ctx context.Context, opts CreateWebhookOptions) (string, error) {
	owner, name, found := strings.Cut(opts.Identifier, "/")
	if !found {
		return "", fmt.Errorf("malformed identifier: %s", opts.Identifier)
	}

	var events []string
	for _, event := range opts.Events {
		switch event {
		case VCSPushEventType:
			events = append(events, "push")
		case VCSPullEventType:
			events = append(events, "pull_request")
		}
	}

	hook, _, err := g.client.Repositories.CreateHook(ctx, owner, name, &github.Hook{
		Events: events,
		Config: map[string]any{
			"url":    opts.Endpoint,
			"secret": opts.Secret,
		},
		Active: Bool(true),
	})
	if err != nil {
		return "", err
	}
	return strconv.FormatInt(hook.GetID(), 10), nil
}

func (g *GithubClient) UpdateWebhook(ctx context.Context, opts UpdateWebhookOptions) error {
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
		case VCSPushEventType:
			events = append(events, "push")
		case VCSPullEventType:
			events = append(events, "pull_request")
		}
	}

	_, _, err = g.client.Repositories.EditHook(ctx, owner, name, intID, &github.Hook{
		Events: events,
		Config: map[string]any{
			"url":    opts.Endpoint,
			"secret": opts.Secret,
		},
		Active: Bool(true),
	})
	if err != nil {
		return err
	}
	return nil
}

func (g *GithubClient) GetWebhook(ctx context.Context, opts GetWebhookOptions) (*VCSWebhook, error) {
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
			return nil, ErrResourceNotFound
		}
		return nil, err
	}

	var events []VCSEventType
	for _, event := range hook.Events {
		switch event {
		case "push":
			events = append(events, VCSPushEventType)
		case "pull_request":
			events = append(events, VCSPullEventType)
		}
	}

	return &VCSWebhook{
		ID:         strconv.FormatInt(hook.GetID(), 10),
		Identifier: opts.Identifier,
		Events:     events,
		Endpoint:   hook.GetURL(),
	}, nil
}

func (g *GithubClient) DeleteWebhook(ctx context.Context, opts DeleteWebhookOptions) error {
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

func (g *GithubClient) SetStatus(ctx context.Context, opts SetStatusOptions) error {
	owner, name, found := strings.Cut(opts.Identifier, "/")
	if !found {
		return fmt.Errorf("malformed identifier: %s", opts.Identifier)
	}

	var status string
	switch opts.Status {
	case VCSPendingStatus, VCSRunningStatus:
		status = "pending"
	case VCSSuccessStatus:
		status = "success"
	case VCSErrorStatus:
		status = "error"
	case VCSFailureStatus:
		status = "failure"
	default:
		return fmt.Errorf("invalid vcs status: %s", opts.Status)
	}

	_, _, err := g.client.Repositories.CreateStatus(ctx, owner, name, opts.Ref, &github.RepoStatus{
		Context:     String(fmt.Sprintf("otf/%s", opts.Workspace)),
		TargetURL:   String(opts.TargetURL),
		Description: String(opts.Description),
		State:       String(status),
	})
	return err
}
