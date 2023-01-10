package otf

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/leg100/otf/cloud"
)

// EventHandler handles incoming events
type EventHandler interface {
	// HandleEvent extracts a cloud-specific event from the http request, converting it into a
	// VCS event. Returns nil if the event is to be ignored.
	HandleEvent(w http.ResponseWriter, r *http.Request, opts HandleEventOptions) cloud.VCSEvent
}

type HandleEventOptions struct {
	Secret    string
	WebhookID uuid.UUID
}
