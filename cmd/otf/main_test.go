package main

import (
	"context"
	"io"
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
	}{
		{
			name: "nothing",
			args: []string{""},
		},
		{
			name: "help",
			args: []string{"-h"},
		},
		{
			name: "address",
			args: []string{"--address", "test.abc:1234"},
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
			err := run(ctx, tt.args, io.Discard)
			if tt.err != "" {
				require.EqualError(t, err, tt.err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
