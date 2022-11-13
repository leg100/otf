package otf

import (
	"context"

	oauth2github "golang.org/x/oauth2/github"
)

type testCloud struct {
	CloudConfigMixin
	user  *User
	repos []*Repo
}

type TestCloudOption func(*testCloud)

func WithUser(user *User) TestCloudOption {
	return func(cloud *testCloud) {
		cloud.user = user
	}
}

func WithHostname(hostname string) TestCloudOption {
	return func(cloud *testCloud) {
		cloud.hostname = hostname
	}
}

func WithRepos(repos ...*Repo) TestCloudOption {
	return func(cloud *testCloud) {
		cloud.repos = repos
	}
}

func NewTestCloudClient(opts ...TestCloudOption) *testCloud {
	cloud := &testCloud{
		CloudConfigMixin: CloudConfigMixin{
			cloudName:           "fake-cloud",
			endpoint:            oauth2github.Endpoint,
			skipTLSVerification: true,
			OAuthCredentials: &OAuthCredentials{
				clientID:     "abc-123",
				clientSecret: "xyz-789",
			},
		},
	}
	for _, o := range opts {
		o(cloud)
	}
	return cloud
}

func (f *testCloud) NewDirectoryClient(context.Context, CloudClientOptions) (CloudClient, error) {
	return &TestDirectoryClient{
		User:  f.user,
		Repos: f.repos,
	}, nil
}

type TestDirectoryClient struct {
	User  *User
	Repos []*Repo
	CloudClient
}

func (f *TestDirectoryClient) GetUser(context.Context) (*User, error) {
	return f.User, nil
}

func (f *TestDirectoryClient) ListRepositories(ctx context.Context, opts ListOptions) (*RepoList, error) {
	return &RepoList{
		Items:      f.Repos,
		Pagination: NewPagination(opts, len(f.Repos)),
	}, nil
}

func (f *TestDirectoryClient) GetRepository(context.Context, string) (*Repo, error) {
	return f.Repos[0], nil
}

func (f *TestDirectoryClient) GetRepoTarball(context.Context, *VCSRepo) ([]byte, error) {
	return nil, nil
}
