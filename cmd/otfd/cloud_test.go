package main

import (
	"testing"

	"github.com/leg100/otf"
	"github.com/leg100/otf/github"
	"github.com/leg100/otf/gitlab"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
)

func TestNewCloudConfigsFromFlags(t *testing.T) {
	flags := pflag.NewFlagSet("testing", pflag.ContinueOnError)
	got := cloudFlags(flags)

	err := flags.Parse([]string{
		"--github-client-id", "my-github-client-id",
		"--github-client-secret", "my-github-client-secret",
		"--github-hostname", "my-own-github.com",
		"--gitlab-client-id", "my-gitlab-client-id",
		"--gitlab-client-secret", "my-gitlab-client-secret",
		"--gitlab-hostname", "my-own-gitlab.com",
		"--gitlab-skip-tls-verification",
	})
	require.NoError(t, err)

	want := []*cloudConfig{
		{
			CloudConfig: otf.CloudConfig{
				Name:     "github",
				Hostname: "my-own-github.com",
				Cloud:    &github.Cloud{},
			},
			Config: &oauth2.Config{
				Scopes: []string{"user:email", "read:org"},
				Endpoint: oauth2.Endpoint{
					AuthURL:  "https://github.com/login/oauth/authorize",
					TokenURL: "https://github.com/login/oauth/access_token",
				},
				ClientID:     "my-github-client-id",
				ClientSecret: "my-github-client-secret",
			},
		},
		{
			CloudConfig: otf.CloudConfig{
				Name:                "gitlab",
				Hostname:            "my-own-gitlab.com",
				Cloud:               &gitlab.Cloud{},
				SkipTLSVerification: true,
			},
			Config: &oauth2.Config{
				Scopes: []string{"read_user", "read_api"},
				Endpoint: oauth2.Endpoint{
					AuthURL:  "https://gitlab.com/oauth/authorize",
					TokenURL: "https://gitlab.com/oauth/token",
				},
				ClientID:     "my-gitlab-client-id",
				ClientSecret: "my-gitlab-client-secret",
			},
		},
	}
	assert.Equal(t, want, got)
}
