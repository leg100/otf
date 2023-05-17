package pubsub

import (
	"reflect"
	"testing"

	"github.com/leg100/otf/internal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMarshaler(t *testing.T) {
	t.Run("marshal", func(t *testing.T) {
		m := newMarshaler()
		m.Register(reflect.TypeOf(&fakeType{}), "fakes", nil)

		got, err := m.marshal(Event{
			Type:    CreatedEvent,
			Payload: &fakeType{ID: "fake-123", Stuff: []byte("stuff")},
		})
		require.NoError(t, err)

		want := `{"table":"fakes","event":"created","payload":{"id":"fake-123","stuff":"c3R1ZmY="}}`
		assert.Equal(t, want, string(got))
	})

	// marshal a big event that exceeds the maximum permitted by postgres - the
	// marshaler should just marshal the event payload's ID.
	t.Run("marshal big event", func(t *testing.T) {
		stuff := internal.GenerateRandomString(notificationMaxSize + 1)
		m := newMarshaler()
		m.Register(reflect.TypeOf(&fakeType{}), "fakes", nil)

		got, err := m.marshal(Event{
			Type:    CreatedEvent,
			Payload: &fakeType{ID: "fake-123", Stuff: []byte(stuff)},
		})
		require.NoError(t, err)

		want := `{"table":"fakes","event":"created","id":"fake-123"}`
		assert.Equal(t, want, string(got))
	})

	t.Run("unmarshal", func(t *testing.T) {
		m := newMarshaler()
		m.Register(reflect.TypeOf(&fakeType{}), "fakes", nil)

		notification := `{"table":"fakes","event":"created","payload":{"id":"fake-123","stuff":"c3R1ZmY="}}`
		got, err := m.unmarshal(notification)
		require.NoError(t, err)

		want := Event{
			Type:    CreatedEvent,
			Payload: &fakeType{ID: "fake-123", Stuff: []byte("stuff")},
		}
		assert.Equal(t, want, got)
	})

	t.Run("unmarshal non-pointer struct", func(t *testing.T) {
		m := newMarshaler()
		m.Register(reflect.TypeOf(fakeType{}), "fakes", nil)

		notification := `{"table":"fakes","event":"created","payload":{"id":"fake-123","stuff":"c3R1ZmY="}}`
		got, err := m.unmarshal(notification)
		require.NoError(t, err)

		want := Event{
			Type:    CreatedEvent,
			Payload: fakeType{ID: "fake-123", Stuff: []byte("stuff")},
		}
		assert.Equal(t, want, got)
	})

	t.Run("unmarshal using ID", func(t *testing.T) {
		m := newMarshaler()
		fake := &fakeType{ID: "fake-123", Stuff: []byte("stuff")}
		m.Register(reflect.TypeOf(&fakeType{}), "fakes", &fakeGetter{fake: fake})

		notification := `{"table":"fakes","event":"created","id":"fake-123"}`
		got, err := m.unmarshal(notification)
		require.NoError(t, err)

		want := Event{Type: CreatedEvent, Payload: fake}
		assert.Equal(t, want, got)
	})
}
