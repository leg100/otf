package otf

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetID(t *testing.T) {
	t.Run("with id", func(t *testing.T) {
		s := struct {
			ID string
		}{
			ID: "foo-123",
		}
		got, ok := GetID(s)
		require.True(t, ok)
		assert.Equal(t, "foo-123", got)
	})

	t.Run("ptr with id", func(t *testing.T) {
		s := struct {
			ID string
		}{
			ID: "foo-123",
		}
		got, ok := GetID(&s)
		require.True(t, ok)
		assert.Equal(t, "foo-123", got)
	})

	t.Run("without id", func(t *testing.T) {
		s := struct {
			SomeOtherField string
		}{
			SomeOtherField: "foo-123",
		}
		_, ok := GetID(s)
		assert.False(t, ok)
	})

	t.Run("nil", func(t *testing.T) {
		_, ok := GetID(nil)
		assert.False(t, ok)
	})
}
