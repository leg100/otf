// Package gitlab provides gitlab related code
package gitlab

import (
	"github.com/leg100/otf/cloud"
	"golang.org/x/oauth2"
	oauth2gitlab "golang.org/x/oauth2/gitlab"
)

func Defaults() cloud.Config {
	return cloud.Config{
		Name:     "gitlab",
		Hostname: "gitlab.com",
		Cloud:    &Cloud{},
	}
}

func OAuthDefaults() *oauth2.Config {
	return &oauth2.Config{
		Endpoint: oauth2gitlab.Endpoint,
		Scopes:   []string{"read_user", "read_api"},
	}
}
