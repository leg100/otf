package otf

import (
	"context"

	"golang.org/x/oauth2"
)

type Cloud interface {
	NewDirectoryClient(context.Context, DirectoryClientOptions) (DirectoryClient, error)
	CloudConfig
}

type DirectoryClientOptions struct {
	OAuthToken    *oauth2.Token
	PersonalToken *string
}

type DirectoryClient interface {
	GetUser(ctx context.Context) (*User, error)
	ListRepositories(ctx context.Context, opts ListOptions) (*RepoList, error)
	GetRepoTarball(ctx context.Context, repo *VCSRepo) ([]byte, error)
}

// Repo is a VCS repository belonging to a cloud
type Repo struct {
	// Identifier is <repo_owner>/<repo_name>
	Identifier string `schema:"identifier,required"`
	// HttpURL is the web url for the repo
	HttpURL string `schema:"http_url,required"`
	// Branch is the default master Branch for a repo
	Branch string `schema:"branch,required"`
}

func (r Repo) ID() string { return r.Identifier }

// RepoList is a paginated list of cloud repositories.
type RepoList struct {
	*Pagination
	Items []*Repo
}