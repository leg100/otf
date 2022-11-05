package otf

import (
	"context"

	oauth2github "golang.org/x/oauth2/github"
)

type testCloud struct {
	cloudConfig
	user *User
}

func NewTestCloud(hostname string, user *User) *testCloud {
	return &testCloud{
		cloudConfig: cloudConfig{
			cloudName:           "fake",
			endpoint:            oauth2github.Endpoint,
			hostname:            hostname,
			skipTLSVerification: true,
			OAuthCredentials: &OAuthCredentials{
				clientID:     "abc-123",
				clientSecret: "xyz-789",
			},
		},
		user: user,
	}
}

func (f *testCloud) NewDirectoryClient(context.Context, DirectoryClientOptions) (DirectoryClient, error) {
	return &TestDirectoryClient{f.user}, nil
}

func (f *testCloud) NewCloud() (Cloud, error) { return nil, nil }

type TestDirectoryClient struct {
	User *User
}

func (f *TestDirectoryClient) GetUser(context.Context) (*User, error) {
	return f.User, nil
}

func (f *TestDirectoryClient) ListRepositories(context.Context) ([]*Repo, error) {
	return nil, nil
}
