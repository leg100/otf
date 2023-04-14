package state

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/leg100/otf"
	"github.com/leg100/otf/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFactory(t *testing.T) {
	ctx := context.Background()
	state := testutils.ReadFile(t, "testdata/terraform.tfstate")

	t.Run("first state version", func(t *testing.T) {
		f := factory{&fakeDB{}}

		got, err := f.create(ctx, CreateStateVersionOptions{
			Serial:      otf.Int64(0),
			State:       state,
			WorkspaceID: otf.String("ws-123"),
		})
		require.NoError(t, err)

		assert.Equal(t, int64(0), got.Serial)
		assert.Equal(t, state, got.State)
		assert.Equal(t, "ws-123", got.WorkspaceID)

		t.Run("created outputs", func(t *testing.T) {
			assert.Equal(t, 3, len(got.Outputs))

			assert.Equal(t, "foo", got.Outputs["foo"].Name)
			assert.Equal(t, "string", got.Outputs["foo"].Type)
			assert.Equal(t, `"stringy"`, got.Outputs["foo"].Value)
			assert.True(t, got.Outputs["foo"].Sensitive)

			assert.Equal(t, "bar", got.Outputs["bar"].Name)
			assert.Equal(t, "tuple", got.Outputs["bar"].Type)
			assert.Equal(t, `["item1","item2"]`, testutils.CompactJSON(t, got.Outputs["bar"].Value))
			assert.False(t, got.Outputs["bar"].Sensitive)

			assert.Equal(t, "baz", got.Outputs["baz"].Name)
			assert.Equal(t, "object", got.Outputs["baz"].Type)
			assert.Equal(t, `{"key1":"value1","key2":"value2"}`, testutils.CompactJSON(t, got.Outputs["baz"].Value))
			assert.False(t, got.Outputs["baz"].Sensitive)
		})
	})

	t.Run("second state version", func(t *testing.T) {
		f := factory{&fakeDB{current: &Version{Serial: 0}}}

		got, err := f.create(ctx, CreateStateVersionOptions{
			Serial:      otf.Int64(1),
			State:       state,
			WorkspaceID: otf.String("ws-123"),
		})
		require.NoError(t, err)

		assert.Equal(t, int64(1), got.Serial)
		assert.Equal(t, state, got.State)
		assert.Equal(t, "ws-123", got.WorkspaceID)
	})

	t.Run("same serial, matching state", func(t *testing.T) {
		f := factory{&fakeDB{current: &Version{Serial: 42, State: state}}}

		_, err := f.create(ctx, CreateStateVersionOptions{
			Serial:      otf.Int64(42),
			State:       state,
			WorkspaceID: otf.String("ws-123"),
		})
		require.NoError(t, err)
	})

	t.Run("same serial, different state", func(t *testing.T) {
		// create slightly different state
		var diffState File
		err := json.Unmarshal(state, &diffState)
		require.NoError(t, err)
		diffState.Version = 99
		state2, err := json.Marshal(diffState)
		require.NoError(t, err)

		f := factory{&fakeDB{current: &Version{Serial: 42, State: state}}}

		_, err = f.create(ctx, CreateStateVersionOptions{
			Serial:      otf.Int64(42),
			State:       state2,
			WorkspaceID: otf.String("ws-123"),
		})
		require.Equal(t, ErrSerialMD5Mismatch, err)
	})

	t.Run("serial less than current", func(t *testing.T) {
		f := factory{&fakeDB{current: &Version{Serial: 99}}}

		_, err := f.create(ctx, CreateStateVersionOptions{
			Serial:      otf.Int64(1),
			State:       state,
			WorkspaceID: otf.String("ws-123"),
		})
		require.Equal(t, ErrSerialLessThanCurrent, err)
	})

	t.Run("rollback", func(t *testing.T) {
		f := factory{&fakeDB{version: &Version{
			Serial:      4,
			State:       state,
			WorkspaceID: "ws-123",
		}}}

		got, err := f.rollback(ctx, "sv-123")
		require.NoError(t, err)
		//
		// should generate new ID
		assert.Regexp(t, "sv-.+", got.ID)
		assert.NotEqual(t, "sv-123", got.ID)

		assert.Equal(t, "ws-123", got.WorkspaceID) // same workspace ID
		assert.Equal(t, int64(4), got.Serial)      // same serial
		assert.Equal(t, state, got.State)          // same state
	})
}
