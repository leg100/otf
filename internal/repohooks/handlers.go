package repohooks

import (
	"context"
	"errors"
	"net/http"
	"path"

	"github.com/leg100/otf/internal/logr"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal/http/decode"
	"github.com/leg100/otf/internal/vcs"
)

// handlerPrefix is the URL path prefix for endpoints receiving vcs events
const handlerPrefix = "/webhooks/vcs"

type (
	// handlers handle VCS events triggered by webhooks
	handlers struct {
		logr.Logger
		vcs.Publisher
		*vcs.Service

		handlerDB
		vcsKindDB
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
	// vcsKindDB is a database of vcs kinds
	vcsKindDB interface {
		GetKind(id vcs.KindID) (vcs.Kind, error)
	}
)

func newHandler(
	logger logr.Logger,
	publisher vcs.Publisher,
	vcsKindDB vcsKindDB,
	handlerDB handlerDB,
) *handlers {
	return &handlers{
		Logger:    logger,
		Publisher: publisher,
		vcsKindDB: vcsKindDB,
		handlerDB: handlerDB,
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
	h.V(2).Info("received vcs event", "repohook_id", opts.ID, "repo", hook.repoPath, "cloud", hook.vcsKindID)

	// look up vcs kind for hook
	kind, err := h.GetKind(hook.vcsKindID)
	if err != nil {
		h.Error(nil, "no event unmarshaler found for event", "repohook_id", opts.ID, "repo", hook.repoPath, "cloud", hook.vcsKindID)
		http.Error(w, "no event unmarshaler found for event", http.StatusNotFound)
		return
	}
	// handle event using vcs kind's event handler
	payload, err := kind.EventHandler(r, hook.secret)
	// either ignore the event, return an error, or publish the event onwards
	var ignore vcs.ErrIgnoreEvent
	if errors.As(err, &ignore) {
		h.V(2).Info("ignoring event: "+err.Error(), "repohook_id", opts.ID, "repo", hook.repoPath, "cloud", hook.vcsKindID)
		return
	} else if err != nil {
		h.Error(err, "handling vcs event", "repohook_id", opts.ID, "repo", hook.repoPath, "cloud", hook.vcsKindID)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	h.Publish(vcs.Event{
		EventHeader: vcs.EventHeader{
			VCSProviderID: hook.vcsProviderID,
			Source:        kind.GetSource(),
		},
		EventPayload: *payload,
	})
}
