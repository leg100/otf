package html

import (
	"context"

	"github.com/google/go-github/v41/github"
)

type fakeGithubClient struct {
	user *github.User
	orgs []*github.Organization
}

func (c *fakeGithubClient) GetUser(_ context.Context, _ string) (*github.User, error) {
	return c.user, nil
}

func (c *fakeGithubClient) ListOrganizations(_ context.Context, _ string) ([]*github.Organization, error) {
	return c.orgs, nil
}
