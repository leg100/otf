package repohooks

import (
	"context"
	"net/http"
	"path"

	"github.com/go-logr/logr"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/http/decode"
	"github.com/leg100/otf/internal/vcs"
)

const (
	// handlerPrefix is the URL path prefix for endpoints receiving vcs events
	handlerPrefix = "/webhooks/vcs"
)

type (
	// handlers handle VCS events triggered by webhooks
	handlers struct {
		logr.Logger
		vcs.Publisher

		cloudHandlers *internal.SafeMap[vcs.Kind, EventUnmarshaler]

		handlerDB
	}

	// EventUnmarshaler does two things:
	// (a) handles incoming request (containing a VCS event) and sends appropriate response
	// (b) unmarshals event from the request; if the event is irrelevant or
	// invalid then nil is returned.
	EventUnmarshaler func(w http.ResponseWriter, r *http.Request, secret string) *vcs.EventPayload

	// handleDB is the database the handler interacts with
	handlerDB interface {
		getHookByID(context.Context, uuid.UUID) (*hook, error)
	}
)

func newHandler(logger logr.Logger, publisher vcs.Publisher, db handlerDB) *handlers {
	return &handlers{
		Logger:        logger,
		Publisher:     publisher,
		handlerDB:     db,
		cloudHandlers: internal.NewSafeMap[vcs.Kind, EventUnmarshaler](),
	}
}

func (h *handlers) AddHandlers(r *mux.Router) {
	r.HandleFunc(path.Join(handlerPrefix, "{webhook_id}"), h.repohookHandler)
}

func (h *handlers) repohookHandler(w http.ResponseWriter, r *http.Request) {
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
	h.V(2).Info("received vcs event", "repohook_id", opts.ID, "repo", hook.repoPath, "cloud", hook.cloud)

	cloudHandler, ok := h.cloudHandlers.Get(hook.cloud)
	if !ok {
		h.Error(nil, "no event unmarshaler found for event", "repohook_id", opts.ID, "repo", hook.repoPath, "cloud", hook.cloud)
		http.Error(w, "no event unmarshaler found for event", http.StatusNotFound)
		return
	}
	if payload := cloudHandler(w, r, hook.secret); payload != nil {
		h.Publish(vcs.Event{
			EventHeader:  vcs.EventHeader{VCSProviderID: hook.vcsProviderID},
			EventPayload: *payload,
		})
	} else {
		h.V(2).Info("ignoring vcs event", "repohook_id", opts.ID, "repo", hook.repoPath, "cloud", hook.cloud)
	}
}
