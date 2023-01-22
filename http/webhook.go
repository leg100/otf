package http

import (
	"net/http"

	"github.com/go-logr/logr"
	"github.com/google/uuid"
	"github.com/leg100/otf"
	"github.com/leg100/otf/cloud"
	"github.com/leg100/otf/http/decode"
)

// webhookHandler is the point of entry for incoming VCS events, relaying them onto
// a cloud-specific handler.
type webhookHandler struct {
	events chan<- cloud.VCSEvent

	logr.Logger
	otf.Application
}

func (h *webhookHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	type options struct {
		ID uuid.UUID `schema:"webhook_id,required"`
	}
	var opts options
	if err := decode.All(&opts, r); err != nil {
		writeError(w, http.StatusUnprocessableEntity, err)
		return
	}

	hook, err := h.GetWebhook(r.Context(), opts.ID)
	if err != nil {
		h.Error(err, "received vcs event")
		writeError(w, http.StatusNotFound, err)
		return
	}
	h.V(1).Info("received vcs event", "id", opts.ID, "repo", hook.Identifier, "cloud", hook.Cloud())

	if event := hook.HandleEvent(w, r); event != nil {
		h.events <- event
	}
}
