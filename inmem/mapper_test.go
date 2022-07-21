package inmem

import (
	"context"
	"testing"

	"github.com/leg100/otf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMapper(t *testing.T) {
	org, err := otf.NewOrganization(otf.OrganizationCreateOptions{Name: otf.String("test-org")})
	require.NoError(t, err)
	ws1, err := otf.NewWorkspace(org, otf.WorkspaceCreateOptions{Name: "test-ws"})
	require.NoError(t, err)
	cv1, err := otf.NewConfigurationVersion(ws1.ID(), otf.ConfigurationVersionCreateOptions{})
	require.NoError(t, err)
	run1 := otf.NewRun(cv1, ws1, otf.RunCreateOptions{})

	m := NewMapper()
	err = m.Populate(&fakeWorkspaceService{
		workspaces: []*otf.Workspace{ws1},
	}, &fakeRunService{
		runs: []*otf.Run{run1},
	})
	require.NoError(t, err)

	t.Run("lookup workspace ID", func(t *testing.T) {
		assert.Equal(t, ws1.ID(), m.LookupWorkspaceID(ws1.SpecName()))
	})

	t.Run("authorized user", func(t *testing.T) {
		ctx := otf.AddSubjectToContext(context.Background(), &fakeSubject{"test-org"})
		assert.True(t, m.CanAccessRun(ctx, run1.ID()))
	})

	t.Run("unauthorized user", func(t *testing.T) {
		ctx := otf.AddSubjectToContext(context.Background(), &fakeSubject{"another-org"})
		assert.False(t, m.CanAccessRun(ctx, run1.ID()))
	})
}
