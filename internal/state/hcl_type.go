package state

import (
	"encoding/json"
	"reflect"
)

// newHCLType returns the equivalent HCL type of a json-encoded value.
func newHCLType(v json.RawMessage) (string, error) {
	var dst any
	if err := json.Unmarshal(v, &dst); err != nil {
		return "", err
	}

	var typ string
	switch dst.(type) {
	case bool:
		typ = "bool"
	case float64:
		typ = "number"
	case string:
		typ = "string"
	case []any:
		typ = "tuple"
	case map[string]any:
		typ = "object"
	case nil:
		typ = "null"
	default:
		panic("unexpected output type: " + reflect.ValueOf(dst).String())
	}
	return typ, nil
}
