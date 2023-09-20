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
	OAuthScopes   = []string{"user:email", "read:org"}
)
