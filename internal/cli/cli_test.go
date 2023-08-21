package cli

import (
	"bytes"
	"context"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetToken_Env(t *testing.T) {
	t.Setenv("OTF_TOKEN", "mytoken")
	got, err := (&CLI{}).getToken("localhost:8080")
	require.NoError(t, err)
	assert.Equal(t, "mytoken", got)
}

func TestSetToken_HostSpecificEnv(t *testing.T) {
	t.Setenv("TF_TOKEN_otf_dev", "mytoken")
	got, err := (&CLI{}).getToken("otf.dev")
	require.NoError(t, err)
	assert.Equal(t, "mytoken", got)
}

func TestSetToken_CredentialStore(t *testing.T) {
	store := CredentialsStore(filepath.Join(t.TempDir(), "creds.json"))
	require.NoError(t, store.Save("otf.dev", "mytoken"))

	got, err := (&CLI{creds: store}).getToken("otf.dev")
	require.NoError(t, err)
	assert.Equal(t, "mytoken", got)
}

// TestCommandTree checks the command tree is as it should be
func TestCommandTree(t *testing.T) {
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
			want: `^Usage:\n  otf \[command\]`,
		},
		{
			name: "help",
			args: []string{"-h"},
			want: `^Usage:\n  otf \[command\]`,
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
			err := NewCLI().Run(ctx, tt.args, got)
			if tt.err != "" {
				require.EqualError(t, err, tt.err)
				return
			}
			require.NoError(t, err)

			assert.Regexp(t, tt.want, got.String())
		})
	}
}
