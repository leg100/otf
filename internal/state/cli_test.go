package state

import (
	"bytes"
	"context"
	"testing"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/resource"
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
				newFakeCLI(
					&workspace.Workspace{ID: testutils.ParseID(t, "ws-123")},
					withStateVersion(&Version{ID: testutils.ParseID(t, "sv-3"), WorkspaceID: testutils.ParseID(t, "ws-123")}),
					withStateVersionList(resource.NewPage(
						[]*Version{
							{ID: testutils.ParseID(t, "sv-3")},
							{ID: testutils.ParseID(t, "sv-2")},
							{ID: testutils.ParseID(t, "sv-1")},
						},
						resource.PageOptions{},
						nil,
					)),
				),
				"sv-3 (current)\nsv-2\nsv-1\n",
			},
			{
				"zero state versions",
				newFakeCLI(&workspace.Workspace{ID: testutils.ParseID(t, "ws-123")}),
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
		cmd := newFakeCLI(nil).stateDeleteCommand()

		cmd.SetArgs([]string{"sv-123"})
		got := bytes.Buffer{}
		cmd.SetOut(&got)
		require.NoError(t, cmd.Execute())

		want := "Deleted state version: sv-123\n"
		assert.Equal(t, want, got.String())
	})

	t.Run("download", func(t *testing.T) {
		want := testutils.ReadFile(t, "./testdata/terraform.tfstate")
		cmd := newFakeCLI(nil, withState(want)).stateDownloadCommand()

		cmd.SetArgs([]string{"sv-123"})
		got := bytes.Buffer{}
		cmd.SetOut(&got)
		require.NoError(t, cmd.Execute())

		assert.JSONEq(t, string(want), got.String())
	})

	t.Run("rollback", func(t *testing.T) {
		sv := &Version{ID: testutils.ParseID(t, "sv-456")}
		cmd := newFakeCLI(nil, withStateVersion(sv)).stateRollbackCommand()

		cmd.SetArgs([]string{"sv-123"})
		got := bytes.Buffer{}
		cmd.SetOut(&got)
		require.NoError(t, cmd.Execute())

		assert.Equal(t, "Successfully rolled back state\n", got.String())
	})
}

type (
	fakeCLIService struct {
		stateVersion     *Version
		stateVersionList *resource.Page[*Version]
		state            []byte
		workspace        *workspace.Workspace
	}

	fakeCLIOption func(*fakeCLIService)
)

func newFakeCLI(ws *workspace.Workspace, opts ...fakeCLIOption) *CLI {
	svc := fakeCLIService{workspace: ws}
	for _, fn := range opts {
		fn(&svc)
	}
	return &CLI{state: &svc, workspaces: &svc}
}

func withStateVersion(sv *Version) fakeCLIOption {
	return func(c *fakeCLIService) {
		c.stateVersion = sv
	}
}

func withStateVersionList(svl *resource.Page[*Version]) fakeCLIOption {
	return func(c *fakeCLIService) {
		c.stateVersionList = svl
	}
}

func withState(state []byte) fakeCLIOption {
	return func(c *fakeCLIService) {
		c.state = state
	}
}

func (f *fakeCLIService) List(context.Context, resource.ID, resource.PageOptions) (*resource.Page[*Version], error) {
	return f.stateVersionList, nil
}

func (f *fakeCLIService) GetCurrent(ctx context.Context, workspaceID resource.ID) (*Version, error) {
	if f.stateVersion == nil {
		return nil, internal.ErrResourceNotFound
	}
	return f.stateVersion, nil
}

func (f *fakeCLIService) Delete(ctx context.Context, svID resource.ID) error {
	return nil
}

func (f *fakeCLIService) Rollback(ctx context.Context, svID resource.ID) (*Version, error) {
	return f.stateVersion, nil
}

func (f *fakeCLIService) Download(ctx context.Context, svID resource.ID) ([]byte, error) {
	return f.state, nil
}

func (f *fakeCLIService) GetByName(context.Context, string, string) (*workspace.Workspace, error) {
	return f.workspace, nil
}
