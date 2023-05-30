package pubsub

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConverter(t *testing.T) {
	ctx := context.Background()

	t.Run("convert INSERT", func(t *testing.T) {
		m := newConverter()
		fake := &fakeType{ID: "fake-123", Stuff: []byte("stuff")}
		m.Register("fakes", &fakeGetter{fake: fake})

		got, err := m.convert(ctx, pgevent{ID: "fake-123", Action: InsertDBAction, Table: "fakes"})
		require.NoError(t, err)

		want := Event{Type: CreatedEvent, Payload: fake}
		assert.Equal(t, want, got)
	})
}
