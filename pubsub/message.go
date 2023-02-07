package pubsub

// message is the schema of the payload for use in the postgres notification channel.
type message struct {
	// Table is the postgres table on which the event occured
	Table string `json:"relation"`
	// Action is the type of change made to the relation
	Action string `json:"action"`
	// ID is the primary key of the changed row
	ID string `json:"id"`
	// PID is the process id that sent this event
	PID string `json:"pid"`
}
