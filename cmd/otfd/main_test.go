package main

import (
	"bytes"
	"context"
	"regexp"
	"testing"

	"github.com/leg100/otf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
			err := parseFlags(ctx, tt.args, got)
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
			err := parseFlags(ctx, tt.args, got)
			require.NoError(t, err)

			assert.Regexp(t, `^otfd is the daemon component of the open terraforming framework.`, got.String())
		})
	}
}
