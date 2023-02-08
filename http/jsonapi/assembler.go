package jsonapi

// Assembler is capable of assembling itself into a JSON-API DTO object.
type Assembler interface {
	// ToJSONAPI assembles a JSON-API DTO using the current request.
	ToJSONAPI() any
}
