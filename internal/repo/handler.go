package repo

import (
	"context"
	"net/http"
	"path"

	"github.com/go-logr/logr"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal/cloud"
	"github.com/leg100/otf/internal/http/decode"
)

// handlerPrefix is the URL path prefix for the endpoint receiving vcs events
const handlerPrefix = "/webhooks/vcs"

type (
	// handler is the first point of entry for incoming VCS events, relaying them onto
	// a cloud-specific handler.
	handler struct {
		logr.Logger

		handlerBroker
		handlerDB
	}

	// handleDB is the database the handler interacts with
	handlerDB interface {
		getHookByID(context.Context, uuid.UUID) (*hook, error)
	}

	handlerBroker interface {
		publish(cloud.VCSEvent)
	}

	// cloudHandler extracts a cloud-specific event from the http request, converting it into a
	// VCS event. Returns nil if the event is to be ignored.
	cloudHandler interface {
		HandleEvent(w http.ResponseWriter, r *http.Request, secret string) *cloud.VCSEvent
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
	h.V(1).Info("received vcs event", "id", opts.ID, "repo", hook.identifier, "cloud", hook.cloud)

	if event := hook.HandleEvent(w, r, hook.secret); event != nil {
		// add non-cloud specific info to event before publishing
		event.RepoID = hook.id
		event.RepoPath = hook.identifier
		event.VCSProviderID = hook.vcsProviderID

		h.publish(*event)
	}
}
