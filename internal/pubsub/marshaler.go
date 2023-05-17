package pubsub

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"sync"

	"github.com/leg100/otf/internal"
)

// marshaler marshals events into postgres notifications and vice-versa.
type marshaler struct {
	tables  map[string]reflect.Type    // maps table name to type
	types   map[reflect.Type]string    // maps type to table name
	getters map[string]internal.Getter // maps table name to getter
	mu      sync.Mutex                 // sync access to maps
}

func newMarshaler() *marshaler {
	return &marshaler{
		tables:  make(map[string]reflect.Type),
		types:   make(map[reflect.Type]string),
		getters: make(map[string]internal.Getter),
	}
}

// Register a type for marshaling.
func (r *marshaler) Register(t reflect.Type, table string, getter internal.Getter) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.tables[table] = t
	r.types[t] = table
	r.getters[table] = getter
}

// marshal an event into a postgres notification. If the notification size is
// bigger than postgres permits then the notification only includes the ID of
// the event payload and the unmarshaler is expected to use the ID to fetch the
// payload instead.
func (r *marshaler) marshal(event internal.Event) ([]byte, error) {
	payload, err := json.Marshal(event.Payload)
	if err != nil {
		return nil, err
	}
	table, err := r.lookupTable(reflect.TypeOf(event.Payload))
	if err != nil {
		return nil, err
	}
	notification, err := json.Marshal(&pgevent{
		Table:   table,
		Event:   event.Type,
		Payload: payload,
	})
	if err != nil {
		return nil, err
	}
	if len(notification) > notificationMaxSize {
		// Payload is expected to be a struct with an .ID field.
		id, hasID := internal.GetID(event.Payload)
		if !hasID {
			return nil, fmt.Errorf("event payload is missing an ID field")
		}
		notification, err = json.Marshal(&pgevent{
			Table: table,
			Event: event.Type,
			ID:    &id,
		})
		if err != nil {
			return nil, err
		}
	}
	return notification, nil
}

// unmarshal postgres notification into an event.
func (r *marshaler) unmarshal(notification string) (internal.Event, error) {
	var event pgevent
	if err := json.Unmarshal([]byte(notification), &event); err != nil {
		return internal.Event{}, err
	}

	var payload any
	if event.ID != nil {
		// only ID is provided, so use that and the table name to retrieve payload.
		getter, err := r.lookupGetter(event.Table)
		if err != nil {
			return internal.Event{}, err
		}
		payload, err = getter.GetByID(context.Background(), *event.ID)
		if err != nil {
			return internal.Event{}, err
		}
	} else {
		// payload is embedded in event; lookup its type and determine if it is
		// a pointer before unmarshaling it
		typ, err := r.lookupType(event.Table)
		if err != nil {
			return internal.Event{}, err
		}
		if typ.Kind() == reflect.Pointer {
			payload = reflect.New(typ.Elem()).Interface()
		} else {
			payload = reflect.New(typ).Interface()
		}
		if err := json.Unmarshal(event.Payload, payload); err != nil {
			return internal.Event{}, err
		}
		if typ.Kind() != reflect.Pointer {
			payload = reflect.ValueOf(payload).Elem().Interface()
		}
	}
	return internal.Event{
		Type:    event.Event,
		Payload: payload,
	}, nil
}

func (r *marshaler) lookupTable(typ reflect.Type) (string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	table, ok := r.types[typ]
	if !ok {
		return "", fmt.Errorf("unregistered type: %s", typ)
	}
	return table, nil
}

func (r *marshaler) lookupType(table string) (reflect.Type, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	typ, ok := r.tables[table]
	if !ok {
		return nil, fmt.Errorf("unregistered type for table: %s", table)
	}
	return typ, nil
}

func (r *marshaler) lookupGetter(table string) (internal.Getter, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	getter, ok := r.getters[table]
	if !ok {
		return nil, fmt.Errorf("unregistered getter for table: %s", table)
	}
	return getter, nil
}
