package otf

import "net/http"

// EventHandler handles incoming events
type EventHandler interface {
	// HandleEvent retrieves an event from the http request corresponding to the
	// given webhook.
	HandleEvent(w http.ResponseWriter, r *http.Request, hook *Webhook) *VCSEvent
}
