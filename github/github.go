/*
Package github interfaces with external github systems.
*/
package github

import (
	"context"
	"net/http"
	gourl "net/url"

	"github.com/google/go-github/v41/github"
)

// Client wraps the upstream github client
type Client struct {
	*github.Client
}

func NewClient(httpClient *http.Client) *Client {
	return &Client{
		Client: github.NewClient(httpClient),
	}
}

func NewEnterpriseClient(hostname string, httpClient *http.Client) (*Client, error) {
	client, err := github.NewEnterpriseClient(enterpriseBaseURL(hostname), enterpriseUploadURL(hostname), httpClient)
	if err != nil {
		return nil, err
	}

	return &Client{Client: client}, nil
}

func (c *Client) GetUser(ctx context.Context, name string) (*github.User, error) {
	user, _, err := c.Users.Get(ctx, name)
	return user, err
}

func (c *Client) ListOrganizations(ctx context.Context, name string) ([]*github.Organization, error) {
	orgs, _, err := c.Organizations.List(ctx, name, nil)
	return orgs, err
}

// Return a github enterprise URL from a hostname
func enterpriseBaseURL(hostname string) string   { return enterpriseURL(hostname, "/api/v3") }
func enterpriseUploadURL(hostname string) string { return enterpriseURL(hostname, "/api/uploads") }
func enterpriseURL(hostname, path string) string {
	return (&gourl.URL{Scheme: "https", Host: hostname, Path: path}).String()
}
