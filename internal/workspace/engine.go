package workspace

import (
	"encoding/json"

	"github.com/leg100/otf/internal/engine"
)

// UpdateEngineOption wraps Engine, in order to differentiate between Engine having been
// explicitly to null (which is interpreted as update to use the default
// engine), or whether it was simply omitted (don't update).
type UpdateEngineOption struct {
	*engine.Engine

	Valid bool `json:"-"`
	Set   bool `json:"-"`
}

// UnmarshalJSON differentiates between Engine having been explicitly
// set to null by the client, or the client has left it out.
func (e *UpdateEngineOption) UnmarshalJSON(data []byte) error {
	// If this method was called, the value was set.
	e.Set = true

	if string(data) == "null" {
		// The key was set to null
		e.Valid = false
		return nil
	}

	// The key isn't set to null
	if err := json.Unmarshal(data, &e.Engine); err != nil {
		return err
	}
	e.Valid = true
	return nil
}

// UnmarshalText differentiates between Engine having been explicitly
// set to an empty string by the client, or the client has left it out.
func (e *UpdateEngineOption) UnmarshalText(text []byte) error {
	// If this method was called, the value was set.
	e.Set = true

	// set to empty string which means set to whatever the system default engine
	// is.
	if string(text) == "" {
		e.Valid = false
		return nil
	}

	// not an empty string, so set to an explicit engine
	e.Valid = true
	return e.Engine.Set(string(text))
}
