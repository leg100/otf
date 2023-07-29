package repo

import (
	"context"
	"net/http"
	"path"

	"github.com/go-logr/logr"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal/http/decode"
)

// handlerPrefix is the URL path prefix for the endpoint receiving vcs events
const handlerPrefix = "/webhooks/vcs"

type (
	// handler is the first point of entry for incoming VCS events, relaying them onto
	// a cloud-specific handler.
	handler struct {
		logr.Logger

		*broker
		handlerDB
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

	hook, err := h.getHookByID(r.Context(), opts.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	h.V(2).Info("received vcs event", "id", opts.ID, "repo", hook.identifier, "cloud", hook.cloud)

	event := hook.HandleEvent(w, r, hook.secret)
	if event != nil {
		// add non-cloud specific info to event before publishing
		event.RepoID = hook.id
		event.VCSProviderID = hook.vcsProviderID
		event.RepoPath = hook.identifier

		h.publish(*event)
	}
}
