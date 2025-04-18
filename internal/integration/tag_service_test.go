package integration

import (
	"errors"
	"testing"

	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/workspace"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIntegration_TagService(t *testing.T) {
	integrationTest(t)

	t.Run("add tags to workspace", func(t *testing.T) {
		svc, org, ctx := setup(t)
		ws := svc.createWorkspace(t, ctx, org)
		err := svc.Workspaces.AddTags(ctx, ws.ID, []workspace.TagSpec{
			{Name: "foo"},
			{Name: "bar"},
			{Name: "baz"},
		})
		require.NoError(t, err)

		// should have 3 tags across org
		got, err := svc.Workspaces.ListTags(ctx, org.Name, workspace.ListTagsOptions{})
		require.NoError(t, err)
		assert.Equal(t, 3, len(got.Items))

		// should have 3 tags on ws
		got, err = svc.Workspaces.ListWorkspaceTags(ctx, ws.ID, workspace.ListWorkspaceTagsOptions{})
		require.NoError(t, err)
		assert.Equal(t, 3, len(got.Items))

		t.Run("add same tags to another workspace", func(t *testing.T) {
			ws := svc.createWorkspace(t, ctx, org)
			err := svc.Workspaces.AddTags(ctx, ws.ID, []workspace.TagSpec{
				{Name: "foo"},
				{Name: "bar"},
				{Name: "baz"},
			})
			require.NoError(t, err)

			// should still have 3 tags across org
			got, err := svc.Workspaces.ListTags(ctx, org.Name, workspace.ListTagsOptions{})
			require.NoError(t, err)
			assert.Equal(t, 3, len(got.Items))

			// should have 3 tags on ws
			got, err = svc.Workspaces.ListWorkspaceTags(ctx, ws.ID, workspace.ListWorkspaceTagsOptions{})
			require.NoError(t, err)
			assert.Equal(t, 3, len(got.Items))
		})

		t.Run("invalid tag spec", func(t *testing.T) {
			err = svc.Workspaces.AddTags(ctx, ws.ID, []workspace.TagSpec{{}})
			assert.True(t, errors.Is(err, workspace.ErrInvalidTagSpec))
		})
	})

	t.Run("remove tags from workspace", func(t *testing.T) {
		svc, _, ctx := setup(t)
		ws := svc.createWorkspace(t, ctx, nil)
		err := svc.Workspaces.AddTags(ctx, ws.ID, []workspace.TagSpec{
			{Name: "foo"},
			{Name: "bar"},
			{Name: "baz"},
		})
		require.NoError(t, err)

		got, err := svc.Workspaces.ListTags(ctx, ws.Organization, workspace.ListTagsOptions{})
		require.NoError(t, err)
		assert.Equal(t, 3, len(got.Items))

		err = svc.Workspaces.RemoveTags(ctx, ws.ID, []workspace.TagSpec{
			{Name: "foo"},
			{Name: "doesnotexist"},
			{Name: "bar"},
			{Name: "baz"},
			{Name: "doesnotexist"},
		})
		require.NoError(t, err)

		got, err = svc.Workspaces.ListTags(ctx, ws.Organization, workspace.ListTagsOptions{})
		require.NoError(t, err)
		assert.Empty(t, got.Items)
	})

	t.Run("tag workspaces", func(t *testing.T) {
		svc, org, ctx := setup(t)
		ws1 := svc.createWorkspace(t, ctx, org)
		ws2 := svc.createWorkspace(t, ctx, org)
		ws3 := svc.createWorkspace(t, ctx, org)

		// create tag first by adding tag to ws1
		err := svc.Workspaces.AddTags(ctx, ws1.ID, []workspace.TagSpec{{Name: "foo"}})
		require.NoError(t, err)

		// retrieve created tag
		list, err := svc.Workspaces.ListTags(ctx, ws1.Organization, workspace.ListTagsOptions{})
		require.NoError(t, err)
		require.Equal(t, 1, len(list.Items))
		tag := list.Items[0]

		// add tag to ws2 and ws3
		err = svc.Workspaces.TagWorkspaces(ctx, tag.ID, []resource.TfeID{ws2.ID, ws3.ID})
		require.NoError(t, err)

		// check ws2 is tagged
		got, err := svc.Workspaces.ListWorkspaceTags(ctx, ws2.ID, workspace.ListWorkspaceTagsOptions{})
		require.NoError(t, err)
		if assert.Equal(t, 1, len(got.Items)) {
			assert.Equal(t, got.Items[0].Organization, ws2.Organization)
		}

		// check ws3 is tagged
		got, err = svc.Workspaces.ListWorkspaceTags(ctx, ws3.ID, workspace.ListWorkspaceTagsOptions{})
		require.NoError(t, err)
		if assert.Equal(t, 1, len(got.Items)) {
			assert.Equal(t, got.Items[0].Organization, ws3.Organization)
		}
	})

	t.Run("delete tags from organization", func(t *testing.T) {
		svc, _, ctx := setup(t)
		ws := svc.createWorkspace(t, ctx, nil)
		err := svc.Workspaces.AddTags(ctx, ws.ID, []workspace.TagSpec{
			{Name: "foo"},
			{Name: "bar"},
			{Name: "baz"},
		})
		require.NoError(t, err)

		list, err := svc.Workspaces.ListTags(ctx, ws.Organization, workspace.ListTagsOptions{})
		require.NoError(t, err)
		require.Equal(t, 3, len(list.Items))

		err = svc.Workspaces.DeleteTags(ctx, ws.Organization, []resource.TfeID{
			list.Items[0].ID,
			list.Items[1].ID,
			list.Items[2].ID,
		})
		require.NoError(t, err)

		got, err := svc.Workspaces.ListTags(ctx, ws.Organization, workspace.ListTagsOptions{})
		require.NoError(t, err)
		assert.Empty(t, got.Items)
	})
}
