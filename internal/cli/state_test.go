package cli

import (
	"bytes"
	"testing"

	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/state"
	"github.com/leg100/otf/internal/testutils"
	"github.com/leg100/otf/internal/workspace"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCLI_State(t *testing.T) {
	// list state versions: sv-1, sv-2, sv-3, where sv-3 is the current state
	// version
	t.Run("list", func(t *testing.T) {
		tests := []struct {
			name string
			app  *CLI
			want string
		}{
			{
				"three state versions",
				fakeApp(
					withWorkspaces(&workspace.Workspace{ID: "ws-123"}),
					withStateVersion(&state.Version{ID: "sv-3", WorkspaceID: "ws-123"}),
					withStateVersionList(resource.NewPage(
						[]*state.Version{
							{ID: "sv-3"},
							{ID: "sv-2"},
							{ID: "sv-1"},
						},
						resource.PageOptions{},
						nil,
					)),
				),
				"sv-3 (current)\nsv-2\nsv-1\n",
			},
			{
				"zero state versions",
				fakeApp(withWorkspaces(&workspace.Workspace{ID: "ws-123"})),
				"No state versions found\n",
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				cmd := tt.app.stateListCommand()

				cmd.SetArgs([]string{"--organization", "acme-corp", "--workspace", "dev"})
				got := bytes.Buffer{}
				cmd.SetOut(&got)
				require.NoError(t, cmd.Execute())

				assert.Equal(t, tt.want, got.String())
			})
		}
	})

	t.Run("delete", func(t *testing.T) {
		cmd := fakeApp().stateDeleteCommand()

		cmd.SetArgs([]string{"sv-123"})
		got := bytes.Buffer{}
		cmd.SetOut(&got)
		require.NoError(t, cmd.Execute())

		want := "Deleted state version: sv-123\n"
		assert.Equal(t, want, got.String())
	})

	t.Run("download", func(t *testing.T) {
		want := testutils.ReadFile(t, "./testdata/terraform.tfstate")
		cmd := fakeApp(withState(want)).stateDownloadCommand()

		cmd.SetArgs([]string{"sv-123"})
		got := bytes.Buffer{}
		cmd.SetOut(&got)
		require.NoError(t, cmd.Execute())

		assert.JSONEq(t, string(want), got.String())
	})

	t.Run("rollback", func(t *testing.T) {
		sv := &state.Version{ID: "sv-456"}
		cmd := fakeApp(withStateVersion(sv)).stateRollbackCommand()

		cmd.SetArgs([]string{"sv-123"})
		got := bytes.Buffer{}
		cmd.SetOut(&got)
		require.NoError(t, cmd.Execute())

		assert.Equal(t, "Successfully rolled back state\n", got.String())
	})
}
