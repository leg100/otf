package github

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"

	"github.com/google/go-github/v41/github"
	"golang.org/x/oauth2"
)

const DefaultGithubHostname = "github.com"

type Cloud struct {
	*GithubConfig
}

func NewGithubCloud(opts *cloudConfigOptions) *Cloud {
	cloud := &Cloud{defaultGithubConfig()}
	cloud.override(opts)
	return cloud
}

func (g *Cloud) NewClient(ctx context.Context, opts otf.CloudClientOptions) (CloudClient, error) {
	var err error
	var client *github.Client

	// Optionally skip TLS verification of github API
	if g.skipTLSVerification {
		ctx = context.WithValue(ctx, oauth2.HTTPClient, &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		})
	}

	// Github's oauth access token never expires
	var src oauth2.TokenSource
	if opts.OAuthToken != nil {
		src = oauth2.StaticTokenSource(opts.OAuthToken)
	} else if opts.PersonalToken != nil {
		src = oauth2.StaticTokenSource(&oauth2.Token{AccessToken: *opts.PersonalToken})
	} else {
		return nil, fmt.Errorf("no credentials provided")
	}

	httpClient := oauth2.NewClient(ctx, src)

	if g.hostname != DefaultGithubHostname {
		client, err = NewGithubEnterpriseClient(g.hostname, httpClient)
		if err != nil {
			return nil, err
		}
	} else {
		client = github.NewClient(httpClient)
	}
	return &githubProvider{client: client}, nil
}
