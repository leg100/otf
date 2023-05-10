package main

import (
	"bytes"
	"context"
	"io"
	"regexp"
	"testing"

	"github.com/leg100/otf/internal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVersion(t *testing.T) {
	ctx := context.Background()

	want := "test-version"
	internal.Version = want

	got := new(bytes.Buffer)
	err := parseFlags(ctx, []string{"--version"}, got)
	require.NoError(t, err)

	regexp.MatchString(want, got.String())
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

func TestInvalidSecret(t *testing.T) {
	ctx := context.Background()

	err := parseFlags(ctx, []string{"--secret", "not-hex"}, io.Discard)
	assert.Error(t, err)
	want := "invalid argument \"not-hex\" for \"--secret\" flag: encoding/hex: invalid byte: U+006E 'n'"
	assert.Equal(t, want, err.Error())
}
