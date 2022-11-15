package main

import (
	"bytes"
	"context"
	"regexp"
	"testing"

	"github.com/leg100/otf"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
)

func TestVersion(t *testing.T) {
	ctx := context.Background()

	want := "test-version"
	otf.Version = want

	tests := []struct {
		name string
		args []string
	}{
		{
			name: "version",
			args: []string{"--version"},
		},
		{
			name: "version - shorthand",
			args: []string{"-v"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := new(bytes.Buffer)
			err := run(ctx, tt.args, got)
			require.NoError(t, err)

			regexp.MatchString(want, got.String())
		})
	}
}

func TestHelp(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name string
		args []string
	}{
		{
			name: "help",
			args: []string{"--help"},
		},
		{
			name: "help - shorthand",
			args: []string{"-h"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := new(bytes.Buffer)
			err := run(ctx, tt.args, got)
			require.NoError(t, err)

			assert.Regexp(t, `^otfd is the daemon component of the open terraforming framework.`, got.String())
		})
	}
}

func TestNewCloudConfigsFromFlags(t *testing.T) {
	flags := pflag.NewFlagSet("testing", pflag.ContinueOnError)
	cfg := newCloudConfigsFromFlags(flags)

	err := flags.Parse([]string{
		"github-client-id", "abc",
	})
	require.NoError(t, err)

	want := []*otf.CloudConfig{
		{
			Hostname: "github.com",
			Name:     otf.GithubCloudName,
			Cloud:    otf.GithubCloud{},
			Scopes:   []string{"user:email", "read:org"},
			Endpoint: oauth2.Endpoint{
				AuthURL:  "https://github.com/login/oauth/authorize",
				TokenURL: "https://github.com/login/oauth/access_token",
			},
		},
		{
			Hostname: "gitlab.com",
			Name:     otf.GitlabCloudName,
			Cloud:    otf.GitlabCloud{},
			Scopes:   []string{"read_user", "read_api"},
			Endpoint: oauth2.Endpoint{
				AuthURL:  "https://gitlab.com/oauth/authorize",
				TokenURL: "https://gitlab.com/oauth/token",
			},
		},
	}

	assert.Equal(t, want, cfg)
}
