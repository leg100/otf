package state

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/leg100/otf/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFactory(t *testing.T) {
	ctx := context.Background()
	state := testutils.ReadFile(t, "testdata/terraform.tfstate")

	t.Run("first state version with state", func(t *testing.T) {
		f := factory{&fakeDB{}}

		got, err := f.new(ctx, CreateStateVersionOptions{
			Serial:      new(int64(1)),
			State:       state,
			WorkspaceID: testutils.ParseID(t, "ws-123"),
		})
		require.NoError(t, err)

		assert.Equal(t, int64(1), got.Serial)
		assert.Equal(t, state, got.State)
		assert.Equal(t, testutils.ParseID(t, "ws-123"), got.WorkspaceID)
		// status should be finalized because state has been uploaded
		assert.Equal(t, Finalized, got.Status)

		t.Run("created outputs", func(t *testing.T) {
			assert.Equal(t, 3, len(got.Outputs))

			assert.Equal(t, "foo", got.Outputs["foo"].Name)
			assert.Equal(t, "string", got.Outputs["foo"].Type)
			assert.True(t, got.Outputs["foo"].Sensitive)
			// check value is both correct type and value
			var foo string
			err := json.Unmarshal(got.Outputs["foo"].Value, &foo)
			require.NoError(t, err)
			assert.Equal(t, "stringy", foo)

			assert.Equal(t, "bar", got.Outputs["bar"].Name)
			assert.Equal(t, "tuple", got.Outputs["bar"].Type)
			assert.JSONEq(t, `["item1","item2"]`, string(got.Outputs["bar"].Value))
			assert.False(t, got.Outputs["bar"].Sensitive)
			// check value is both correct type and value
			var bar []string
			err = json.Unmarshal(got.Outputs["bar"].Value, &bar)
			require.NoError(t, err)
			assert.Equal(t, []string{"item1", "item2"}, bar)

			assert.Equal(t, "baz", got.Outputs["baz"].Name)
			assert.Equal(t, "object", got.Outputs["baz"].Type)
			assert.False(t, got.Outputs["baz"].Sensitive)
			// check value is both correct type and value
			var baz map[string]string
			err = json.Unmarshal(got.Outputs["baz"].Value, &baz)
			require.NoError(t, err)
			assert.Equal(t, map[string]string{"key1": "value1", "key2": "value2"}, baz)
		})
	})

	t.Run("first state version without state", func(t *testing.T) {
		f := factory{&fakeDB{}}

		got, err := f.new(ctx, CreateStateVersionOptions{
			Serial:      new(int64(1)),
			WorkspaceID: testutils.ParseID(t, "ws-123"),
		})
		require.NoError(t, err)

		assert.Equal(t, int64(1), got.Serial)
		assert.Equal(t, testutils.ParseID(t, "ws-123"), got.WorkspaceID)
		assert.Empty(t, got.Outputs)
		// status should be pending because state is yet to be uploaded
		assert.Equal(t, Pending, got.Status)
	})

	t.Run("second state version with state", func(t *testing.T) {
		// seed db with first state version with serial 0
		f := factory{&fakeDB{current: &Version{Serial: 0}}}

		got, err := f.new(ctx, CreateStateVersionOptions{
			Serial:      new(int64(1)),
			State:       state,
			WorkspaceID: testutils.ParseID(t, "ws-123"),
		})
		require.NoError(t, err)

		assert.Equal(t, int64(1), got.Serial)
		assert.Equal(t, state, got.State)
		assert.Equal(t, testutils.ParseID(t, "ws-123"), got.WorkspaceID)

		// status should be finalized because state has been uploaded
		assert.Equal(t, Finalized, got.Status)
	})

	t.Run("allow creating another state version with same serial as long as state is identical", func(t *testing.T) {
		f := factory{&fakeDB{current: &Version{Serial: 1, State: state}}}

		_, err := f.new(ctx, CreateStateVersionOptions{
			Serial:      new(int64(1)),
			State:       state,
			WorkspaceID: testutils.ParseID(t, "ws-123"),
		})
		require.NoError(t, err)
	})

	t.Run("disallow creating another state version with same serial but different state", func(t *testing.T) {
		// create slightly different state
		var diffState File
		err := json.Unmarshal(state, &diffState)
		require.NoError(t, err)
		diffState.Version = 99
		state2, err := json.Marshal(diffState)
		require.NoError(t, err)

		// seed db with first state version with serial 1
		f := factory{&fakeDB{current: &Version{Serial: 1, State: state}}}

		// try to create another state version, same serial but different state
		_, err = f.new(ctx, CreateStateVersionOptions{
			Serial:      new(int64(1)),
			State:       state2,
			WorkspaceID: testutils.ParseID(t, "ws-123"),
		})
		require.Equal(t, ErrSerialMD5Mismatch, err)
	})

	t.Run("disallow creating state version with serial lower than the current state version", func(t *testing.T) {
		f := factory{&fakeDB{current: &Version{Serial: 99}}}

		_, err := f.new(ctx, CreateStateVersionOptions{
			Serial:      new(int64(1)),
			State:       state,
			WorkspaceID: testutils.ParseID(t, "ws-123"),
		})
		require.Equal(t, ErrSerialNotGreaterThanCurrent, err)
	})

	t.Run("rollback state", func(t *testing.T) {
		// seed db with a state version - it should be this version that we'll
		// rollback to.
		f := factory{&fakeDB{version: &Version{
			ID:          testutils.ParseID(t, "sv-123"),
			Serial:      1,
			State:       state,
			WorkspaceID: testutils.ParseID(t, "ws-123"),
		}}}

		got, err := f.rollback(ctx, testutils.ParseID(t, "sv-123"))
		require.NoError(t, err)

		// should create an identical state version to the one used to seed the
		// db above, albeit with a new ID.
		assert.Regexp(t, "sv-.+", got.ID)
		assert.NotEqual(t, "sv-123", got.ID)

		assert.Equal(t, testutils.ParseID(t, "ws-123"), got.WorkspaceID) // same workspace ID
		assert.Equal(t, int64(1), got.Serial)                            // same serial
		assert.Equal(t, state, got.State)                                // same state
	})
}
