// Package dto provides DTO models for serialization/deserialization to/from
// JSON-API
package dto

// Assembler is capable of assembling itself into a JSONAPI DTO object.
type Assembler interface {
	ToJSONAPI() any
}
