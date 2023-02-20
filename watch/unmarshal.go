package watch

import (
	"encoding/json"
	"fmt"

	"github.com/leg100/otf/run"
)

// registry of json:api type to unmarshaler
var registry = map[string]func(b []byte) (any, error){
	"runs": func(b []byte) (any, error) {
		return run.UnmarshalJSONAPI(b)
	},
}

// unmarshal parses a json:api document and returns its equivalent Go Type
// populated with the result.
func unmarshal(b []byte) (any, error) {
	// lookup json:api type
	t, err := unmarshalJSONAPIType(b)
	if err != nil {
		return nil, err
	}

	// lookup unmarshaler for type
	unmarshaler, ok := registry[t]
	if !ok {
		return nil, fmt.Errorf("no json:api unmarshaler found for type %s", t)
	}

	return unmarshaler(b)
}

type jsonapiBody struct {
	Data jsonapiData `json:"data"`
}

type jsonapiData struct {
	Typ string `json:"type"`
}

// unmarshalJSONAPIType unmarshals the json:api type field.
func unmarshalJSONAPIType(b []byte) (string, error) {
	var body jsonapiBody
	err := json.Unmarshal(b, &body)
	if err != nil {
		return "", err
	}
	return body.Data.Typ, nil
}
