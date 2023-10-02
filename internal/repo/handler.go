package repo

import (
	"context"
	"net/http"
	"path"

	"github.com/go-logr/logr"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/cloud"
	"github.com/leg100/otf/internal/http/decode"
	"github.com/leg100/otf/internal/vcs"
)

// handlerPrefix is the URL path prefix for the endpoint receiving vcs events
const handlerPrefix = "/webhooks/vcs"

type (
	// handler handles repohook VCS events.
	handler struct {
		logr.Logger

		vcs.Publisher
		handlerDB
		cloudHandlers *internal.SafeMap[cloud.Kind, CloudHandler]
	}

	CloudHandler func(w http.ResponseWriter, r *http.Request, secret string) *vcs.Event

	// handleDB is the database the handler interacts with
	handlerDB interface {
		getHookByID(context.Context, uuid.UUID) (*hook, error)
	}
)

func newHandler(logger logr.Logger, publisher vcs.Publisher, db handlerDB) *handler {
	return &handler{
		Logger:        logger,
		Publisher:     publisher,
		handlerDB:     db,
		cloudHandlers: internal.NewSafeMap[cloud.Kind, CloudHandler](),
	}
}

func (h *handler) AddHandlers(r *mux.Router) {
	r.Handle(path.Join(handlerPrefix, "{webhook_id}"), h)
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var opts struct {
		ID uuid.UUID `schema:"webhook_id,required"`
	}
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

	cloudHandler, ok := h.cloudHandlers.Get(hook.cloud)
	if !ok {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	if event := cloudHandler(w, r, hook.secret); event != nil {
		// add non-cloud specific info to event before publishing
		event.RepoID = hook.id
		event.RepoPath = hook.identifier
		event.VCSProviderID = hook.vcsProviderID

		h.Publish(*event)
	}
}
