package http

import (
	"net/http"

	"github.com/go-logr/logr"
	"github.com/google/uuid"
	"github.com/leg100/otf"
	"github.com/leg100/otf/http/decode"
)

type webhookHandler struct {
	events chan<- otf.VCSEvent

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

	hook, err := h.DB().GetWebhook(r.Context(), opts.ID)
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}

	if event := hook.HandleEvent(w, r, hook); event != nil {
		h.events <- *event
	}
}
