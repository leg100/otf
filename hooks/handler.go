package hooks

import (
	"net/http"

	"github.com/go-logr/logr"
	"github.com/google/uuid"
	"github.com/leg100/otf"
	"github.com/leg100/otf/cloud"
	"github.com/leg100/otf/http/decode"
)

// handler is the first point of entry for incoming VCS events, relaying them onto
// a cloud-specific handler.
type handler struct {
	logr.Logger
	otf.PubSubService

	db
}

func NewHandler(logger logr.Logger, app otf.Application) *handler {
	return &handler{
		Logger:        logger,
		PubSubService: app,
		db:            newPGDB(app.DB(), newFactory(app, app)),
	}
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	type options struct {
		ID uuid.UUID `schema:"webhook_id,required"`
	}
	var opts options
	if err := decode.All(&opts, r); err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	hook, err := h.get(r.Context(), opts.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	h.V(1).Info("received vcs event", "id", opts.ID, "repo", hook.identifier, "cloud", hook.cloud)

	event := hook.HandleEvent(w, r, cloud.HandleEventOptions{Secret: hook.secret, WebhookID: hook.id})
	if event != nil {
		h.Publish(otf.Event{
			Type:    otf.EventVCS,
			Payload: event,
		})
	}
}
