package main

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestMain checks the command tree is as it should be
func TestMain(t *testing.T) {
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
			name: "invalid",
			args: []string{"invalid", "-h"},
			err:  "unknown command \"invalid\" for \"otf\"",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Run(context.Background(), tt.args)
			if tt.err != "" {
				require.EqualError(t, err, tt.err)
			} else {
				require.NoError(t, Run(context.Background(), tt.args))
			}
		})
	}
}
