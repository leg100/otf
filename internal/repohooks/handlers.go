package repohooks

import (
	"context"
	"errors"
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

	// EventUnmarshaler validates the request using the secret and unmarshals
	// the event contained in the request body. If the request is to be ignored
	// then the unmarshaler should return vcs.ErrIgnoreEvent, explaining why the
	// event was ignored.
	EventUnmarshaler func(r *http.Request, secret string) (*vcs.EventPayload, error)

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

	// look up cloud-specific handler for event
	cloudHandler, ok := h.cloudHandlers.Get(hook.cloud)
	if !ok {
		h.Error(nil, "no event unmarshaler found for event", "repohook_id", opts.ID, "repo", hook.repoPath, "cloud", hook.cloud)
		http.Error(w, "no event unmarshaler found for event", http.StatusNotFound)
		return
	}
	// handle event
	payload, err := cloudHandler(r, hook.secret)
	// either ignore the event, return an error, or publish the event onwards
	var ignore vcs.ErrIgnoreEvent
	if errors.As(err, &ignore) {
		h.V(2).Info("ignoring event: "+err.Error(), "repohook_id", opts.ID, "repo", hook.repoPath, "cloud", hook.cloud)
		return
	} else if err != nil {
		h.Error(err, "handling vcs event", "repohook_id", opts.ID, "repo", hook.repoPath, "cloud", hook.cloud)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	h.Publish(vcs.Event{
		EventHeader:  vcs.EventHeader{VCSProviderID: hook.vcsProviderID},
		EventPayload: *payload,
	})
}
