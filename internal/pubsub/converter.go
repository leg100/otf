package pubsub

import (
	"context"
	"fmt"
	"sync"
)

type (
	// converter converts database events into OTF events.
	converter struct {
		getters map[string]Getter // maps table name to getter
		mu      sync.Mutex        // sync access to map
	}

	// Getter retrieves an event payload using its ID.
	Getter interface {
		GetByID(context.Context, string, DBAction) (any, error)
	}

	// GetterFunc is a function wrapper for Getter.
	GetterFunc func(context.Context, string, DBAction) (any, error)
)

func (f GetterFunc) GetByID(ctx context.Context, id string, action DBAction) (any, error) {
	return f(ctx, id, action)
}

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

// Register a table and getter function with the pubsub broker, to enable the broker to
// convert a database event into an OTF event.
func (r *converter) RegisterFunc(table string, getter GetterFunc) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.getters[table] = getter
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

	getter, ok := r.getters[string(event.Table)]
	if !ok {
		return Event{}, fmt.Errorf("unregistered getter for table: %s", event.Table)
	}
	payload, err := getter.GetByID(ctx, event.ID, event.Action)
	if err != nil {
		return Event{}, err
	}
	return Event{Type: eventType, Payload: payload}, nil
}
