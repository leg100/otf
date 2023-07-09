package pubsub

import (
	"context"
	"fmt"
	"sync"
)

type (
	// converter converts database events into OTF events.
	converter struct {
		getters      map[string]Getter      // maps table name to getter
		unmarshalers map[string]unmarshaler // maps table name to unmarshaler
		mu           sync.Mutex             // sync access to map
	}

	// Getter retrieves an event payload using its ID.
	Getter interface {
		GetByID(context.Context, string, DBAction) (any, error)
	}

	// unmarshaler unmarshals an event payload
	unmarshaler interface {
		UnmarshalRow(data []byte) (any, error)
	}
)

func newConverter() *converter {
	return &converter{
		getters: make(map[string]Getter),
	}
}

// Register a table and getter with the pubsub broker, to enable the broker to
// convert a database event into an OTF event.
func (r *converter) Register(table string, getter Getter) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.getters[table] = getter
}

// Register a table and unmarshaler with the pubsub broker, to enable the broker to
// convert a database event into an OTF event.
func (r *converter) RegisterUnmarshaler(table string, getter unmarshaler) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.unmarshalers[table] = getter
}

// convert a database event into an OTF event
func (r *converter) convert(ctx context.Context, event pgevent) (Event, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	var eventType EventType
	switch event.Action {
	case InsertDBAction:
		eventType = CreatedEvent
	case UpdateDBAction:
		eventType = UpdatedEvent
	case DeleteDBAction:
		eventType = DeletedEvent
	default:
		return Event{}, fmt.Errorf("unknown database action: %s", event.Action)
	}

	// if payload is non-empty then expect an unmarshaler to have been
	// registered, with which to unmarshal the payload.
	if event.Payload != nil {
		u, ok := r.unmarshalers[string(event.Table)]
		if !ok {
			return Event{}, fmt.Errorf("unregistered unmarshaler for table: %s", event.Table)
		}
		payload, err := u.UnmarshalRow(event.Payload)
		if err != nil {
			return Event{}, nil
		}
		return Event{Type: eventType, Payload: payload}, nil
	}

	// otherwise expect a getter to have been registered, with which to retrieve
	// the payload
	getter, ok := r.getters[string(event.Table)]
	if !ok {
		return Event{}, fmt.Errorf("unregistered getter for table: %s", event.Table)
	}
	payload, err := getter.GetByID(ctx, event.ID, event.Action)
	if err != nil {
		return Event{}, nil
	}
	return Event{Type: eventType, Payload: payload}, nil
}
