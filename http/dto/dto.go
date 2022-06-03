// Package dto provides DTO models for serialization/deserialization to/from
// JSON-API
package dto

import "net/http"

// Assembler is capable of assembling itself into a JSON-API DTO object.
type Assembler interface {
	// ToJSONAPI assembles a JSON-API DTO using the current request.
	ToJSONAPI(*http.Request) any
}
