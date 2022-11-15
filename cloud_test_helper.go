package otf

import (
	"context"
)

type testCloud struct {
	user  *User
	repos []*Repo
}

func NewTestCloud(opts ...TestCloudOption) *testCloud {
	cloud := testCloud{}
	for _, o := range opts {
		o(&cloud)
	}
	return &cloud
}

type TestCloudOption func(*testCloud)

func WithUser(user *User) TestCloudOption {
	return func(cloud *testCloud) {
		cloud.user = user
	}
}

func WithRepos(repos ...*Repo) TestCloudOption {
	return func(cloud *testCloud) {
		cloud.repos = repos
	}
}

func (f *testCloud) NewClient(context.Context, ClientConfig) (CloudClient, error) {
	return &TestClient{
		User:  f.user,
		Repos: f.repos,
	}, nil
}

type TestClient struct {
	User  *User
	Repos []*Repo
	CloudClient
}

func (f *TestClient) GetUser(context.Context) (*User, error) {
	return f.User, nil
}

func (f *TestClient) ListRepositories(ctx context.Context, opts ListOptions) (*RepoList, error) {
	return &RepoList{
		Items:      f.Repos,
		Pagination: NewPagination(opts, len(f.Repos)),
	}, nil
}

func (f *TestClient) GetRepository(context.Context, string) (*Repo, error) {
	return f.Repos[0], nil
}

func (f *TestClient) GetRepoTarball(context.Context, *VCSRepo) ([]byte, error) {
	return nil, nil
}
