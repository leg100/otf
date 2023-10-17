package vcs

import (
	"context"
)

type (
	Client interface {
		// ListRepositories lists repositories accessible to the current user.
		ListRepositories(ctx context.Context, opts ListRepositoriesOptions) ([]string, error)
		GetRepository(ctx context.Context, identifier string) (Repository, error)
		// GetRepoTarball retrieves a .tar.gz tarball of a git repository
		GetRepoTarball(ctx context.Context, opts GetRepoTarballOptions) ([]byte, string, error)
		// CreateWebhook creates a webhook on the cloud provider, returning the
		// provider's unique ID for the webhook.
		CreateWebhook(ctx context.Context, opts CreateWebhookOptions) (string, error)
		UpdateWebhook(ctx context.Context, id string, opts UpdateWebhookOptions) error
		GetWebhook(ctx context.Context, opts GetWebhookOptions) (Webhook, error)
		DeleteWebhook(ctx context.Context, opts DeleteWebhookOptions) error
		SetStatus(ctx context.Context, opts SetStatusOptions) error
		// ListTags lists git tags on a repository. Each tag should be prefixed with
		// 'tags/'.
		ListTags(ctx context.Context, opts ListTagsOptions) ([]string, error)
		// ListPullRequestFiles returns the paths of files that are modified in the pull request
		ListPullRequestFiles(ctx context.Context, repo string, pull int) ([]string, error)
		// GetCommit retrieves commit from the repo with the given git ref
		GetCommit(ctx context.Context, repo, ref string) (Commit, error)
	}

	// NewTokenClientOptions are options for creating a client using a personal
	// access token (PAT).
	NewTokenClientOptions struct {
		Token               string
		Hostname            string
		SkipTLSVerification bool
	}

	GetRepoTarballOptions struct {
		Repo string  // repo identifier, <owner>/<repo>
		Ref  *string // branch/tag/SHA ref, nil means default branch
	}

	ListRepositoriesOptions struct {
		PageSize int
	}

	// ListTagsOptions are options for listing tags on a vcs repository
	ListTagsOptions struct {
		Repo   string // repo identifier, <owner>/<repo>
		Prefix string // only list tags that start with this string
	}

	// Webhook is a cloud's configuration for a webhook.
	Webhook struct {
		ID       string // cloud's webhook ID
		Repo     string // identifier is <repo_owner>/<repo_name>
		Events   []EventType
		Endpoint string // the OTF URL that receives events
	}

	CreateWebhookOptions struct {
		Repo     string // repo identifier, <owner>/<repo>
		Secret   string // secret string for generating signature
		Endpoint string // otf's external-facing host[:port]
		Events   []EventType
	}

	UpdateWebhookOptions CreateWebhookOptions

	// GetWebhookOptions are options for retrieving a webhook.
	GetWebhookOptions struct {
		Repo string // Repository identifier, <owner>/<repo>
		ID   string // vcs' webhook ID
	}

	// DeleteWebhookOptions are options for deleting a webhook.
	DeleteWebhookOptions struct {
		Repo string // Repository identifier, <owner>/<repo>
		ID   string // vcs' webhook ID
	}

	// SetStatusOptions are options for setting a status on a VCS repo
	SetStatusOptions struct {
		Workspace   string // workspace name
		Repo        string // <owner>/<repo>
		Ref         string // git ref
		Status      Status
		TargetURL   string
		Description string
	}

	Repository struct {
		Path          string
		DefaultBranch string
	}

	Commit struct {
		SHA    string
		URL    string
		Author CommitAuthor
	}

	CommitAuthor struct {
		Username   string
		ProfileURL string
		AvatarURL  string
	}
)
