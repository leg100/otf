// Package github provides github related code
package github

import (
	"github.com/leg100/otf/cloud"
	"golang.org/x/oauth2"
	oauth2github "golang.org/x/oauth2/github"
)

const (
	DefaultGithubHostname string = "github.com"
)

func Defaults() cloud.Config {
	return cloud.Config{
		Name:     "github",
		Hostname: DefaultGithubHostname,
		Cloud:    &Cloud{},
	}
}

func OAuthDefaults() *oauth2.Config {
	return &oauth2.Config{
		Endpoint: oauth2github.Endpoint,
		Scopes:   []string{"user:email", "read:org"},
	}
}
