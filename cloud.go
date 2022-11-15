package otf

import (
	"context"
	"errors"

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
	GetRepoTarball(ctx context.Context, repo *VCSRepo) ([]byte, error)
}

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
	cfg.Endpoint.AuthURL, err = UpdateHost(cfg.Endpoint.AuthURL, cfg.Hostname)
	if err != nil {
		return err
	}
	cfg.Endpoint.TokenURL, err = UpdateHost(cfg.Endpoint.TokenURL, cfg.Hostname)
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
