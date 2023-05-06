package repo

import (
	"context"
	"net/http"
	"path"

	"github.com/go-logr/logr"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	internal "github.com/leg100/otf"
	"github.com/leg100/otf/cloud"
	"github.com/leg100/otf/http/decode"
)

// handlerPrefix is the URL path prefix for the endpoint receiving vcs events
const handlerPrefix = "/webhooks/vcs"

type (
	// handler is the first point of entry for incoming VCS events, relaying them onto
	// a cloud-specific handler.
	handler struct {
		logr.Logger
		internal.Publisher

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
		h.Publish(internal.Event{
			Type:    internal.EventVCS,
			Payload: event,
			// publish vcs events only to the local node; a "run spawner" and a
			// "module publisher" subscribe to vcs events on each node, so it is
			// only necessary to send event to local node; sending it to other
			// nodes would lead to duplicate runs and modules being spawned and
			// published.
			Local: true,
		})
	}
}
