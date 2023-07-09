package hooks

// Event is a wrapper containing the data being dispatched through a hook
type Event[T any] struct {
	// Msg contains the data being dispatched through a hook
	Msg T

	// Hook contains the hook that dispatched this event
	Hook *Hook[T]
}

// newEvent creates a new event
func newEvent[T any](hook *Hook[T], message T) Event[T] {
	return Event[T]{
		Msg:  message,
		Hook: hook,
	}
}
