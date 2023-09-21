// Package github provides github related code
package github

import (
	oauth2github "golang.org/x/oauth2/github"
)

const (
	DefaultHostname = "github.com"
)

var (
	OAuthEndpoint = oauth2github.Endpoint

	// TODO: don't think read:org scope is necessary any more...not since OTF
	// stopped sync'ing org and team memberships from github.
	OAuthScopes = []string{"user:email", "read:org"}
)
