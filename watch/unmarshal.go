package watch

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/leg100/otf"
	"github.com/leg100/otf/run"
	"github.com/r3labs/sse/v2"
)

// unmarshal parses an SSE event and returns the equivalent OTF event
func unmarshal(event *sse.Event) (otf.Event, error) {
	if !strings.HasPrefix("run_", string(event.Event)) {
		return otf.Event{}, fmt.Errorf("no unmarshaler available for event %s", string(event.Event))
	}

	var run run.Run
	if err := json.Unmarshal(event.Data, &run); err != nil {
		return otf.Event{}, nil
	}

	return otf.Event{Type: otf.EventType(event.Event), Payload: run}, nil
}
