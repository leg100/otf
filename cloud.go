package otf

import (
	"context"

	"golang.org/x/oauth2"
)

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
