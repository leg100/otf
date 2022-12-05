package gitlab

import (
	"github.com/leg100/otf"
	"golang.org/x/oauth2"
	oauth2gitlab "golang.org/x/oauth2/gitlab"
)

func Defaults() otf.CloudConfig {
	return otf.CloudConfig{
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
