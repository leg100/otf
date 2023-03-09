package repo

import (
	"context"
	"net/http"
	"path"

	"github.com/go-logr/logr"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/leg100/otf"
	"github.com/leg100/otf/cloud"
	"github.com/leg100/otf/http/decode"
)

type (
	// handler is the first point of entry for incoming VCS events, relaying them onto
	// a cloud-specific handler.
	handler struct {
		logr.Logger
		otf.PubSubService

		db handlerDB
	}

	// handleDB is the database the handler interacts with
	handlerDB interface {
		getHookByID(context.Context, uuid.UUID) (*hook, error)
	}
)

func (h *handler) AddHandlers(r *mux.Router) {
	r.Handle(path.Join(handlerPrefix, "{webhook_id}"), h)
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	opts := struct {
		ID uuid.UUID `schema:"webhook_id,required"`
	}{}
	if err := decode.All(&opts, r); err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	hook, err := h.db.getHookByID(r.Context(), opts.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	h.V(1).Info("received vcs event", "id", opts.ID, "repo", hook.identifier, "cloud", hook.cloud)

	event := hook.HandleEvent(w, r, cloud.HandleEventOptions{Secret: hook.secret, RepoID: hook.id})
	if event != nil {
		h.Publish(otf.Event{
			Type:    otf.EventVCS,
			Payload: event,
		})
	}
}
