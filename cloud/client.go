package cloud

import (
	"context"
)

type Client interface {
	GetUser(ctx context.Context) (*User, error)
	// ListRepositories lists repositories accessible to the current user.
	ListRepositories(ctx context.Context, opts ListRepositoriesOptions) ([]Repo, error)
	GetRepository(ctx context.Context, identifier string) (Repo, error)
	// GetRepoTarball retrieves a .tar.gz tarball of a git repository
	GetRepoTarball(ctx context.Context, opts GetRepoTarballOptions) ([]byte, error)
	// CreateWebhook creates a webhook on the cloud provider, returning the
	// provider's unique ID for the webhook.
	CreateWebhook(ctx context.Context, opts CreateWebhookOptions) (string, error)
	UpdateWebhook(ctx context.Context, opts UpdateWebhookOptions) error
	GetWebhook(ctx context.Context, opts GetWebhookOptions) (Webhook, error)
	DeleteWebhook(ctx context.Context, opts DeleteWebhookOptions) error
	SetStatus(ctx context.Context, opts SetStatusOptions) error
	// ListTags lists git tags on a repository. Each tag should be prefixed with
	// 'tags/'.
	ListTags(ctx context.Context, opts ListTagsOptions) ([]string, error)
}

// ClientOptions are options for constructing a cloud client
type ClientOptions struct {
	Hostname            string
	SkipTLSVerification bool

	Credentials
}

type GetRepoTarballOptions struct {
	Identifier string // repo identifier, <owner>/<repo>
	Ref        string // branch/tag/SHA ref
}

type ListRepositoriesOptions struct {
	PageSize int
}

// ListTagsOptions are options for listing tags on a vcs repository
type ListTagsOptions struct {
	Identifier string // repo identifier, <owner>/<repo>
	Prefix     string // only list tags that start with this string
}

// Webhook is a cloud's configuration for a webhook on OTF.
type Webhook struct {
	ID         string // cloud's webhook ID
	Identifier string // identifier is <repo_owner>/<repo_name>
	Events     []VCSEventType
	Endpoint   string // the OTF URL that receives events
}

type CreateWebhookOptions struct {
	Identifier string // repo identifier, <owner>/<repo>
	Secret     string // secret string for generating signature
	Endpoint   string // otf's external-facing host[:port]
	Events     []VCSEventType
}

type UpdateWebhookOptions struct {
	ID string // vcs' webhook ID

	CreateWebhookOptions
}

// GetWebhookOptions are options for retrieving a webhook.
type GetWebhookOptions struct {
	Identifier string // Repository identifier, <owner>/<repo>
	ID         string // vcs' webhook ID
}

// DeleteWebhookOptions are options for deleting a webhook.
type DeleteWebhookOptions struct {
	Identifier string // Repository identifier, <owner>/<repo>
	ID         string // vcs' webhook ID
}

// SetStatusOptions are options for setting a status on a VCS repo
type SetStatusOptions struct {
	Workspace   string
	Identifier  string // <owner>/<repo>
	Ref         string // git ref
	Status      VCSStatus
	TargetURL   string
	Description string
}
