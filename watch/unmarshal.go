package watch

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/leg100/otf"
	"github.com/leg100/otf/run"
	"github.com/r3labs/sse/v2"
)

// registry of json:api type to unmarshaler
var registry = map[string]func(b []byte) (any, error){
	"runs": func(b []byte) (any, error) {
		return run.UnmarshalJSONAPI(b)
	},
}

// unmarshal parses an SSE event and returns the equivalent OTF event
func unmarshal(event *sse.Event) (otf.Event, error) {
	if !strings.HasPrefix("run_", string(event.Event)) {
		return otf.Event{}, fmt.Errorf("no unmarshaler available for event %s", string(event.Event))
	}

	var run otf.Run
	if err := json.Unmarshal(event.Data, &run); err != nil {
		return otf.Event{}, nil
	}

	return otf.Event{Type: otf.EventType(event.Event), Payload: run}, nil
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
