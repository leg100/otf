package main

import (
	"bytes"
	"context"
	"regexp"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestMain checks the command tree is as it should be
func TestMain(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name string
		args []string
		err  string
		// want is a regex of wanted output
		want string
	}{
		{
			name: "nothing",
			args: []string{""},
			want: `^Usage:\n\totf \[command\]`,
		},
		{
			name: "help",
			args: []string{"-h"},
			want: `^Usage:\n\totf \[command\]`,
		},
		{
			name: "organization new",
			args: []string{"organizations", "new", "-h"},
		},
		{
			name: "workspace lock",
			args: []string{"workspaces", "lock", "-h"},
		},
		{
			name: "workspace unlock",
			args: []string{"workspaces", "unlock", "-h"},
		},
		{
			name: "agent token create",
			args: []string{"agents", "tokens", "new", "-h"},
		},
		{
			name: "invalid",
			args: []string{"invalid", "-h"},
			err:  "unknown command \"invalid\" for \"otf\"",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := new(bytes.Buffer)
			err := run(ctx, tt.args, got)
			if tt.err != "" {
				require.EqualError(t, err, tt.err)
				return
			}
			require.NoError(t, err)

			regexp.MatchString(tt.want, got.String())
		})
	}
}
