package integration

import (
	"testing"

	"github.com/leg100/otf"
	"github.com/leg100/otf/tags"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIntegration_TagService(t *testing.T) {
	t.Parallel()

	t.Run("add tags to workspace", func(t *testing.T) {
		svc := setup(t, nil)
		ws := svc.createWorkspace(t, ctx, nil)
		err := svc.AddTags(ctx, ws.ID, []tags.TagSpec{
			{Name: otf.String("foo")},
			{Name: otf.String("bar")},
			{Name: otf.String("baz")},
		})
		require.NoError(t, err)

		got, err := svc.ListWorkspaceTags(ctx, ws.ID, tags.ListWorkspaceTagsOptions{})
		require.NoError(t, err)
		assert.Equal(t, 3, len(got.Items))

		t.Run("invalid tag spec", func(t *testing.T) {
			err = svc.AddTags(ctx, ws.ID, []tags.TagSpec{{}})
			assert.Equal(t, tags.ErrInvalidTagSpec, err)
		})
	})

	t.Run("remove tags from workspace", func(t *testing.T) {
		svc := setup(t, nil)
		ws := svc.createWorkspace(t, ctx, nil)
		err := svc.AddTags(ctx, ws.ID, []tags.TagSpec{
			{Name: otf.String("foo")},
			{Name: otf.String("bar")},
			{Name: otf.String("baz")},
		})
		require.NoError(t, err)

		err = svc.RemoveTags(ctx, ws.ID, []tags.TagSpec{
			{Name: otf.String("foo")},
			{Name: otf.String("bar")},
			{Name: otf.String("baz")},
		})
		require.NoError(t, err)

		got, err := svc.ListTags(ctx, ws.Organization, tags.ListTagsOptions{})
		require.NoError(t, err)
		assert.Empty(t, got.Items)
	})

	t.Run("tag workspaces", func(t *testing.T) {
		svc := setup(t, nil)
		org := svc.createOrganization(t, ctx)
		ws1 := svc.createWorkspace(t, ctx, org)
		ws2 := svc.createWorkspace(t, ctx, org)
		ws3 := svc.createWorkspace(t, ctx, org)

		// create tag first by adding tag to ws1
		err := svc.AddTags(ctx, ws1.ID, []tags.TagSpec{{Name: otf.String("foo")}})
		require.NoError(t, err)

		// retrieve created tag
		list, err := svc.ListTags(ctx, ws1.Organization, tags.ListTagsOptions{})
		require.NoError(t, err)
		require.Equal(t, 1, len(list.Items))
		tag := list.Items[0]

		// add tag to ws2 and ws3
		err = svc.TagWorkspaces(ctx, tag.ID, []string{ws2.ID, ws3.ID})
		require.NoError(t, err)

		// check ws2 is tagged
		got, err := svc.ListWorkspaceTags(ctx, ws2.ID, tags.ListWorkspaceTagsOptions{})
		require.NoError(t, err)
		if assert.Equal(t, 1, len(got.Items)) {
			assert.Equal(t, got.Items[0].Organization, ws2.Organization)
		}

		// check ws3 is tagged
		got, err = svc.ListWorkspaceTags(ctx, ws3.ID, tags.ListWorkspaceTagsOptions{})
		require.NoError(t, err)
		if assert.Equal(t, 1, len(got.Items)) {
			assert.Equal(t, got.Items[0].Organization, ws3.Organization)
		}
	})
}
