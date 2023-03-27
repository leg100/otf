// Package json provides helpers for the JSON encoding.
package json

import "encoding/json"

// MustMarshal marshals a value into json and panics upon error
func MustMarshal(v any) []byte {
	encoded, err := json.Marshal(v)
	if err != nil {
		panic(err.Error())
	}
	return encoded
}
