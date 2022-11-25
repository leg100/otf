package otf

import (
	"context"
	"errors"
	"net/url"

	"golang.org/x/oauth2"
)

var ErrOAuthCredentialsIncomplete = errors.New("must specify both client ID and client secret")

// CloudName uniquely identifies a cloud provider
type CloudName string

// Cloud is an external provider of various cloud services e.g. identity provider, VCS
// repositories etc.
type Cloud interface {
	NewClient(context.Context, ClientConfig) (CloudClient, error)
}

// ClientConfig is configuration for creating a new cloud client
type ClientConfig struct {
	Hostname            string
	SkipTLSVerification bool

	// one and only one token must be non-nil
	OAuthToken    *oauth2.Token
	PersonalToken *string
}

type CloudClient interface {
	GetUser(ctx context.Context) (*User, error)
	// ListRepositories lists repositories accessible to the current user.
	//
	// TODO: add optional filters
	ListRepositories(ctx context.Context, opts ListOptions) (*RepoList, error)
	GetRepository(ctx context.Context, identifier string) (*Repo, error)
	// GetRepoTarball retrieves a .tar.gz tarball of a git repository
	GetRepoTarball(ctx context.Context, opts GetRepoTarballOptions) ([]byte, error)

	// CreateWebhook creates a webhook on the cloud provider, subscribing to
	// VCS events to trigger runs.
	CreateWebhook(ctx context.Context, opts CreateWebhookOptions) (string, error)
	UpdateWebhook(ctx context.Context, opts UpdateWebhookOptions) error
	GetWebhook(ctx context.Context, opts GetWebhookOptions) (*VCSWebhook, error)
	DeleteWebhook(ctx context.Context, opts DeleteWebhookOptions) error

	SetStatus(ctx context.Context, opts SetStatusOptions) error
}

type VCSWebhook struct {
	ID         string // vcs' ID
	Identifier string // identifier is <repo_owner>/<repo_name>
	HTTPURL    string // HTTPURL is the web url for the repo
	Events     []VCSEventType
	Endpoint   string
}

type GetRepoTarballOptions struct {
	Identifier string // repo identifier, <owner>/<repo>
	Ref        string // branch/tag/SHA ref
}

type CreateWebhookOptions struct {
	Identifier string // repo identifier, <owner>/<repo>
	Secret     string // secret string for generating signature
	URL        string // external-facing host[:port]
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

type VCSStatus string

const (
	VCSPendingStatus VCSStatus = "pending"
	VCSRunningStatus VCSStatus = "running"
	VCSSuccessStatus VCSStatus = "success"
	VCSErrorStatus   VCSStatus = "error"
	VCSFailureStatus VCSStatus = "failure"
)

// CloudConfig is configuation for a cloud provider
type CloudConfig struct {
	Name                CloudName
	Hostname            string
	Cloud               Cloud
	SkipTLSVerification bool

	// OAuth config
	ClientID     string
	ClientSecret string
	Endpoint     oauth2.Endpoint
	Scopes       []string
}

func (cfg *CloudConfig) String() string {
	return string(cfg.Name)
}

func (cfg *CloudConfig) Validate() error {
	if cfg.ClientID == "" && cfg.ClientSecret != "" {
		return ErrOAuthCredentialsIncomplete
	}
	if cfg.ClientID != "" && cfg.ClientSecret == "" {
		return ErrOAuthCredentialsIncomplete
	}
	return nil
}

// UpdateEndpoint updates a cloud's OAuth endpoint to use the configured hostname
func (cfg *CloudConfig) UpdateEndpoint() (err error) {
	cfg.Endpoint.AuthURL, err = updateHost(cfg.Endpoint.AuthURL, cfg.Hostname)
	if err != nil {
		return err
	}
	cfg.Endpoint.TokenURL, err = updateHost(cfg.Endpoint.TokenURL, cfg.Hostname)
	if err != nil {
		return err
	}
	return nil
}

// Repo is a VCS repository belonging to a cloud
//
// TODO: remove or do something to this because there is too much overlap with
// VCSRepo
type Repo struct {
	// Identifier is <repo_owner>/<repo_name>
	Identifier string `schema:"identifier,required"`
	// HTTPURL is the web url for the repo
	HTTPURL string `schema:"http_url,required"`
	// Branch is the default master Branch for a repo
	Branch string `schema:"branch,required"`
}

func (r Repo) ID() string { return r.Identifier }

// RepoList is a paginated list of cloud repositories.
type RepoList struct {
	*Pagination
	Items []*Repo
}

// updateHost updates the hostname in a URL
func updateHost(u, host string) (string, error) {
	parsed, err := url.Parse(u)
	if err != nil {
		return "", err
	}

	parsed.Host = host

	return parsed.String(), nil
}
